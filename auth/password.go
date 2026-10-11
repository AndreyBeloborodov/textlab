package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argon2Memory  uint32 = 64 * 1024
	argon2Time    uint32 = 3
	argon2Threads uint8  = 2
	argon2KeyLen  uint32 = 32
	argon2SaltLen        = 16
)

var ErrInvalidPassword = errors.New("invalid password")

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(password, encodedHash string) error
}

type Argon2PasswordHasher struct{}

func NewArgon2PasswordHasher() *Argon2PasswordHasher {
	return &Argon2PasswordHasher{}
}

func (h *Argon2PasswordHasher) Hash(password string) (string, error) {
	salt := make([]byte, argon2SaltLen)

	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		argon2Time,
		argon2Memory,
		argon2Threads,
		argon2KeyLen,
	)

	encoded := fmt.Sprintf(
		"argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argon2Memory,
		argon2Time,
		argon2Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

func (h *Argon2PasswordHasher) Compare(password, encodedHash string) error {
	salt, expectedHash, memory, iterations, threads, err :=
		parsePasswordHash(encodedHash)
	if err != nil {
		return ErrInvalidPassword
	}

	actualHash := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		threads,
		uint32(len(expectedHash)),
	)

	if subtle.ConstantTimeCompare(actualHash, expectedHash) != 1 {
		return ErrInvalidPassword
	}

	return nil
}

func parsePasswordHash(encoded string) (
	salt []byte,
	hash []byte,
	memory uint32,
	iterations uint32,
	threads uint8,
	err error,
) {
	parts := strings.Split(encoded, "$")

	if len(parts) != 5 || parts[0] != "argon2id" || parts[1] != "v=19" {
		err = errors.New("invalid password hash format")
		return
	}

	var params map[string]string
	params = make(map[string]string)

	for _, param := range strings.Split(parts[2], ",") {
		keyValue := strings.SplitN(param, "=", 2)
		if len(keyValue) != 2 {
			err = errors.New("invalid password hash parameters")
			return
		}
		params[keyValue[0]] = keyValue[1]
	}

	memoryValue, parseErr := strconv.ParseUint(params["m"], 10, 32)
	if parseErr != nil {
		err = parseErr
		return
	}

	iterationsValue, parseErr := strconv.ParseUint(params["t"], 10, 32)
	if parseErr != nil {
		err = parseErr
		return
	}

	threadsValue, parseErr := strconv.ParseUint(params["p"], 10, 8)
	if parseErr != nil {
		err = parseErr
		return
	}

	memory = uint32(memoryValue)
	iterations = uint32(iterationsValue)
	threads = uint8(threadsValue)

	// Не позволяем повреждённой записи вызвать чрезмерные затраты ресурсов.
	if memory < 8*uint32(threads) ||
		memory > 256*1024 ||
		iterations < 1 ||
		iterations > 10 ||
		threads < 1 ||
		threads > 8 {
		err = errors.New("invalid password hash parameters")
		return
	}

	salt, err = base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return
	}

	hash, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return
	}

	if len(salt) < 8 || len(salt) > 64 || len(hash) < 16 || len(hash) > 64 {
		err = errors.New("invalid password hash length")
		return
	}

	return
}
