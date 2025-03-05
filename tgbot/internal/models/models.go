package models

type User struct {
	ID       int    `json: user_id`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Item struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Price int `json:"price"`
	Multiplier float64 `json:"multiplier"`
}

type Tasks struct{
	ID int 
	Task string
}