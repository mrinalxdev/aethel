package main

import (
	"database/sql"
	_ "github.com/lib/pq"
)

func InitDB(dsn string) *sql.DB {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	if err := db.Ping(); err != nil {
		panic(err)
	}
	// Set connection pooling
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	return db
}