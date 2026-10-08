package models

import "sync"

type MemoryRepository struct {
	mu     sync.RWMutex
	models map[int]modelRecord
	nextID int
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		models: make(map[int]modelRecord),
		nextID: 1,
	}
}

func (r *MemoryRepository) Create(model modelRecord) (modelRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	model.ID = r.nextID
	r.nextID++

	r.models[model.ID] = model

	return model, nil
}

func (r *MemoryRepository) Get(id int) (modelRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	model, ok := r.models[id]
	if !ok {
		return modelRecord{}, ErrNotFound
	}

	return model, nil
}

func (r *MemoryRepository) GetAll() ([]modelRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]modelRecord, 0, len(r.models))

	for _, model := range r.models {
		result = append(result, model)
	}

	return result, nil
}

func (r *MemoryRepository) Update(model modelRecord) (modelRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.models[model.ID]; !ok {
		return modelRecord{}, ErrNotFound
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