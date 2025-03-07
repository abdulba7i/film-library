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
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Gender      string `json:"bio"`
	DateOfBirth string `json:"date_of_birth"`
}

type Film struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Releasedate string  `json:"release_date"`
	Rating      float32 `json:"rating"`
	Listactors  []Actor `json:"list_actors"`
}

func Connect(c config.Database) (*Storage, error) {
	const op = "storage.postgre.New"
	sqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", c.Host, c.Port, c.User, c.Password, c.Dbname)

	db, err := sql.Open("postgres", sqlInfo)
	if err != nil {
		return nil, fmt.Errorf("%s, %w", op, err)
	}

	defer db.Close()

	log.Printf("Database connected was created: %s", sqlInfo)

	if err = goose.Up(db, "./migrations"); err != nil {
		return nil, fmt.Errorf("%s, %w", op, err)
	}

	if err = goose.Down(db, "./migrations"); err != nil {
		return nil, fmt.Errorf("%s, %w", op, err)
	}
	return &Storage{db: db}, nil
}

func (s *Storage) AddedInfoActor(tx *sql.Tx, actor Actor) error {
	const op = "storage.postgres.AddedInfoActor"

	query := `INSERT INTO actor (name, bio, date_of_birth) VALUES ($1, $2, $3)`

	_, err := tx.Exec(query, actor.Name, actor.Gender, actor.DateOfBirth)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) ChangeInfoActor(tx *sql.Tx, actor Actor) error {
	const op = "storage.postgres.ChangeInfoActor"

	query := `UPDATE actor SET name = $1, bio = $2, date_of_birth = $3 WHERE id = $4`

	_, err := tx.Exec(query, actor.Name, actor.Gender, actor.DateOfBirth, actor.Id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
