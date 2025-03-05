-- +goose Up
-- +goose StatementBegin
CREATE TABLE users(
  id INT PRIMARY KEY,
  experience INT NOT NULL,
  coins int
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
