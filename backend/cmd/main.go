package main

import (
	"context"
	"log"
	"main/cmd/api"
	"main/internal/database"
	"main/services/absence"
	"main/services/scheduler"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Load enviroment - local enviroment
	// path: ../.env -> because GO is running from go run ./cmd
	// .env is store in project, not the conntainer
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No ../.env file found, using environment variables")
	}

	// gromDB - DB instance, error
	// Grom is used only for connection to DB,
	// communication with DB is implemented with SQL
	gormDB, err := database.NewPostgreSQLStorage()
	if err != nil {
		log.Fatal("Error connection to PostgreSQL DB", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatal("Cannot get sql.DB from gorm", err)
	}

	// Connection to MongoDB
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

	// Check connection and Ping to see if the connection is real
	// if is not FATAL
	if err := mongoClient.Ping(context.Background(), nil); err != nil {
		log.Fatal("MongoDB ping failed:", err)
	}
	log.Println("Connected to MongoDB")

	mongoDBName := os.Getenv("MONGO_DB")
	if mongoDBName == "" {
		mongoDBName = "hrsystem_audit"
	}
	mongoDB := mongoClient.Database(mongoDBName)

	// Configuration for Port - http.ListenAndServe expect 8034
	// Nomralization 8034 -> :8034
	port := os.Getenv("PORT")
	if port == "" {
		port = "8034"
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	// Leave accrual / carry-over / expiry scheduler.

	if os.Getenv("SCHEDULER_ENABLED") != "false" {
		interval := time.Hour
		if v := os.Getenv("SCHEDULER_INTERVAL"); v != "" {
			if d, err := time.ParseDuration(v); err == nil {
				interval = d
			}
		}
		// Scheduler gets DB for calculations
		go scheduler.New(absence.NewStore(sqlDB), interval).Start(context.Background())
	}

	// HTTP Server
	server := api.NewAPIServer(port, sqlDB, mongoDB)
	if err := server.Run(); err != nil {
		log.Fatal("Error running server:", err)
	}
}
