-- +goose Up
-- +goose StatementBegin
CREATE TABLE users_courses(
  user_id INT NOT NULL REFERENCES users(id),
  course_id INT NOT NULL REFERENCES courses(id),
  PRIMARY KEY (user_id, course_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
