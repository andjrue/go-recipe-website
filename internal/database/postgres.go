// Package database implements DB connection
package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func NewPostgresDatabase(db *pgxpool.Pool) *Storage {
	return &Storage{db: db}
}

func ConnectToPostgres() *pgxpool.Pool {

	db, err := pgxpool.New(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal("Unable to connect to postgres: ", err)
	}

	fmt.Println("Successfully connected to postgres")
	return db
}
