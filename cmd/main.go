package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ToanPM0510/social-network/api"
	"github.com/ToanPM0510/social-network/internal/db"
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

	authHandler := &api.AuthHandler{DB: database}

	http.HandleFunc("/register", authHandler.Register)
	http.HandleFunc("/login", authHandler.Login)

	fmt.Println("Server running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
