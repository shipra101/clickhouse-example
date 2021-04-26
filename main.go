package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/ClickHouse/clickhouse-go"
)

func main() {
	connect, err := sql.Open("clickhouse", "tcp://127.0.0.1:9001?debug=true")
	if err != nil {
		log.Fatal(err)
	}
	if err := connect.Ping(); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("[%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		} else {
			fmt.Println(err)
		}
		return
	}

	_, err = connect.Exec(`
	CREATE DATABASE IF NOT EXISTS db
	`)

	_, err = connect.Exec(`
	CREATE TABLE IF NOT EXISTS db.entries(
		timestamp DateTime,
		parameter String,
		value Float64)
		ENGINE = MergeTree()
		PARTITION BY parameter
		ORDER BY (timestamp, parameter)
	`)

	if err != nil {
		log.Fatal(err)
	}
	var (
		tx, _   = connect.Begin()
		stmt, _ = tx.Prepare("INSERT INTO db.entries (timestamp, parameter, value) VALUES (?, ?, ?)")
	)
	defer stmt.Close()

	for i := 0; i < 100; i++ {
		if _, err := stmt.Exec(
			time.Now(),
			"11"+fmt.Sprintf("%d", i),
			10.2,
		); err != nil {
			log.Fatal(err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	rows, err := connect.Query("SELECT timestamp, parameter, value FROM db.entries")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			parameter string
			value     float64
			timestamp time.Time
		)
		if err := rows.Scan(&timestamp, &parameter, &value); err != nil {
			log.Fatal(err)
		}
		log.Printf("parameter: %s, value: %f,  timestamp: %s", parameter, value, timestamp)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
}
