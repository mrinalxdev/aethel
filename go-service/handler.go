package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type Event struct {
	UserID    string    `json:"user_id"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
}

func HandleEvent(w http.ResponseWriter, r *http.Request, db *sql.DB, cache *redis.Client) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Insert into Postgres
	_, err := db.Exec("INSERT INTO events (user_id, event_type, timestamp) VALUES ($1, $2, $3)",
		event.UserID, event.EventType, event.Timestamp)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	key := "user_events:" + event.UserID
	pipe := cache.Pipeline()
	pipe.Incr(context.Background(), key)
	pipe.Expire(context.Background(), key, 5*time.Minute)
	_, err = pipe.Exec(context.Background())
	if err != nil {
		http.Error(w, "Cache error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "event ingested"})
}

func HandleBatchEvents(w http.ResponseWriter, r *http.Request, db *sql.DB, cache *redis.Client) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var events []Event
	if err := json.NewDecoder(r.Body).Decode(&events); err != nil {
		http.Error(w, "Invalid JSON", http.StatusMethodNotAllowed)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	pipe := cache.Pipeline()
	for _, event := range events {
		_, err := tx.Exec("INSERT INTO events (user_id, event_type, timestamp) VALUES ($1, $2, $3)",
			event.UserID, event.EventType, event.Timestamp)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		key := "user_events:" + event.UserID
		pipe.Incr(context.Background(), key)
		pipe.Expire(context.Background(), key, 5*time.Minute)
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if _, err := pipe.Exec(context.Background()); err != nil {
		http.Error(w, "Cache error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "batch events ingested"})
}