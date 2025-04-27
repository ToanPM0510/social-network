package feed

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

type NewsfeedHandler struct {
	DB *sql.DB
}

type NewsfeedItem struct {
	PostID         int       `json:"post_id"`
	AuthorUsername string    `json:"author_username"`
	Content        string    `json:"content"`
	ImageURL       string    `json:"image_url"`
	CreatedAt      time.Time `json:"created_at"`
}

// Reuse function parseUsernameFromToken
func parseUsernameFromToken(token string) (string, error) {
	prefix := "dummy-token-for-user-"
	if !strings.HasPrefix(token, prefix) {
		return "", errors.New("invalid token format")
	}
	return strings.TrimPrefix(token, prefix), nil
}

func (h *NewsfeedHandler) GetNewsfeed(w http.ResponseWriter, r *http.Request) {
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

	rows, err := h.DB.Query(`
        SELECT posts.id, users.username, posts.content, posts.image_url, posts.created_at
        FROM posts
        JOIN users ON posts.user_id = users.id
        WHERE posts.user_id = $1
           OR posts.user_id IN (SELECT followee_id FROM follows WHERE follower_id = $1)
        ORDER BY posts.created_at DESC
    `, userID)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var feed []NewsfeedItem
	for rows.Next() {
		var item NewsfeedItem
		err := rows.Scan(&item.PostID, &item.AuthorUsername, &item.Content, &item.ImageURL, &item.CreatedAt)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		feed = append(feed, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feed)
}
