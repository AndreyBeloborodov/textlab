package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/AndreyBeloborodov/textlab/users"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidRequest     = errors.New("invalid request")
)

type Service interface {
	Register(login, password string) (users.User, string, string, error)
	Login(login, password string) (users.User, string, string, error)
	Refresh(refreshToken string) (string, string, error)
	Logout(refreshToken string) error
}

type service struct {
	users             users.Service
	passwordHasher    PasswordHasher
	tokenManager      TokenManager
	refreshRepository RefreshRepository
	refreshTTL        time.Duration
}

func NewService(
	userService users.Service,
	passwordHasher PasswordHasher,
	tokenManager TokenManager,
	refreshRepository RefreshRepository,
	refreshTTL time.Duration,
) Service {
	return &service{
		users:             userService,
		passwordHasher:    passwordHasher,
		tokenManager:      tokenManager,
		refreshRepository: refreshRepository,
		refreshTTL:        refreshTTL,
	}
}

func (s *service) Register(
	login, password string,
) (users.User, string, string, error) {
	login = strings.TrimSpace(login)

	if login == "" || password == "" {
		return users.User{}, "", "", ErrInvalidRequest
	}

	passwordHash, err := s.passwordHasher.Hash(password)
	if err != nil {
		return users.User{}, "", "", err
	}

	user, err := s.users.Create(login, passwordHash)
	if err != nil {
		return users.User{}, "", "", err
	}

	accessToken, refreshToken, err := s.issueTokens(user.ID)
	if err != nil {
		return users.User{}, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *service) Login(
	login, password string,
) (users.User, string, string, error) {
	login = strings.TrimSpace(login)

	if login == "" || password == "" {
		return users.User{}, "", "", ErrInvalidCredentials
	}

	credentials, err := s.users.GetByLogin(login)
	if err != nil {
		// Не раскрываем, существует ли пользователь с таким логином.
		return users.User{}, "", "", ErrInvalidCredentials
	}

	if err = s.passwordHasher.Compare(
		password,
		credentials.PasswordHash,
	); err != nil {
		return users.User{}, "", "", ErrInvalidCredentials
	}

	accessToken, refreshToken, err := s.issueTokens(credentials.User.ID)
	if err != nil {
		return users.User{}, "", "", err
	}

	return credentials.User, accessToken, refreshToken, nil
}

func (s *service) Refresh(
	refreshToken string,
) (string, string, error) {
	if refreshToken == "" {
		return "", "", ErrRefreshTokenInvalid
	}

	userID, err := s.refreshRepository.Consume(
		hashRefreshToken(refreshToken),
		time.Now(),
	)
	if err != nil {
		return "", "", ErrRefreshTokenInvalid
	}

	// Не выдаём новые токены удалённому пользователю.
	if _, err = s.users.Get(userID); err != nil {
		return "", "", ErrRefreshTokenInvalid
	}

	accessToken, newRefreshToken, err := s.issueTokens(userID)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}

func (s *service) Logout(refreshToken string) error {
	if refreshToken == "" {
		return ErrRefreshTokenInvalid
	}

	_, err := s.refreshRepository.Consume(
		hashRefreshToken(refreshToken),
		time.Now(),
	)
	if err != nil {
		return ErrRefreshTokenInvalid
	}

	return nil
}

func (s *service) issueTokens(userID int) (string, string, error) {
	accessToken, err := s.tokenManager.CreateAccessToken(userID)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		return "", "", err
	}

	err = s.refreshRepository.Create(refreshRecord{
		TokenHash: hashRefreshToken(refreshToken),
		UserID:    userID,
		ExpiresAt: time.Now().Add(s.refreshTTL),
	})
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func generateRefreshToken() (string, error) {
	randomBytes := make([]byte, 32)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

func hashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))

	return base64.RawURLEncoding.EncodeToString(hash[:])
}
