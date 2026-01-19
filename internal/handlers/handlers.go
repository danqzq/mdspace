package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/danqzq/mdspace/internal/models"
	"github.com/danqzq/mdspace/internal/storage"
	"github.com/go-chi/chi/v5"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	store   *storage.Store
	baseURL string
}

// NewHandler creates a new Handler instance
func NewHandler(store *storage.Store, baseURL string) *Handler {
	return &Handler{
		store:   store,
		baseURL: strings.TrimSuffix(baseURL, "/"),
	}
}

// getSessionID extracts the session ID from the request
func getSessionID(r *http.Request) string {
	return r.Header.Get("X-Session-ID")
}

// respondJSON writes a JSON response
func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError writes a JSON error response
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// CreateMarkdown handles POST /api/markdown
func (h *Handler) CreateMarkdown(w http.ResponseWriter, r *http.Request) {
	var req models.CreateMarkdownRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Content) == "" {
		respondError(w, http.StatusBadRequest, "Content cannot be empty")
		return
	}

	if len(req.Content) > 1024*1024 {
		respondError(w, http.StatusBadRequest, "Content too large (max 1MB)")
		return
	}

	sessionID := getSessionID(r)
	md, err := h.store.SaveMarkdown(r.Context(), req.Content, sessionID)
	if err != nil {
		if strings.Contains(err.Error(), "rate limit") {
			respondError(w, http.StatusTooManyRequests, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to save markdown")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{
		"id":         md.ID,
		"share_url":  h.baseURL + "/view/" + md.ID,
		"expires_at": md.ExpiresAt,
	})
}

// GetMarkdown handles GET /api/markdown/{id}
func (h *Handler) GetMarkdown(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "Missing markdown ID")
		return
	}

	md, err := h.store.GetMarkdown(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "Markdown not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get markdown")
		return
	}

	views, _ := h.store.IncrementViews(r.Context(), id)
	md.Views = views

	sessionID := getSessionID(r)
	isOwner := md.OwnerID == sessionID

	respondJSON(w, http.StatusOK, map[string]any{
		"id":         md.ID,
		"content":    md.Content,
		"views":      md.Views,
		"is_owner":   isOwner,
		"created_at": md.CreatedAt,
		"expires_at": md.ExpiresAt,
	})
}

// DeleteMarkdown handles DELETE /api/markdown/{id}
func (h *Handler) DeleteMarkdown(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "Missing markdown ID")
		return
	}

	sessionID := getSessionID(r)
	err := h.store.DeleteMarkdown(r.Context(), id, sessionID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "Markdown not found")
			return
		}
		if strings.Contains(err.Error(), "permission denied") {
			respondError(w, http.StatusForbidden, "Permission denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete markdown")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateComment handles POST /api/markdown/{id}/comments
func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "Missing markdown ID")
		return
	}

	_, err := h.store.GetMarkdown(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "Markdown not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to verify markdown")
		return
	}

	var req models.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Line < 1 {
		respondError(w, http.StatusBadRequest, "Line number must be positive")
		return
	}

	if strings.TrimSpace(req.Text) == "" {
		respondError(w, http.StatusBadRequest, "Comment text cannot be empty")
		return
	}

	if len(req.Text) > 1000 {
		respondError(w, http.StatusBadRequest, "Comment too long (max 1000 characters)")
		return
	}

	author := req.Author
	if author == "" {
		author = "Anonymous"
	}

	comment := &models.Comment{
		Line:   req.Line,
		Text:   req.Text,
		Author: author,
	}

	if err := h.store.AddComment(r.Context(), id, comment); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to add comment")
		return
	}

	respondJSON(w, http.StatusCreated, comment)
}

// GetComments handles GET /api/markdown/{id}/comments
func (h *Handler) GetComments(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "Missing markdown ID")
		return
	}

	comments, err := h.store.GetComments(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get comments")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"comments": comments,
	})
}

// GetUserStats handles GET /api/user/stats
func (h *Handler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	sessionID := getSessionID(r)
	count, err := h.store.GetUserFileCount(r.Context(), sessionID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get stats")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"files_count": count,
		"files_limit": storage.MaxFilesPerUser,
	})
}
