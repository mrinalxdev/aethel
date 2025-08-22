import requests
import redis
import pandas as pd
import pg8000


def get_postgres_conn():
    return pg8000.connect(
        database="db",
        user="user",
        password="pass",
        host="localhost",
        port=5432
    )

redis_client = redis.Redis(host='localhost', port=6379, db=0)


def validate_analytics():
    try:
        response = requests.get("http://localhost:8080/analytics/users")
        response.raise_for_status()
        api_analytics = response.json()

        conn = get_postgres_conn()
        cursor = conn.cursor()
        cursor.execute("SELECT user_id, COUNT(*) as event_count FROM events GROUP BY user_id")
        rows = cursor.fetchall()
        conn.close()

        df = pd.DataFrame(rows, columns=['user_id', 'event_count'])

        print("\nAPI Analytics:")
        print(pd.DataFrame(api_analytics).to_string(index=False))

        print("\nValidation (Postgres):")
        print(df.to_string(index=False))

        print("\nRedis Cache Validation:")
        for _, row in df.iterrows():
            user_id = row['user_id']
            cached_count = redis_client.get(f"user_events:{user_id}")
            cached_count = int(cached_count.decode()) if cached_count else None
            print(f"User {user_id}: Postgres count = {row['event_count']}, Redis cache = {cached_count or 'None'}")

    except Exception as e:
        print(f"Error validating analytics: {e}")

if __name__ == "__main__":
    print("Running analytics validation...")
    validate_analytics()