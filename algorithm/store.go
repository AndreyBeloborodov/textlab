package algorithm

import (
    "errors"
    "strconv"
)

var ErrNotFound = errors.New("data not found")

type Store interface {
	Save(key string, data []byte) error
	Get(key string) ([]byte, error)
	Delete(key string) error
}

func ModelKey(modelID int) string {
	return "model:" + strconv.Itoa(modelID)
}
