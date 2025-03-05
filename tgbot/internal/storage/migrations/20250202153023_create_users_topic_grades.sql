-- +goose Up
-- +goose StatementBegin
CREATE TABLE users_topic_grades(
  user_id INT NOT NULL REFERENCES users(id),
  topic_id INT NOT NULL REFERENCES topics(id),
  grade INT NOT NULL,
  PRIMARY KEY (user_id, topic_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
