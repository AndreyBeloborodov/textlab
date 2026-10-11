
package training

import (
    "errors"
    "strings"
    "time"

    "github.com/AndreyBeloborodov/textlab/models"
)

var (
    ErrModelNotFound = errors.New("model not found")
    ErrEmptyText     = errors.New("training text must not be empty")
)

type Service interface {
    Start(ownerID, modelID int, text string) (Training, error)
    GetStatus(ownerID, trainingID int) (Training, error)
}

type service struct {
    models  models.Service
    repo    Repository
    manager Manager
}

func NewService(
    modelService models.Service,
    repo Repository,
    manager Manager,
) Service {
    return &service{
        models:  modelService,
        repo:    repo,
        manager: manager,
    }
}

func (s *service) Start(
    ownerID, modelID int,
    text string,
) (Training, error) {
    if strings.TrimSpace(text) == "" {
        return Training{}, ErrEmptyText
    }

    if _, err := s.models.Get(ownerID, modelID); err != nil {
        return Training{}, ErrModelNotFound
    }

    t, err := s.repo.Create(Training{
        ModelID:   modelID,
        Text:      text,
        Status:    StatusPending,
        CreatedAt: time.Now().UTC(),
    })
    if err != nil {
        return Training{}, err
    }

    if err := s.manager.Submit(t.ID); err != nil {
        t.Status = StatusFailed
        t.Error = err.Error()

        if updateErr := s.repo.Update(t); updateErr != nil {
            return Training{}, updateErr
        }

        return t, err
    }

    // Менеджер уже обновил статус до running.
    // Перечитываем запись, чтобы вернуть актуальное состояние.
    return s.repo.Get(t.ID)
}

func (s *service) GetStatus(
    ownerID, trainingID int,
) (Training, error) {
    t, err := s.repo.Get(trainingID)
    if err != nil {
        return Training{}, ErrNotFound
    }

    if _, err := s.models.Get(ownerID, t.ModelID); err != nil {
        // Не раскрываем существование чужих задач.
        return Training{}, ErrNotFound
    }

    return t, nil
}
