package postgres

import (
	"bot/internal/errors"
	"bot/internal/models"
	"database/sql"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

type Storage struct {
	db *sqlx.DB
	logger *zap.Logger
}

func NewStorage(connectionString string,logger *zap.Logger) Storage{
	db,err:=sqlx.Open("postgres",connectionString)
	if err!= nil{
		panic(err)
	}
		err = db.Ping()
	if err != nil {
		logger.Fatal("Database connection error", zap.Error(err))
	}
	return Storage{db: db,logger: logger}
}

func (s Storage)InsertTask(userID int,task string)error{
	_, err := s.db.Exec("INSERT INTO tasks (user_id, task, active) VALUES ($1, $2, $3)", userID, task, true)
	return err}

func (s Storage)UpdateTaskStatus(taskID int) error {
	_, err := s.db.Exec("UPDATE tasks SET active = $1 WHERE id = $2", false, taskID)
	return err
}

func (s Storage)GetTasks(userID int) ([]models.Tasks, error) {
	var tasks = make([]models.Tasks,0,2)
	rows,err := s.db.Query("SELECT id,task FROM tasks WHERE user_id = $1 AND active = $2", userID, true)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var task models.Tasks
		if err := rows.Scan(&task.ID, &task.Task); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, err
}
func (s Storage)InsertNewUser(id int) error {
	_, err := s.db.Exec("INSERT INTO users (id, experience, coins) VALUES ($1, $2, $3)", id, 0, 0)
	if err !=nil {
			if pqErr,ok:=err.(*pq.Error);ok && pqErr.Code == "23505"{
				return errors.ErrUserAlreadyExists
			}
			return err
		}
	return err
}

func (s Storage)UpdateUser(id int, experience int, coins int) error {
	exp,coin,err:= s.GetUser(id)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("UPDATE users SET experience = $1, coins = $2 WHERE id = $3", experience+exp, coins+coin, id)
	return err
}
func (s Storage)GetUser(id int) (int, int, error) {
	var  experience, coins int
	err := s.db.QueryRow("SELECT experience, coins FROM users WHERE id = $1", id).
		Scan(&experience, &coins)
	return experience, coins, err
}

func (s Storage)GetCourse(title string) ( int , string, float64, int, error) {
	var link string
	var rating float64
	var id , coins int
	err := s.db.QueryRow("SELECT id, link rating, coins FROM courses WHERE title = $1", title).
		Scan(&id, &link, &rating, &coins)
	return id, link, rating, coins, err
}

func (s Storage)GetCourseByLink(link string) ( int , string,[]string,[]float64,error){
	var id int
	var title string
	rows,err := s.db.Query("SELECT id, title FROM courses WHERE link = $1", link)
	if !rows.Next() {
		return 0, "", []string{}, []float64{},sql.ErrNoRows
	}
	if err != nil {
		return 0, "", []string{},[]float64{},err 
	}
	if err := rows.Scan(&id, &title); err != nil {
		return 0, "", []string{}, []float64{},err
	}
	topics,_,grades,err := s.GetTopics(id)
	if err != nil {
		return 0, "", []string{}, []float64{},err
	}
	return id, title, topics,grades, nil
}
func (s Storage)GetCourseByID(id int) ( string, string, float64, int, error){
	var link, title string
	var rating float64
	var coins int
	err := s.db.QueryRow("SELECT link, title, rating, coins FROM courses WHERE id = $1", id).
		Scan(&link, &title, &rating, &coins)
	return link, title,  rating, coins, err
}

func (s Storage)GetCourseID(title string) (int, error) {
	    rows, err := s.db.Query("SELECT id FROM courses WHERE title = $1", title)
    if err != nil {
        return 0, err
    }
    defer rows.Close()

    if !rows.Next() {
        return 0, sql.ErrNoRows
    }
				var id int
    if err := rows.Scan(&id); err != nil {
        return 0, err
    }

    return id, nil
}

func (s Storage)GetTopics(courseID int) ([]string,[]int,[]float64, error) {
	var topics []string
	var ids []int
	var avgGrades []float64
	rows,err:=s.db.Query("SELECT id,title,avg_grade FROM topics WHERE course_id = $1", courseID)
	if err != nil {
		return nil,[]int{},[]float64{}, err
	}
	for rows.Next() {
		var topic string
		var id int
		var grade float64
		err = rows.Scan(&id, &topic,&grade)
		if err != nil {
			return nil,[]int{},[]float64{}, err
		}
		topics = append(topics, topic)
		ids = append(ids, id)
		avgGrades = append(avgGrades, grade)
	}
	return topics, ids,avgGrades, nil
}

func (s Storage)InsertGrade(userID, topicID, grade int) error {
	_, err := s.db.Exec("INSERT INTO users_topic_grades (user_id, topic_id, grade) VALUES ($1, $2, $3)",
		userID, topicID, grade)
		if err !=nil {
			if pqErr,ok:=err.(*pq.Error);ok && pqErr.Code == "23505"{
				return errors.ErrGradeAlreadyExists
			}
			return err
		}
	return err
}
func (s Storage)InsertCourse(title, link string,knowledges []string, rating float64, coins int) (int,error) {
	var id int
	tx , err:=s.db.Begin()
	if err != nil {
		return 0,err
	}
	s.logger.Info("insert course",zap.String("title",title),zap.String("link",link),zap.Float64("rating",rating),zap.Int("coins",coins))
	err = tx.QueryRow("INSERT INTO courses (title, link, rating, coins) VALUES ($1, $2, $3, $4) RETURNING id",
		title, link, rating, coins).Scan(&id)
	if err != nil {
		tx.Rollback()
		return 0,err
	}
		for i:=0;i<len(knowledges);i++{
							knowledge := knowledges[i]
							if len(knowledge)>255{
								knowledge = removeLastWord(knowledge[:255])
							}
							_, err = tx.Exec("INSERT INTO topics (course_id, title) VALUES ($1, $2)", id, knowledge)
							if err != nil {
								tx.Rollback()
								s.logger.Error("Database error", zap.Error(err),zap.String("title",knowledge),zap.Int("len",len(knowledges[i])))
								return 0,err
							}
		}
	err = tx.Commit()
	if err != nil {
			tx.Rollback()
			return 0,err
	}
	return id,err
}

func (s Storage)InsertItem(name string, price int, multiplier float64) error {
	_, err := s.db.Exec("INSERT INTO items (name, price, multiplier) VALUES ($1, $2, $3)", name, price, multiplier)
	return err
}

func (s Storage)InsertUserCourses(userID int, courseID int) error {
	_, err := s.db.Exec("INSERT INTO users_courses (user_id, course_id) VALUES ($1, $2)", userID, courseID)
	if pqErr,ok:=err.(*pq.Error);ok && pqErr.Code == "23505"{
				return errors.ErrUserCourseAlreadyExists
	}
	return err
}

func (s Storage)InsertUserItems(userID int, itemID int,itemPrice int) error {
	tx , err := s.db.Begin()
	if err!=nil{
		return err
	}

	_, err = tx.Exec("INSERT INTO users_items (user_id, item_id) VALUES ($1, $2)", userID, itemID)
	if err !=nil{
		if pqErr,ok:=err.(*pq.Error);ok && pqErr.Code == "23505"{
				tx.Rollback()
				return errors.ErrUserItemAlreadyExists
			}
		tx.Rollback()
		return err
	}
	var coins int
	err = s.db.QueryRow("SELECT coins FROM users WHERE id = $1", userID).
		Scan(&coins)
	if coins<itemPrice{
		tx.Rollback()
		return errors.ErrNotEnothCoins
	}
	_,err = tx.Exec("UPDATE users SET coins = coins - $1 WHERE id = $2",itemPrice,userID)
	if err != nil{
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s Storage)GetUserItems(id int) ([]models.Item,error) {
	var items []models.Item
	rows,err := s.db.Query("SELECT items.multiplier, items.price, items.name FROM users_items join items on users_items.item_id = items.id WHERE user_id = $1", id)
	if err != nil {
		return []models.Item{},err
	}
	defer rows.Close()
	for rows.Next() {
		var item models.Item
		err = rows.Scan(&item.Multiplier, &item.Price, &item.Name)
		if err != nil {
			return []models.Item{},err
		}
		items = append(items, item)
	}
	return items , nil
}
func (s Storage)GetItems() ([]models.Item, error) {
	var items []models.Item
	rows, err := s.db.Query("SELECT id, name, price,multiplier FROM items")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Price,&item.Multiplier); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (s Storage)Close(){
	s.db.Close()
}

func (s Storage) GetUserByEmail(email , password string ) (models.User, error) {
	user:= models.User{Email:email}
	s.logger.Info("GetUserByEmail",zap.String("email",email),zap.String("password",password))
	query := `SELECT user_id , password from admins where email = $1`
	row := s.db.QueryRow(query,email)
	if err:= row.Scan(&user.ID,&user.Password); err!=nil{
		return models.User{},err
	}
	if user.Password != password{
		return models.User{} , errors.ErrUncorrectPassword
	}
	return user , nil
}



func removeLastWord(input string) string {
	// Находим последний пробел
	lastSpaceIndex := strings.LastIndex(input, " ")
	if lastSpaceIndex == -1 {
		// Если пробела нет, возвращаем пустую строку или исходную строку
		return ""
	}
	// Обрезаем строку до последнего пробела
	return input[:lastSpaceIndex]
}




