package users

import (
	"sync"
)

type MemoryRepository struct {
	mu      sync.RWMutex
	users   map[int]Record
	nextID  int
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		users:  make(map[int]Record),
		nextID: 1,
	}
}

func (r *MemoryRepository) Create(user Record) (Record, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, existing := range r.users {
		if existing.Login == user.Login {
			return Record{}, ErrLoginTaken
		}
	}

	user.ID = r.nextID
	r.nextID++

	r.users[user.ID] = user

	return user, nil
}

func (r *MemoryRepository) Get(id int) (Record, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return Record{}, ErrNotFound
	}

	return user, nil
}

func (r *MemoryRepository) GetByLogin(login string) (Record, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Login == login {
			return user, nil
		}
	}

	return Record{}, ErrNotFound
}

func (r *MemoryRepository) GetAll() ([]Record, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Record, 0, len(r.users))

	for _, user := range r.users {
		result = append(result, user)
	}

	return result, nil
}

func (r *MemoryRepository) Update(user Record) (Record, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[user.ID]; !ok {
		return Record{}, ErrNotFound
	}

	for id, existing := range r.users {
		if id != user.ID && existing.Login == user.Login {
			return Record{}, ErrLoginTaken
		}
	}

	r.users[user.ID] = user

	return user, nil
}

func (r *MemoryRepository) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[id]; !ok {
		return ErrNotFound
	}

	delete(r.users, id)

	return nil
}