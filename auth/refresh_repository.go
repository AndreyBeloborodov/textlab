package auth

import (
	"errors"
	"time"
)

var ErrRefreshTokenInvalid = errors.New("invalid refresh token")

type refreshRecord struct {
	TokenHash string
	UserID    int
	ExpiresAt time.Time
}

type RefreshRepository interface {
	Create(token refreshRecord) error
	Consume(tokenHash string, now time.Time) (int, error)
}
