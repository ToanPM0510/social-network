package main

import (
	"fmt"
	"log"

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
}
