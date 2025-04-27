package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/ToanPM0510/social-network/internal/auth"
)

type AuthHandler struct {
	DB *sql.DB
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	_, err = h.DB.Exec("INSERT INTO users (name, username, password) VALUES ($1, $2, $3)",
		req.Name, req.Username, hashedPassword,
	)
	if err != nil {
		http.Error(w, "Username already taken or server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User registered successfully"))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var storedHash string
	var userID int
	err := h.DB.QueryRow("SELECT id, password FROM users WHERE username = $1", req.Username).Scan(&userID, &storedHash)
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	if !auth.CheckPasswordHash(req.Password, storedHash) {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	//simple token
	token := "dummy-token-for-user-" + req.Username

	resp := map[string]string{"token": token}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
