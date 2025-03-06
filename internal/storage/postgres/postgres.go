package main

import (
	"database/sql"
	"film-library/internal/config"
	"fmt"
	"log"

	"github.com/pressly/goose"
)

type Storage struct {
	db *sql.DB
}

type Actor struct {
	Id            int    `json:"id"`
	Name          string `json:"name"`
	Bio           string `json:"bio"`
	Date_of_birth string `json:"date_of_birth"`
}

type Film struct {
	Id           int     `json:"id"`
	Name         string  `json:"name"`
	Release_date string  `json:"release_date"`
	Rating       float32 `json:"rating"`
}

func Connect(c config.Database) (*Storage, error) {
	const op = "storage.postgre.New"
	sqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", c.Host, c.Port, c.User, c.Password, c.Dbname)

	db, err := sql.Open("postgres", sqlInfo)
	if err != nil {
		return nil, fmt.Errorf("%s, %w", op, err)
	}

	log.Printf("Database connected was created: %s", sqlInfo)

	err = goose.Up(db, "file://./migrations")

	return &Storage{}, nil
}
