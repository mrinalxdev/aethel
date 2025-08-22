#!/bin/bash

# Script to populate 500 dummy events into the event analytics pipeline
# Sends 400 single events to /event and 5 batches of 20 events to /events/batch

BASE_URL="http://localhost:8080"

check_server() {
    if ! curl -s -f -o /dev/null "$BASE_URL/event"; then
        echo "Error: Server at $BASE_URL is not running. Please start the Go service."
        exit 1
    fi
}

send_single_event() {
    local user_id=$1
    local event_type=$2
    local timestamp=$3
    echo "Sending single event: user_id=$user_id, event_type=$event_type, timestamp=$timestamp"
    response=$(curl -s -w "%{http_code}" -X POST "$BASE_URL/event" \
        -H "Content-Type: application/json" \
        -d "{\"user_id\": \"$user_id\", \"event_type\": \"$event_type\", \"timestamp\": \"$timestamp\"}")
    if [ "$response" -eq 201 ]; then
        echo "Success: Event ingested"
    else
        echo "Error: Failed to ingest event (HTTP status: $response)"
        exit 1
    fi
}

send_batch_events() {
    local batch_json=$1
    echo "Sending batch of events"
    response=$(curl -s -w "%{http_code}" -X POST "$BASE_URL/events/batch" \
        -H "Content-Type: application/json" \
        -d "$batch_json")
    if [ "$response" -eq 201 ]; then
        echo "Success: Batch events ingested"
    else
        echo "Error: Failed to ingest batch events (HTTP status: $response)"
        exit 1
    fi
}
echo "Starting event ingestion..."
check_server
users=("user1" "user2" "user3" "user4" "user5" "user6" "user7" "user8" "user9" "user10")
event_types=("click" "view" "purchase")
echo "Sending 400 single events..."
for i in {1..400}; do
    user_id=${users[$((RANDOM % 10))]}  # Random user from 10
    event_type=${event_types[$((RANDOM % 3))]}  # Random event type
    # Generate timestamp between Aug 20 and Aug 22, 2025
    timestamp=$(date -u -d "2025-08-20T00:00:00Z + $((RANDOM % 172800)) seconds" +"%Y-%m-%dT%H:%M:%SZ")
    send_single_event "$user_id" "$event_type" "$timestamp"
    sleep 0.01  # Small delay to avoid overwhelming the server
done
echo "Sending 5 batches of 20 events..."
for batch in {1..5}; do
    batch_json="["
    for i in {1..20}; do
        user_id=${users[$((RANDOM % 10))]}
        event_type=${event_types[$((RANDOM % 3))]}
        timestamp=$(date -u -d "2025-08-20T00:00:00Z + $((RANDOM % 172800)) seconds" +"%Y-%m-%dT%H:%M:%SZ")
        batch_json+="{\"user_id\": \"$user_id\", \"event_type\": \"$event_type\", \"timestamp\": \"$timestamp\"}"
        if [ $i -lt 20 ]; then
            batch_json+=","
        fi
    done
    batch_json+="]"
    send_batch_events "$batch_json"
    sleep 0.1  # Delay between batches
done

echo "Event ingestion completed. Total: 500 events."