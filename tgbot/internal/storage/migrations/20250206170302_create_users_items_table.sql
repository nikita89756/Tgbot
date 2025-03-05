-- +goose Up
-- +goose StatementBegin
CREATE TABLE users_items(
  user_id INT NOT NULL REFERENCES users(id),
  item_id INT NOT NULL REFERENCES items(id)
  UNIQUE(user_id, item_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
