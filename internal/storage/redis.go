package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/danqzq/mdspace/internal/models"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	// MarkdownTTL is the time-to-live for markdown documents (24 hours)
	MarkdownTTL = 24 * time.Hour
	// MaxFilesPerUser is the maximum number of files a user can store
	MaxFilesPerUser = 10
)

// Store handles all Redis storage operations
type Store struct {
	client *redis.Client
}

// NewStore creates a new Redis store
func NewStore(redisURL string) (*Store, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis URL: %w", err)
	}

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Store{client: client}, nil
}

// Close closes the Redis connection
func (s *Store) Close() error {
	return s.client.Close()
}

// markdownKey returns the Redis key for a markdown document
func markdownKey(id string) string {
	return fmt.Sprintf("md:%s", id)
}

// commentsKey returns the Redis key for comments on a markdown document
func commentsKey(id string) string {
	return fmt.Sprintf("md:%s:comments", id)
}

// userFilesKey returns the Redis key for a user's file set
func userFilesKey(userID string) string {
	return fmt.Sprintf("user:%s:files", userID)
}

// SaveMarkdown stores a new markdown document with 24h TTL
func (s *Store) SaveMarkdown(ctx context.Context, content, ownerID string) (*models.Markdown, error) {
	count, err := s.GetUserFileCount(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if count >= MaxFilesPerUser {
		return nil, fmt.Errorf("rate limit exceeded: maximum %d files per user", MaxFilesPerUser)
	}

	id := uuid.New().String()[:8]
	now := time.Now()
	expiresAt := now.Add(MarkdownTTL)

	md := &models.Markdown{
		ID:        id,
		Content:   content,
		Views:     0,
		OwnerID:   ownerID,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}

	key := markdownKey(id)
	pipe := s.client.Pipeline()

	pipe.HSet(ctx, key, map[string]any{
		"id":         md.ID,
		"content":    md.Content,
		"views":      md.Views,
		"owner_id":   md.OwnerID,
		"created_at": md.CreatedAt.Unix(),
		"expires_at": md.ExpiresAt.Unix(),
	})
	pipe.Expire(ctx, key, MarkdownTTL)

	userKey := userFilesKey(ownerID)
	pipe.SAdd(ctx, userKey, id)
	pipe.Expire(ctx, userKey, MarkdownTTL)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to save markdown: %w", err)
	}

	return md, nil
}

// GetMarkdown retrieves a markdown document by ID
func (s *Store) GetMarkdown(ctx context.Context, id string) (*models.Markdown, error) {
	key := markdownKey(id)
	result, err := s.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get markdown: %w", err)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("markdown not found")
	}

	var views int64
	fmt.Sscanf(result["views"], "%d", &views)

	var createdAt, expiresAt int64
	fmt.Sscanf(result["created_at"], "%d", &createdAt)
	fmt.Sscanf(result["expires_at"], "%d", &expiresAt)

	return &models.Markdown{
		ID:        result["id"],
		Content:   result["content"],
		Views:     views,
		OwnerID:   result["owner_id"],
		CreatedAt: time.Unix(createdAt, 0),
		ExpiresAt: time.Unix(expiresAt, 0),
	}, nil
}

// IncrementViews atomically increments the view count
func (s *Store) IncrementViews(ctx context.Context, id string) (int64, error) {
	key := markdownKey(id)
	return s.client.HIncrBy(ctx, key, "views", 1).Result()
}

// DeleteMarkdown removes a markdown document
func (s *Store) DeleteMarkdown(ctx context.Context, id, ownerID string) error {
	md, err := s.GetMarkdown(ctx, id)
	if err != nil {
		return err
	}
	if md.OwnerID != ownerID {
		return fmt.Errorf("permission denied")
	}

	pipe := s.client.Pipeline()
	pipe.Del(ctx, markdownKey(id))
	pipe.Del(ctx, commentsKey(id))
	pipe.SRem(ctx, userFilesKey(ownerID), id)

	_, err = pipe.Exec(ctx)
	return err
}

// GetUserFileCount returns the number of files owned by a user
func (s *Store) GetUserFileCount(ctx context.Context, userID string) (int64, error) {
	return s.client.SCard(ctx, userFilesKey(userID)).Result()
}

// AddComment adds a comment to a markdown document
func (s *Store) AddComment(ctx context.Context, markdownID string, comment *models.Comment) error {
	comment.ID = uuid.New().String()[:8]
	comment.CreatedAt = time.Now()

	data, err := json.Marshal(comment)
	if err != nil {
		return fmt.Errorf("failed to marshal comment: %w", err)
	}

	key := commentsKey(markdownID)
	pipe := s.client.Pipeline()
	pipe.RPush(ctx, key, data)

	ttl, err := s.client.TTL(ctx, markdownKey(markdownID)).Result()
	if err == nil && ttl > 0 {
		pipe.Expire(ctx, key, ttl)
	}

	_, err = pipe.Exec(ctx)
	return err
}

// GetComments retrieves all comments for a markdown document
func (s *Store) GetComments(ctx context.Context, markdownID string) ([]models.Comment, error) {
	key := commentsKey(markdownID)
	results, err := s.client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}

	comments := make([]models.Comment, 0, len(results))
	for _, data := range results {
		var comment models.Comment
		if err := json.Unmarshal([]byte(data), &comment); err != nil {
			continue
		}
		comments = append(comments, comment)
	}

	return comments, nil
}
