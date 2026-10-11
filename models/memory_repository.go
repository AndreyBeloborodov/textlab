package models

import "sync"

type MemoryRepository struct {
	mu     sync.RWMutex
	models map[int]Record
	nextID int
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		models: make(map[int]Record),
		nextID: 1,
	}
}

func (r *MemoryRepository) Create(model Record) (Record, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	model.ID = r.nextID
	r.nextID++

	r.models[model.ID] = model

	return model, nil
}

func (r *MemoryRepository) Get(id int) (Record, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	model, ok := r.models[id]
	if !ok {
		return Record{}, ErrNotFound
	}

	return model, nil
}

func (r *MemoryRepository) GetAll() ([]Record, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Record, 0, len(r.models))

	for _, model := range r.models {
		result = append(result, model)
	}

	return result, nil
}

func (r *MemoryRepository) Update(model Record) (Record, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.models[model.ID]; !ok {
		return Record{}, ErrNotFound
	}

	r.models[model.ID] = model

	return model, nil
}

func (r *MemoryRepository) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.models[id]; !ok {
		return ErrNotFound
	}

	delete(r.models, id)

	return nil
}
