// Package database implements DB connection
package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

type Storage struct {
	db *pgx.Conn
}

func NewPostgresDatabase(db *pgx.Conn) *Storage {
	return &Storage{db: db}
}

func ConnectToPostgres() *pgx.Conn {

	db, err := pgx.Connect(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal("Unable to connect to postgres: ", err)
	}

	fmt.Println("Successfully connected to postgres")
	return db
}
