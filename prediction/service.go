package prediction

import (
	"errors"

	"github.com/AndreyBeloborodov/textlab/algorithm"
	"github.com/AndreyBeloborodov/textlab/models"
)

var (
	ErrModelNotFound       = errors.New("model not found")
	ErrModelNotTrained     = errors.New("model is not trained")
	ErrAlgorithmUnavailable = errors.New("algorithm factory is not configured")
	ErrEmptyInput          = errors.New("prediction input must not be empty")
)

type Service interface {
	Predict(ownerID, modelID int, input string) (string, error)
}

type service struct {
	models  models.Service
	store   algorithm.Store
	factory algorithm.Factory
}

func NewService(
	modelService models.Service,
	store algorithm.Store,
	factory algorithm.Factory,
) Service {
	return &service{
		models:  modelService,
		store:   store,
		factory: factory,
	}
}

func (s *service) Predict(
	ownerID, modelID int,
	input string,
) (string, error) {
	if input == "" {
		return "", ErrEmptyInput
	}

	// Проверяем существование модели и права пользователя.
	if _, err := s.models.Get(ownerID, modelID); err != nil {
		return "", ErrModelNotFound
	}

	if s.factory == nil || s.store == nil {
		return "", ErrAlgorithmUnavailable
	}

	// Загружаем обученное состояние.
	state, err := s.store.Get(algorithm.ModelKey(modelID))
	if err != nil {
		if errors.Is(err, algorithm.ErrNotFound) {
			return "", ErrModelNotTrained
		}
		return "", err
	}

	// Создаём алгоритм и выполняем предсказание.
	alg := s.factory()
	if alg == nil {
		return "", ErrAlgorithmUnavailable
	}

	result, err := alg.Predict(state, []byte(input))
	if err != nil {
		return "", err
	}

	return string(result), nil
}
