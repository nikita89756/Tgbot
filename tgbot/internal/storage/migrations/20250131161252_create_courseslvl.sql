-- +goose Up
-- +goose StatementBegin
CREATE TABLE courses(
  id SERIAL PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  link VARCHAR(255) NOT NULL UNIQUE,
  rating FLOAT,
  coins INT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
