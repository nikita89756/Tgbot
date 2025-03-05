-- +goose Up
-- +goose StatementBegin
CREATE TABLE items(
  id SERIAL PRIMARY KEY,
  name varchar(255) NOT NULL,
  price INT NOT NULL,
  multiplier FLOAT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
