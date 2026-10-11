package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrInvalidConfig = errors.New("invalid token configuration")
)

type TokenManager interface {
	CreateAccessToken(userID int) (string, error)
	ParseAccessToken(token string) (int, error)
}

type JWTManager struct {
	secret []byte
	ttl    time.Duration
}

type jwtHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

type jwtClaims struct {
	Subject string `json:"sub"`
	Expires int64  `json:"exp"`
	Issued  int64  `json:"iat"`
}

func NewJWTManager(secret string, ttl time.Duration) (*JWTManager, error) {
	if len(secret) < 32 || ttl <= 0 {
		return nil, ErrInvalidConfig
	}

	return &JWTManager{
		secret: []byte(secret),
		ttl:    ttl,
	}, nil
}

func (m *JWTManager) CreateAccessToken(userID int) (string, error) {
	if userID <= 0 {
		return "", ErrInvalidToken
	}

	now := time.Now()

	header, err := json.Marshal(jwtHeader{
		Algorithm: "HS256",
		Type:      "JWT",
	})
	if err != nil {
		return "", err
	}

	claims, err := json.Marshal(jwtClaims{
		Subject: strconv.Itoa(userID),
		Expires: now.Add(m.ttl).Unix(),
		Issued:  now.Unix(),
	})
	if err != nil {
		return "", err
	}

	encodedHeader := base64.RawURLEncoding.EncodeToString(header)
	encodedClaims := base64.RawURLEncoding.EncodeToString(claims)

	unsigned := encodedHeader + "." + encodedClaims
	signature := m.sign(unsigned)

	return unsigned + "." + signature, nil
}

func (m *JWTManager) ParseAccessToken(token string) (int, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return 0, ErrInvalidToken
	}

	unsigned := parts[0] + "." + parts[1]
	expectedSignature := m.sign(unsigned)

	if !hmac.Equal(
		[]byte(expectedSignature),
		[]byte(parts[2]),
	) {
		return 0, ErrInvalidToken
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return 0, ErrInvalidToken
	}

	var header jwtHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return 0, ErrInvalidToken
	}

	if header.Algorithm != "HS256" || header.Type != "JWT" {
		return 0, ErrInvalidToken
	}

	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, ErrInvalidToken
	}

	var claims jwtClaims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return 0, ErrInvalidToken
	}

	userID, err := strconv.Atoi(claims.Subject)
	if err != nil || userID <= 0 {
		return 0, ErrInvalidToken
	}

	now := time.Now().Unix()

	if claims.Expires <= now || claims.Issued > now {
		return 0, ErrInvalidToken
	}

	return userID, nil
}

func (m *JWTManager) sign(value string) string {
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(value))

	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (m *JWTManager) String() string {
	return fmt.Sprintf("JWTManager{ttl:%s}", m.ttl)
}
