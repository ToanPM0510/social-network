package profile

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type ProfileHandler struct {
	DB *sql.DB
}

type ProfileResponse struct {
	Name           string        `json:"name"`
	Username       string        `json:"username"`
	FollowersCount int           `json:"followers_count"`
	FollowingCount int           `json:"following_count"`
	Posts          []ProfilePost `json:"posts"`
}

type ProfilePost struct {
	PostID       int       `json:"post_id"`
	Content      string    `json:"content"`
	ImageURL     string    `json:"image_url"`
	CreatedAt    time.Time `json:"created_at"`
	LikeCount    int       `json:"like_count"`
	CommentCount int       `json:"comment_count"`
}

func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	var userID int
	var name, uname string
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
	err := h.DB.QueryRow(`
		SELECT id, name, username 
		FROM users 
		WHERE username = $1
	`, username).Scan(&userID, &name, &uname)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	var followersCount, followingCount int
	err = h.DB.QueryRow("SELECT COUNT(*) FROM follows WHERE followee_id = $1", userID).Scan(&followersCount)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	err = h.DB.QueryRow("SELECT COUNT(*) FROM follows WHERE follower_id = $1", userID).Scan(&followingCount)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	rows, err := h.DB.Query(`
		SELECT 
			p.id, p.content, p.image_url, p.created_at,
			COALESCE(l.like_count, 0),
			COALESCE(c.comment_count, 0)
		FROM posts p
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
		WHERE p.user_id = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)

	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []ProfilePost
	for rows.Next() {
		var p ProfilePost
		err := rows.Scan(
			&p.PostID,
			&p.Content,
			&p.ImageURL,
			&p.CreatedAt,
			&p.LikeCount,
			&p.CommentCount)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		posts = append(posts, p)
	}

	profile := ProfileResponse{
		Name:           name,
		Username:       uname,
		FollowersCount: followersCount,
		FollowingCount: followingCount,
		Posts:          posts,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}
