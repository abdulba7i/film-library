-- +goose Up
CREATE TABLE IF NOT EXISTS actor (
    id SERIAL PRIMARY KEY,
	name VARCHAR(255) NOT NULL,
    gender VARCHAR(255) NOT NULL,
    date_of_birth VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS film (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description VARCHAR(1000) NOT NULL,
    release_date VARCHAR(255) NOT NULL,
    rating NUMERIC(2, 1) CHECK (rating >= 0 AND rating <= 10) NOT NULL
);

CREATE TABLE IF NOT EXISTS actor_film (
    film_id INT REFERENCES film(id),
    actor_id INT REFERENCES actor(id),
    PRIMARY KEY (film_id, actor_id)

    -- actor_id INT NOT NULL,
    -- film_id INT NOT NULL,
    -- PRIMARY KEY (actor_id, film_id),
    -- FOREIGN KEY (actor_id) REFERENCES actor (id),
    -- FOREIGN KEY (film_id) REFERENCES film (id)
);


-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
