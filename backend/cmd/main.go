package main

import (
	"log"
	"main/cmd/api"
	"main/internal/database"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("/.env not found.")
	}
	gormDB, err := database.NewPostgreSQLStorage()
	if err != nil {
		log.Fatal("Error connection to PostgreSQL DB", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatal("Cannot get sql.DB from gorm", err)
	}

	server := api.NewAPIServer(":8034", sqlDB)
	err = server.Run()

	if err != nil {
		log.Fatal("Error running server.", err)
	}
}
