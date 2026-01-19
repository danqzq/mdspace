package models

import "time"

// Markdown represents a stored markdown document
type Markdown struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Views     int64     `json:"views"`
	OwnerID   string    `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Comment represents a line comment on a markdown document
type Comment struct {
	ID        string    `json:"id"`
	Line      int       `json:"line"`
	Text      string    `json:"text"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateMarkdownRequest is the request body for creating a new markdown
type CreateMarkdownRequest struct {
	Content string `json:"content"`
}

// CreateCommentRequest is the request body for adding a comment
type CreateCommentRequest struct {
	Line   int    `json:"line"`
	Text   string `json:"text"`
	Author string `json:"author"`
}
