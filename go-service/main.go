package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	db := InitDB("postgres://user:pass@localhost:5432/db?sslmode=disable")
	defer db.Close()

	cache := InitRedis("localhost:6379")

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS events (
			id SERIAL PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			event_type VARCHAR(255) NOT NULL,
			timestamp TIMESTAMP NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_user_id ON events(user_id);
		CREATE INDEX IF NOT EXISTS idx_timestamp ON events(timestamp);
	`)
	if err != nil {
		log.Fatal("Failed to create table/indexes:", err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) {
		HandleEvent(w, r, db, cache)
	}).Methods("POST")
	r.HandleFunc("/events/batch", func(w http.ResponseWriter, r *http.Request) {
		HandleBatchEvents(w, r, db, cache)
	}).Methods("POST")
	r.HandleFunc("/analytics/users", func(w http.ResponseWriter, r *http.Request) {
		HandleAnalytics(w, r, db, cache)
	}).Methods("GET")
	r.HandleFunc("/events/filter", func(w http.ResponseWriter, r *http.Request) {
		HandleFilterEvents(w, r, db, cache)
	}).Methods("GET")

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8000"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	}).Handler(r)

	fmt.Println("Go microservice running on :8080")
	log.Fatal(http.ListenAndServe(":8080", corsHandler))
}