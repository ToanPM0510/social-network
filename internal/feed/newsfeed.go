package feed

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type NewsfeedHandler struct {
	DB *sql.DB
}

type PostResponse struct {
	PostID         int       `json:"post_id"`
	AuthorUsername string    `json:"author_username"`
	Content        string    `json:"content"`
	ImageURL       string    `json:"image_url"`
	CreatedAt      time.Time `json:"created_at"`
	LikeCount      int       `json:"like_count"`
	CommentCount   int       `json:"comment_count"`
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
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // default
	offset := 0 // default

	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil && l > 0 {
			limit = l
		}
	}
	if offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err == nil && o >= 0 {
			offset = o
		}
	}

	username, err := parseUsernameFromToken(token)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var userID int
	err = h.DB.QueryRow("SELECT id FROM users WHERE username=$1", username).Scan(&userID)
	if err == sql.ErrNoRows {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	rows, err := h.DB.Query(`
		SELECT 
			p.id, u.username, p.content, p.image_url, p.created_at,
			COALESCE(l.like_count, 0),
			COALESCE(c.comment_count, 0)
		FROM posts p
		JOIN users u ON p.user_id = u.id
		JOIN follows f ON f.followee_id = p.user_id
		LEFT JOIN (
			SELECT post_id, COUNT(*) AS like_count
			FROM likes
			GROUP BY post_id
		) l ON p.id = l.post_id
		LEFT JOIN (
			SELECT post_id, COUNT(*) AS comment_count
			FROM comments
			GROUP BY post_id
		) c ON p.id = c.post_id
		WHERE f.follower_id = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)

	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []PostResponse
	for rows.Next() {
		var post PostResponse
		err := rows.Scan(
			&post.PostID,
			&post.AuthorUsername,
			&post.Content,
			&post.ImageURL,
			&post.CreatedAt,
			&post.LikeCount,
			&post.CommentCount,
		)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(posts); err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
}
