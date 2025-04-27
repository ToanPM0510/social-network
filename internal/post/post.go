package post

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type PostHandler struct {
	DB *sql.DB
}

type CreatePostRequest struct {
	Content  string `json:"content"`
	ImageURL string `json:"image_url"`
}

// parseUsernameFromToken (reuse code follow phase)
func parseUsernameFromToken(token string) (string, error) {
	prefix := "dummy-token-for-user-"
	if !strings.HasPrefix(token, prefix) {
		return "", errors.New("invalid token format")
	}
	return strings.TrimPrefix(token, prefix), nil
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Content) == "" {
		http.Error(w, "Content cannot be empty", http.StatusBadRequest)
		return
	}

	token := r.Header.Get("Authorization")
	username, err := parseUsernameFromToken(token)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var userID int
	err = h.DB.QueryRow("SELECT id FROM users WHERE username=$1", username).Scan(&userID)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	_, err = h.DB.Exec("INSERT INTO posts (user_id, content, image_url) VALUES ($1, $2, $3)",
		userID, req.Content, req.ImageURL,
	)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Post created successfully"))
}
