package models

import "strings"

type Service interface {
	Create(name string) (Model, error)
	Get(id int) (Model, error)
	GetAll() ([]Model, error)
	Update(id int, name string) (Model, error)
	Delete(id int) error
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{
		repository: repository,
	}
}

func (s *service) Create(name string) (Model, error) {
	name = strings.TrimSpace(name)

	record, err := s.repository.Create(modelRecord{
		Name: name,
	})
	if err != nil {
		return Model{}, err
	}

	return modelFromRecord(record), nil
}

func (s *service) Get(id int) (Model, error) {
	record, err := s.repository.Get(id)
	if err != nil {
		return Model{}, err
	}

	return modelFromRecord(record), nil
}

func (s *service) GetAll() ([]Model, error) {
	records, err := s.repository.GetAll()
	if err != nil {
		return nil, err
	}

	models := make([]Model, 0, len(records))

	for _, record := range records {
		models = append(models, modelFromRecord(record))
	}

	return models, nil
}

func (s *service) Update(id int, name string) (Model, error) {
	name = strings.TrimSpace(name)

	record, err := s.repository.Update(modelRecord{
		ID:   id,
		Name: name,
	})
	if err != nil {
		return Model{}, err
	}

	return modelFromRecord(record), nil
}

func (s *service) Delete(id int) error {
	return s.repository.Delete(id)
}

func modelFromRecord(record modelRecord) Model {
	return Model{
		ID:   record.ID,
		Name: record.Name,
	}
}