package main

import (
	"encoding/json"
	"log"
	"main/internal/database"
	"net/http"
)

func main() {
	log.Printf("Start")

	db, err := database.NewConnection()
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
	})
	http.ListenAndServe(":8080", nil)
}
