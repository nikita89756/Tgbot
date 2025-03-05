-- +goose Up
-- +goose StatementBegin
CREATE TABLE topics(
  id SERIAL PRIMARY KEY,
  course_id INT NOT NULL REFERENCES courses(id),
  title VARCHAR(255) NOT NULL,
  avg_grade DECIMAL(5, 2)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
