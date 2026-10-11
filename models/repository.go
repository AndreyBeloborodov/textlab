package models

import "errors"

var (
	ErrNotFound = errors.New("model not found")
)

type Record struct {
	ID      int
	OwnerID int
	Name    string
}

type Repository interface {
	Create(model Record) (Record, error)
	Get(id int) (Record, error)
	GetAll() ([]Record, error)
	Update(model Record) (Record, error)
	Delete(id int) error
}
