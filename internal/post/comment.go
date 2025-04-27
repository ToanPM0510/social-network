package post

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type CommentHandler struct {
	DB *sql.DB
}

type CommentRequest struct {
	Content string `json:"content"`
}

func (h *CommentHandler) CommentOnPost(w http.ResponseWriter, r *http.Request) {
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

	var req CommentRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Content == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var userID int
	err = h.DB.QueryRow("SELECT id FROM users WHERE username=$1", username).Scan(&userID)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	_, err = h.DB.Exec(
		"INSERT INTO comments (post_id, user_id, content) VALUES ($1, $2, $3)",
		postID, userID, req.Content,
	)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
