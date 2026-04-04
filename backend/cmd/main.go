package main

import (
	"context"
	"log"
	"main/cmd/api"
	"main/internal/database"
)

func main() {
	log.Printf("Start")
	db, err := database.NewPostgreSQLStorage()
	if err != nil {
		log.Fatal("Error connection to PostgreSQL DB", err)
	}
	db.Ping()

	ctx := context.Background()

	if err := db.PingContext(ctx); err != nil {
		panic(err)
	}
	server := api.NewAPIServer(":8034", nil)
	err = server.Run()

	if err != nil {
		log.Fatal("Error running server.", err)
	}
}
