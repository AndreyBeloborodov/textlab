package algorithm

import "sync"

type MemoryStore struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[string][]byte),
	}
}

func (s *MemoryStore) Save(key string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = append([]byte(nil), data...)

	return nil
}

func (s *MemoryStore) Get(key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, ok := s.data[key]
	if !ok {
		return nil, ErrNotFound
	}

	return append([]byte(nil), data...), nil
}

func (s *MemoryStore) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data[key]; !ok {
		return ErrNotFound
	}

	delete(s.data, key)

	return nil
}
