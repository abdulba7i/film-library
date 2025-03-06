package main

import "database/sql"

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
	Id           int    `json:"id"`
	Name         string `json:"name"`
	Release_date string `json:"release_date"`
	Rating       string `json:"rating"`
}
