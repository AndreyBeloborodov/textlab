
package training

import (
    "sort"
    "sync"
)

type MemoryRepository struct {
    mu        sync.RWMutex
    nextID    int
    trainings map[int]Training
}

func NewMemoryRepository() *MemoryRepository {
    return &MemoryRepository{
        nextID:    1,
        trainings: make(map[int]Training),
    }
}

func (r *MemoryRepository) Create(t Training) (Training, error) {
    r.mu.Lock()
    defer r.mu.Unlock()

    t.ID = r.nextID
    r.nextID++

    r.trainings[t.ID] = t

    return t, nil
}

func (r *MemoryRepository) Get(id int) (Training, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    t, ok := r.trainings[id]
    if !ok {
        return Training{}, ErrNotFound
    }

    return t, nil
}

func (r *MemoryRepository) Update(t Training) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    if _, ok := r.trainings[t.ID]; !ok {
        return ErrNotFound
    }

    r.trainings[t.ID] = t
    return nil
}

func (r *MemoryRepository) GetByModelID(modelID int) ([]Training, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    result := make([]Training, 0)

    for _, t := range r.trainings {
        if t.ModelID == modelID {
            result = append(result, t)
        }
    }

    sort.Slice(result, func(i, j int) bool {
        return result[i].ID < result[j].ID
    })

    return result, nil
}
