package profile

import (
	"database/sql"
	"encoding/json"
	"net/http"

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
	PostID    int       `json:"post_id"`
	Content   string    `json:"content"`
	ImageURL  string    `json:"image_url"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	var userID int
	var name, uname string
	err := h.DB.QueryRow("SELECT id, name, username FROM users WHERE username = $1", username).
		Scan(&userID, &name, &uname)
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
        SELECT id, content, image_url, created_at
        FROM posts
        WHERE user_id = $1
        ORDER BY created_at DESC
    `, userID)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []ProfilePost
	for rows.Next() {
		var p ProfilePost
		err := rows.Scan(&p.PostID, &p.Content, &p.ImageURL, &p.CreatedAt)
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
