package main

import (
	"log"
	"main/cmd/api"
	"main/internal/database"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("/.env not found.")
	}
	db, err := database.NewPostgreSQLStorage()
	if err != nil {
		log.Fatal("Error connection to PostgreSQL DB", err)
	}
	var now time.Time

	err = db.Raw("SELECT NOW()").Scan(&now).Error
	if err != nil {
		panic("DB is not connected.")
	}
	log.Println("DB time:", now)

	server := api.NewAPIServer(":8034", nil)
	err = server.Run()

	if err != nil {
		log.Fatal("Error running server.", err)
	}
}
