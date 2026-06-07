package main

import (
	"context"
	"log"
	"main/cmd/api"
	"main/internal/database"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}
	mongoClient, err := mongo.Connect(
		context.Background(),
		options.Client().ApplyURI(mongoURI).SetConnectTimeout(10*time.Second),
	)
	if err != nil {
		log.Fatal("Error connecting to MongoDB:", err)
	}
	defer mongoClient.Disconnect(context.Background())

	if err := mongoClient.Ping(context.Background(), nil); err != nil {
		log.Fatal("MongoDB ping failed:", err)
	}
	log.Println("Connected to MongoDB")

	mongoDB := mongoClient.Database(os.Getenv("MONGO_DB"))
	if mongoDB == nil {
		mongoDB = mongoClient.Database("hrsystem_audit")
	}

	server := api.NewAPIServer(":8034", sqlDB, mongoDB)
	if err := server.Run(); err != nil {
		log.Fatal("Error running server:", err)
	}
}
