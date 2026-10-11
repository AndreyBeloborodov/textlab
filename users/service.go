package users

import (
	"errors"
	"strings"
)

var ErrInvalidUser = errors.New("invalid user")

type Service interface {
	Create(login, passwordHash string) (User, error)
	Get(id int) (User, error)
	GetByLogin(login string) (Credentials, error)
	GetAll() ([]User, error)
	Update(id int, login string) (User, error)
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

func (s *service) Create(login, passwordHash string) (User, error) {
	login = strings.TrimSpace(login)

	if login == "" || passwordHash == "" {
		return User{}, ErrInvalidUser
	}

	record, err := s.repository.Create(Record{
		Login:        login,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return User{}, err
	}

	return userFromRecord(record), nil
}

func (s *service) Get(id int) (User, error) {
	record, err := s.repository.Get(id)
	if err != nil {
		return User{}, err
	}

	return userFromRecord(record), nil
}

func (s *service) GetByLogin(login string) (Credentials, error) {
	login = strings.TrimSpace(login)

	if login == "" {
		return Credentials{}, ErrNotFound
	}

	record, err := s.repository.GetByLogin(login)
	if err != nil {
		return Credentials{}, err
	}

	return Credentials{
		User:         userFromRecord(record),
		PasswordHash: record.PasswordHash,
	}, nil
}

func (s *service) GetAll() ([]User, error) {
	records, err := s.repository.GetAll()
	if err != nil {
		return nil, err
	}

	result := make([]User, 0, len(records))

	for _, record := range records {
		result = append(result, userFromRecord(record))
	}

	return result, nil
}

func (s *service) Update(id int, login string) (User, error) {
	login = strings.TrimSpace(login)

	if login == "" {
		return User{}, ErrInvalidUser
	}

	record, err := s.repository.Get(id)
	if err != nil {
		return User{}, err
	}

	record.Login = login

	updated, err := s.repository.Update(record)
	if err != nil {
		return User{}, err
	}

	return userFromRecord(updated), nil
}

func (s *service) Delete(id int) error {
	return s.repository.Delete(id)
}

func userFromRecord(record Record) User {
	return User{
		ID:    record.ID,
		Login: record.Login,
	}
}
