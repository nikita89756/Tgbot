package storage

import (
	"bot/internal/models"
	"bot/internal/storage/postgres"

	"go.uber.org/zap"
)



type Storage interface{
	Close()
	GetUserByEmail(email , password string ) (models.User, error)
	InsertCourse(title, link string,knowledges []string, rating float64, coins int) (int,error)
	InsertNewUser(id int) error
	UpdateUser(id int, experience int, coins int) error
	GetUser(id int) (int, int, error)
	GetCourse(title string) ( int ,string, float64, int, error)
	InsertGrade(userID, topicID, grade int) error
	GetCourseID(title string) (int, error)
	GetTopics(courseID int) ([]string,[]int,[]float64, error)
	GetItems() ([]models.Item, error)
	InsertItem(name string, price int, multiplier float64) error 
	InsertUserCourses(userID int, courseID int) error
	InsertUserItems(userID int, itemID int,itemPrice int) error
	GetUserItems(id int) ([]models.Item,error)
	GetCourseByID(id int) ( string , string, float64, int, error)
	GetCourseByLink(link string) ( int , string,[]string,[]float64, error)
	InsertTask(userID int,task string)error
	GetTasks(userID int) ([]models.Tasks, error)
	UpdateTaskStatus(taskID int) error 
}

func NewStorage(typeStorage string,connectionString string,logger *zap.Logger) Storage{
	switch typeStorage {
	case "postgres":
		return postgres.NewStorage(connectionString,logger)
	case "mysql":
		return nil
	}
	return nil
}