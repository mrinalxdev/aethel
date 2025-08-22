package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type UserAnalytics struct {
	UserID     string `json:"user_id"`
	EventCount int64  `json:"event_count"`
}

type FilterResult struct {
	Events []Event `json:"events"`
}

func HandleAnalytics(w http.ResponseWriter, r *http.Request, db *sql.DB, cache *redis.Client) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check Redis cache
	cacheKey := "analytics:users"
	cached, err := cache.Get(context.Background(), cacheKey).Result()
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cached))
		return
	}

	// Query Postgres
	rows, err := db.Query("SELECT user_id, COUNT(*) as event_count FROM events GROUP BY user_id")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var analytics []UserAnalytics
	for rows.Next() {
		var a UserAnalytics
		if err := rows.Scan(&a.UserID, &a.EventCount); err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		analytics = append(analytics, a)
	}


	data, err := json.Marshal(analytics)
	if err != nil {
		http.Error(w, "Serialization error", http.StatusInternalServerError)
		return
	}
	cache.Set(context.Background(), cacheKey, data, 5*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func HandleFilterEvents(w http.ResponseWriter, r *http.Request, db *sql.DB, cache *redis.Client) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	

	userID := r.URL.Query().Get("user_id")
	eventType := r.URL.Query().Get("event_type")
	startTime := r.URL.Query().Get("start_time")
	endTime := r.URL.Query().Get("end_time")

	cacheKey := "filter:" + userID + ":" + eventType + ":" + startTime + ":" + endTime
	cached, err := cache.Get(context.Background(), cacheKey).Result()
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cached))
		return
	}

	query := "SELECT user_id, event_type, timestamp FROM events WHERE 1=1"
	args := []interface{}{}
	if userID != "" {
		query += " AND user_id = $1"
		args = append(args, userID)
	}
	if eventType != "" {
		query += " AND event_type = $" + string(len(args)+1)
		args = append(args, eventType)
	}
	if startTime != "" {
		query += " AND timestamp >= $" + string(len(args)+1)
		args = append(args, startTime)
	}
	if endTime != "" {
		query += " AND timestamp <= $" + string(len(args)+1)
		args = append(args, endTime)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.UserID, &e.EventType, &e.Timestamp); err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		events = append(events, e)
	}

	result := FilterResult{Events: events}
	data, err := json.Marshal(result)
	if err != nil {
		http.Error(w, "Serialization error", http.StatusInternalServerError)
		return
	}

	cache.Set(context.Background(), cacheKey, data, 5*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}