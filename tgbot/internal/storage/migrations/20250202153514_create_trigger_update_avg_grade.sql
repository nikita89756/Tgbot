-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_avg_grade()
  RETURNS TRIGGER AS $$
  BEGIN
    
    UPDATE topics
    SET avg_grade = (
      SELECT AVG(grade)
      FROM users_topic_grades
      WHERE topic_id = NEW.topic_id
    ) 
    WHERE id = NEW.topic_id;
    RETURN NEW;
  END;
  $$ LANGUAGE plpgsql;

CREATE Trigger update_avg_grade_trigger
AFTER INSERT ON users_topic_grades
FOR EACH ROW 
EXECUTE FUNCTION update_avg_grade();
  
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
