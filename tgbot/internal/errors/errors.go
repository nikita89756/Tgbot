package errors

import (
	"errors"
)
var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrGradeAlreadyExists = errors.New("grade already exists")
	ErrUncorrectPassword = errors.New("uncorrect password")
	ErrUserCourseAlreadyExists = errors.New("user course already exists")
	ErrUserItemAlreadyExists = errors.New("user have this item")
	ErrNotEnothCoins = errors.New("not enough coins")
)