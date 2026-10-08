package models

import "errors"

var ErrNotFound = errors.New("model not found")

type modelRecord struct {
	ID   int
	Name string
}

type Repository interface {
	Create(model modelRecord) (modelRecord, error)
	Get(id int) (modelRecord, error)
	GetAll() ([]modelRecord, error)
	Update(model modelRecord) (modelRecord, error)
	Delete(id int) error
}