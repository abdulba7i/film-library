-- +goose Up
CREATE TABLE IF NOT EXISTS actor (
    id SERIAL PRIMARY KEY,
	name VARCHAR(255) NOT NULL,
    bio VARCHAR(255) NOT NULL,
    date_of_birth VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS film (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    release_date VARCHAR(255) NOT NULL,
    rating NUMERIC(2, 1) NOT NULL
);
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
