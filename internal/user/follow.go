package user

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type FollowHandler struct {
	DB *sql.DB
}

// parseUsernameFromToken expects "dummy-token-for-user-username"
func parseUsernameFromToken(token string) (string, error) {
	prefix := "dummy-token-for-user-"
	if !strings.HasPrefix(token, prefix) {
		return "", errors.New("invalid token format")
	}
	return strings.TrimPrefix(token, prefix), nil
}

func (h *FollowHandler) Follow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	followeeIDStr := vars["userID"]
	followeeID, err := strconv.Atoi(followeeIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	token := r.Header.Get("Authorization")
	username, err := parseUsernameFromToken(token)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var followerID int
	err = h.DB.QueryRow("SELECT id FROM users WHERE username=$1", username).Scan(&followerID)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if followeeID == followerID {
		http.Error(w, "Can't follow your self", http.StatusBadRequest)
		return
	}
	var exists bool
	err = h.DB.QueryRow("SELECT EXISTS (SELECT 1 FROM users WHERE id=$1)", followeeID).Scan(&exists)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Followee not found", http.StatusBadRequest)
		return
	}

	err = h.DB.QueryRow("SELECT EXISTS (SELECT 1 FROM follows WHERE follower_id=$1 AND followee_id=$2)", followerID, followeeID).Scan(&exists)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Already followed", http.StatusConflict)
		return
	}

	result, err := h.DB.Exec("INSERT INTO follows (follower_id, followee_id) VALUES ($1, $2)", followerID, followeeID)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Failed to follow", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Followed successfully"))
}

func (h *FollowHandler) Unfollow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	followeeIDStr := vars["userID"]
	followeeID, err := strconv.Atoi(followeeIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	token := r.Header.Get("Authorization")
	username, err := parseUsernameFromToken(token)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var followerID int
	err = h.DB.QueryRow("SELECT id FROM users WHERE username=$1", username).Scan(&followerID)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if followerID == followeeID {
		http.Error(w, "Cannot unfollow yourself", http.StatusBadRequest)
		return
	}

	var exists bool
	err = h.DB.QueryRow("SELECT EXISTS (SELECT 1 FROM users WHERE id=$1)", followeeID).Scan(&exists)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Followee not found", http.StatusBadRequest)
		return
	}

	err = h.DB.QueryRow("SELECT EXISTS (SELECT 1 FROM follows WHERE follower_id=$1 AND followee_id=$2)", followerID, followeeID).Scan(&exists)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Not following", http.StatusBadRequest)
		return
	}
	result, err := h.DB.Exec(
		"DELETE FROM follows WHERE follower_id=$1 AND followee_id=$2",
		followerID, followeeID,
	)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Failed to unfollow", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Unfollowed sucessfully"))
}
