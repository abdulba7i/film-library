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
	Gender      string `json:"gender"`
	DateOfBirth string `json:"date_of_birth"`
}

type Film struct {
	Id          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Releasedate string  `json:"release_date"`
	Rating      float32 `json:"rating"`
	ListActors  []Actor `json:"list_actors"`
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

func (s *Storage) AddedInfoActor(actor *Actor) error {
	const op = "storage.postgres.AddedInfoActor"

	query := `INSERT INTO actors (name, gender, date_of_birth) VALUES ($1, $2, $3) RETURNING id`
	err := s.db.QueryRow(query, actor.Name, actor.Gender, actor.DateOfBirth).Scan(&actor.Id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) UpdateActor(actor Actor) error {
	const op = "storage.postgres.ChangeInfoActor"

	query := `UPDATE actors SET name = $1, gender = $2, date_of_birth = $3 WHERE id = $4`

	_, err := s.db.Exec(query, actor.Name, actor.Gender, actor.DateOfBirth, actor.Id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) DeleteInfoActor(id int) error {
	const op = "storage.postgres.DeleteInfoActor"

	query := `DELETE FROM actors WHERE id = $1`

	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// Film:

func (s *Storage) AddedInfoFilm(film Film) error {
	const op = "storage.postgres.AddedInfoFilm"

	// Начинаем транзакцию
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("%s: failed to begin transaction: %w", op, err)
	}
	defer tx.Rollback()

	// Добавляем фильм
	var filmID int
	err = tx.QueryRow(`
        INSERT INTO films (name, description, release_date, rating)
        VALUES ($1, $2, $3, $4)
        RETURNING id`,
		film.Name, film.Description, film.Releasedate, film.Rating,
	).Scan(&filmID)
	if err != nil {
		return fmt.Errorf("%s: failed to insert film: %w", op, err)
	}

	// Добавляем актёров
	for _, actor := range film.ListActors {
		var actorID int

		// Проверяем, существует ли актёр
		err := tx.QueryRow(`
            SELECT id FROM actors WHERE name = $1`,
			actor.Name,
		).Scan(&actorID)

		if err != nil {
			if err == sql.ErrNoRows {
				// Актёр не существует, создаём нового
				err = tx.QueryRow(`
                    INSERT INTO actors (name, gender, date_of_birth)
                    VALUES ($1, $2, $3)
                    RETURNING id`,
					actor.Name, actor.Gender, actor.DateOfBirth,
				).Scan(&actorID)
				if err != nil {
					return fmt.Errorf("%s: failed to insert actor: %w", op, err)
				}
			} else {
				return fmt.Errorf("%s: failed to check actor existence: %w", op, err)
			}
		}

		// Связываем фильм и актёра
		_, err = tx.Exec(`
            INSERT INTO actor_film (film_id, actor_id)
            VALUES ($1, $2)`,
			filmID, actorID,
		)
		if err != nil {
			return fmt.Errorf("%s: failed to insert film-actor link: %w", op, err)
		}
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s: failed to commit transaction: %w", op, err)
	}

	return nil
}

func (s *Storage) UpdateFilm(film Film) error {
	const op = "storage.postgres.ChangeInfoFilm"

	query := `UPDATE films SET name = $1, description = $2, release_date = $3, rating = $4 WHERE id = $5`

	_, err := s.db.Exec(query, film.Name, film.Description, film.Releasedate, film.Rating, film.Id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) DeleteInfoFilm(id int) error {
	const op = "storage.postgres.DeleteInfoFilm"

	query := `DELETE FROM films WHERE id = $1`

	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// func (s *Storage) GetAllFilms(sortBy string) ([]Film, error) {
// 	const op = "storage.postgres.GetAllFilms"

// 	orderClause := "ORDER BY rating DESC"
// 	switch sortBy {
// 	case "name":
// 		orderClause = "ORDER BY name"
// 	case "release_date":
// 		orderClause = "ORDER BY release_date"
// 	}

// 	query := fmt.Sprintf(`
// 	SELECT f.id, f.name, f.description, f.release_date, f.rating
// 	FROM films f
// 		%s`, orderClause)

// 	var films []Film

// 	rows, err := s.db.Query(query)
// 	if err != nil {
// 		return nil, fmt.Errorf("%s: %w", op, err)
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var film Film
// 		if err := rows.Scan(&film.Id, &film.Name, &film.Description, &film.Releasedate, &film.Rating); err != nil {
// 			return nil, fmt.Errorf("%s: %w", op, err)
// 		}
// 		films = append(films, film)
// 	}

// 	return films, nil
// }

func (s *Storage) GetAllFilms(sortBy string) ([]Film, error) {
	const op = "storage.postgres.GetAllFilms"

	orderClause := "ORDER BY rating DESC" // По умолчанию сортировка по рейтингу
	switch sortBy {
	case "name":
		orderClause = "ORDER BY name"
	case "release_date":
		orderClause = "ORDER BY release_date"
	}

	query := fmt.Sprintf(`
        SELECT f.id, f.name, f.description, f.release_date, f.rating, 
               a.id, a.name, a.gender, a.date_of_birth
        FROM films f
        JOIN actor_film af ON f.id = af.film_id
        JOIN actors a ON a.id = af.actor_id
        %s`, orderClause)

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	filmMap := make(map[int]*Film)

	for rows.Next() {
		var film Film
		var actor Actor

		err = rows.Scan(
			&film.Id, &film.Name, &film.Description, &film.Releasedate, &film.Rating,
			&actor.Id, &actor.Name, &actor.Gender, &actor.DateOfBirth,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		// Если фильм ещё не добавлен в filmMap, добавляем его
		if _, exists := filmMap[film.Id]; !exists {
			film.ListActors = []Actor{} // Инициализируем пустой слайс актёров
			filmMap[film.Id] = &film
		}

		// Добавляем актёра к фильму
		filmMap[film.Id].ListActors = append(filmMap[film.Id].ListActors, actor)
	}

	// Преобразуем filmMap в слайс фильмов
	films := make([]Film, 0, len(filmMap))
	for _, film := range filmMap {
		films = append(films, *film)
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
	filmFound.ListActors = []Actor{} // Инициализация списка актёров

	if rows.Next() {
		var foundFilm Film
		var foundActor Actor

		err = rows.Scan(&foundFilm.Id, &foundFilm.Name, &foundFilm.Description, &foundFilm.Releasedate, &foundFilm.Rating, &foundActor.Id, &foundActor.Name, &foundActor.Gender, &foundActor.DateOfBirth)
		if err != nil {
			return Film{}, fmt.Errorf("%s: %w", op, err)
		}

		filmFound = foundFilm
		filmFound.ListActors = append(filmFound.ListActors, foundActor) // Добавляем актёра
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
