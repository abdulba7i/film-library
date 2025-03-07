-- +goose Up
CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE
)

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    role_id INT REFERENCES roles(id)
)

CREATE TABLE IF NOT EXISTS actor (
    id SERIAL PRIMARY KEY,
	name VARCHAR(255) NOT NULL,
    gender VARCHAR(255) NOT NULL CHECK (gender IN ('male', 'female', 'other')),
    date_of_birth VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS film (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description VARCHAR(1000) NOT NULL,
    release_date DATE NOT NULL,
    rating NUMERIC(2, 1) CHECK (rating >= 0 AND rating <= 10) NOT NULL
);

CREATE TABLE IF NOT EXISTS actor_film (
    film_id INT REFERENCES film(id),
    actor_id INT REFERENCES actor(id),
    PRIMARY KEY (film_id, actor_id)
);

INSERT INTO roles (name) VALUES ('user'), ('admin') ON CONFLICT (name)DO NOTHING;

CREATE INDEX idx_film_name ON film (name);
CREATE INDEX idx_actor_name ON actor (name);
CREATE INDEX idx_film_rating ON film (rating);
CREATE INDEX idx_film_release_date ON film (release_date);
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
DROP TABLE IF EXISTS actor_film;
DROP TABLE IF EXISTS film;
DROP TABLE IF EXISTS actor;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS roles;