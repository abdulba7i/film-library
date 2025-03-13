package postgres

import (
	"database/sql"
	"film-library/internal/config"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
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
	Id          int     `json:"id"`
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

	log.Printf("Database connected was created: %s", sqlInfo)

	dir, _ := os.Getwd()
	log.Println("Current working directory:", dir)

	if err = goose.Up(db, "./internal/storage/postgres/migrations"); err != nil {
		return nil, fmt.Errorf("%s, %w", op, err)
	}

	return &Storage{db: db}, nil
}

// Actor:
func (s *Storage) AddedInfoActor(actor Actor) error {
	const op = "storage.postgres.AddedInfoActor"

	query := `INSERT INTO actors (name, gender, date_of_birth) VALUES ($1, $2, $3)`

	_, err := s.db.Exec(query, actor.Name, actor.Gender, actor.DateOfBirth)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// if err = goose.Down(db, "./migrations"); err != nil {
// 	return nil, fmt.Errorf("%s, %w", op, err)
// }

func (s *Storage) ChangeInfoActor(actor Actor) error {
	const op = "storage.postgres.ChangeInfoActor"

	query := `UPDATE actor SET name = $1, bio = $2, date_of_birth = $3 WHERE id = $4`

	_, err := s.db.Exec(query, actor.Name, actor.Gender, actor.DateOfBirth, actor.Id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) DeleteInfoActor(id int) error {
	const op = "storage.postgres.DeleteInfoActor"

	query := `DELETE FROM actor WHERE id = $1`

	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// Film:

func (s *Storage) AddedInfoFilm(film Film) error {
	const op = "storage.postgres.AddedInfoFilm"

	query := `INSERT INTO film (name, description, release_date, rating) VALUES ($1, $2, $3, $4) RETURNING id`
	var filmID int

	err := s.db.QueryRow(query, film.Name, film.Description, film.Releasedate, film.Rating).Scan(&filmID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	for _, actor := range film.Listactors {
		query = `INSERT INTO actor_film (actor_id, film_id) VALUES ($1, $2)`

		_, err = s.db.Exec(query, actor.Id, filmID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func (s *Storage) ChangeInfoFilm(film Film) error {
	const op = "storage.postgres.ChangeInfoFilm"

	query := `UPDATE film SET name = $1, description = $2, release_date = $3, rating = $4 WHERE id = $5`

	_, err := s.db.Exec(query, film.Name, film.Description, film.Releasedate, film.Rating, film.Id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) DeleteInfoFilm(id int) error {
	const op = "storage.postgres.DeleteInfoFilm"

	query := `DELETE FROM film WHERE id = $1`

	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetAllFilms(sortBy string) ([]Film, error) {
	const op = "storage.postgres.GetAllFilms"

	orderClause := "ORDER BY rating DESC" // По умолчанию сортировка по рейтингу
	switch sortBy {
	case "name":
		orderClause = "ORDER BY name"
	case "release_date":
		orderClause = "ORDER BY release_date"
	}

	query := fmt.Sprintf(`SELECT f.id, f.name, f.description, f.release_date, f.rating, a.id, a.name, a.gender, a.date_of_birth
	FROM film f
	JOIN actor_film af ON f.id = af.film_id
	JOIN actor a ON a.id = af.actor_id %s`, orderClause)

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close() // Обязательно закрыть rows после использования

	var films []Film
	filmMap := make(map[int]*Film)

	for rows.Next() {
		var film Film
		var actor Actor

		err = rows.Scan(&film.Id, &film.Name, &film.Description, &film.Releasedate, &film.Rating, &actor.Id, &actor.Name, &actor.Gender, &actor.DateOfBirth)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		// Проверяем, есть ли уже фильм в списке
		if _, exists := filmMap[film.Id]; !exists {
			filmMap[film.Id] = &film
			films = append(films, film)
		}

		// Добавляем актера к фильму
		filmMap[film.Id].Listactors = append(filmMap[film.Id].Listactors, actor)
	}

	return films, nil
}

func (s *Storage) SearchFilm(actor, film string) (Film, error) {
	const op = "storage.postgres.SearchFilm"

	query := `SELECT f.id, f.name, f.description, f.release_date, f.rating, a.id, a.name, a.gender, a.date_of_birth
	FROM film f
	JOIN actor_film af ON f.id = af.film_id
	JOIN actor a ON a.id = af.actor_id
	WHERE f.name ILIKE $1 AND a.name ILIKE $2 LIMIT 1`

	rows, err := s.db.Query(query, "%"+film+"%", "%"+actor+"%")
	if err != nil {
		return Film{}, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var filmFound Film
	filmFound.Listactors = []Actor{} // Инициализация списка актёров

	if rows.Next() {
		var foundFilm Film
		var foundActor Actor

		err = rows.Scan(&foundFilm.Id, &foundFilm.Name, &foundFilm.Description, &foundFilm.Releasedate, &foundFilm.Rating, &foundActor.Id, &foundActor.Name, &foundActor.Gender, &foundActor.DateOfBirth)
		if err != nil {
			return Film{}, fmt.Errorf("%s: %w", op, err)
		}

		filmFound = foundFilm
		filmFound.Listactors = append(filmFound.Listactors, foundActor) // Добавляем актёра
	}

	return filmFound, nil
}

func (s *Storage) GetActorsWithFilms() (map[Actor][]Film, error) {
	const op = "storage.postgres.GetActorsWithFilms"

	query := `SELECT a.id, a.name, a.gender, a.date_of_birth, f.id, f.name, f.description, f.release_date, f.rating
	FROM actor a
	JOIN actor_film af ON a.id = af.actor_id
	JOIN film f ON f.id = af.film_id`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	actorsWithFilms := make(map[Actor][]Film)

	for rows.Next() {
		var actor Actor
		var film Film

		err = rows.Scan(&actor.Id, &actor.Name, &actor.Gender, &actor.DateOfBirth, &film.Id, &film.Name, &film.Description, &film.Releasedate, &film.Rating)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		// Проверка на наличие актера в мапе
		if _, exists := actorsWithFilms[actor]; !exists {
			actorsWithFilms[actor] = []Film{}
		}
		actorsWithFilms[actor] = append(actorsWithFilms[actor], film)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return actorsWithFilms, nil
}
