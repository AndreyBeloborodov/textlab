package auth

import (
	"sync"
	"time"
)

type MemoryRefreshRepository struct {
	mu     sync.Mutex
	tokens map[string]refreshRecord
}

func NewMemoryRefreshRepository() *MemoryRefreshRepository {
	return &MemoryRefreshRepository{
		tokens: make(map[string]refreshRecord),
	}
}

func (r *MemoryRefreshRepository) Create(token refreshRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tokens[token.TokenHash] = token

	return nil
}

func (r *MemoryRefreshRepository) Consume(
	tokenHash string,
	now time.Time,
) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	token, ok := r.tokens[tokenHash]
	if !ok {
		return 0, ErrRefreshTokenInvalid
	}

	// Удаляем токен независимо от того, истёк он или нет.
	delete(r.tokens, tokenHash)

	if !now.Before(token.ExpiresAt) {
		return 0, ErrRefreshTokenInvalid
	}

	return token.UserID, nil
}
