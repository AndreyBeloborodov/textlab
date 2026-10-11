package users

import "errors"

var (
	ErrNotFound   = errors.New("user not found")
	ErrLoginTaken = errors.New("login already taken")
)

type Record struct {
	ID           int
	Login        string
	PasswordHash string
}

type Credentials struct {
	User         User
	PasswordHash string
}

type Repository interface {
	Create(user Record) (Record, error)
	Get(id int) (Record, error)
	GetByLogin(login string) (Record, error)
	GetAll() ([]Record, error)
	Update(user Record) (Record, error)
	Delete(id int) error
}
