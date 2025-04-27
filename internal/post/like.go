package post

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type LikeHandler struct {
	DB *sql.DB
}

func (h *LikeHandler) LikePost(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	username, err := parseUsernameFromToken(token)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	postIDStr := vars["postID"]
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post id", http.StatusBadRequest)
		return
	}

	var userID int
	err = h.DB.QueryRow("SELECT id FROM users WHERE username=$1", username).Scan(&userID)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	_, err = h.DB.Exec(
		"INSERT INTO likes (post_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		postID, userID,
	)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
