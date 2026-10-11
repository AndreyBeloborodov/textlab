package models

import (
	"errors"
	"strings"
)

var (
	ErrInvalidName = errors.New("invalid model name")
	ErrInvalidOwner = errors.New("invalid owner id")
)

type Service interface {
	Create(ownerID int, name string) (Model, error)
	Get(ownerID, id int) (Model, error)
	GetAll(ownerID int) ([]Model, error)
	Update(ownerID, id int, name string) (Model, error)
	Delete(ownerID, id int) error
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{
		repository: repository,
	}
}

func (s *service) Create(ownerID int, name string) (Model, error) {
	name = strings.TrimSpace(name)

	if ownerID <= 0 {
		return Model{}, ErrInvalidOwner
	}

	if name == "" {
		return Model{}, ErrInvalidName
	}

	record, err := s.repository.Create(Record{
		OwnerID: ownerID,
		Name:    name,
	})
	if err != nil {
		return Model{}, err
	}

	return modelFromRecord(record), nil
}

func (s *service) Get(ownerID, id int) (Model, error) {
	record, err := s.repository.Get(id)
	if err != nil {
		return Model{}, err
	}

	if record.OwnerID != ownerID {
		return Model{}, ErrNotFound
	}

	return modelFromRecord(record), nil
}

func (s *service) GetAll(ownerID int) ([]Model, error) {
	if ownerID <= 0 {
		return nil, ErrInvalidOwner
	}

	records, err := s.repository.GetAll()
	if err != nil {
		return nil, err
	}

	result := make([]Model, 0)

	for _, record := range records {
		if record.OwnerID == ownerID {
			result = append(result, modelFromRecord(record))
		}
	}

	return result, nil
}

func (s *service) Update(ownerID, id int, name string) (Model, error) {
	name = strings.TrimSpace(name)

	if ownerID <= 0 {
		return Model{}, ErrInvalidOwner
	}

	if name == "" {
		return Model{}, ErrInvalidName
	}

	record, err := s.repository.Get(id)
	if err != nil {
		return Model{}, err
	}

	if record.OwnerID != ownerID {
		return Model{}, ErrNotFound
	}

	record.Name = name

	updated, err := s.repository.Update(record)
	if err != nil {
		return Model{}, err
	}

	return modelFromRecord(updated), nil
}

func (s *service) Delete(ownerID, id int) error {
	if ownerID <= 0 {
		return ErrInvalidOwner
	}

	record, err := s.repository.Get(id)
	if err != nil {
		return err
	}

	if record.OwnerID != ownerID {
		return ErrNotFound
	}

	return s.repository.Delete(id)
}

func modelFromRecord(record Record) Model {
	return Model{
		ID:      record.ID,
		OwnerID: record.OwnerID,
		Name:    record.Name,
	}
}
