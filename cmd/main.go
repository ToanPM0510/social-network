package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ToanPM0510/social-network/api"
	"github.com/ToanPM0510/social-network/internal/db"
	"github.com/ToanPM0510/social-network/internal/user"
	"github.com/gorilla/mux"
)

func main() {
	config := db.DBConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "admin",
		Password: "secret",
		DBName:   "social_network",
		SSLMode:  "disable", // docker local ==> disable ssl
	}

	database, err := db.ConnectDB(config)
	if err != nil {
		log.Fatal("Cannot connect to database: ", err)
	}
	defer database.Close()

	fmt.Println("Connected to database successfully!")

	r := mux.NewRouter()

	authHandler := &api.AuthHandler{DB: database}
	followHandler := &user.FollowHandler{DB: database}

	r.HandleFunc("/register", authHandler.Register).Methods("POST")
	r.HandleFunc("/login", authHandler.Login).Methods("POST")
	r.HandleFunc("/follow/{userID}", followHandler.Follow).Methods("POST")
	r.HandleFunc("/unfollow/{userID}", followHandler.Unfollow).Methods("POST")

	fmt.Println("Server running at http://localhost:8080")
	http.ListenAndServe(":8080", r)
}
