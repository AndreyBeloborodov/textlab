package algorithm

import (
	"bytes"
	"errors"
)

var (
	ErrEmptyTraining = errors.New("training data is empty")
	ErrNoMatch       = errors.New("no matching suffix found")
	ErrNoContinuation = errors.New("no continuation available")
)

const (
	predictionLength = 10
	hashBase         = uint64(257)
)

type HashAlgorithm struct{}

func NewHashAlgorithm() Algorithm {
	return &HashAlgorithm{}
}

// Train возвращает состояние обученной модели.
// Для HashAlgorithm состоянием является сам обучающий текст.
func (a *HashAlgorithm) Train(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, ErrEmptyTraining
	}

	state := make([]byte, len(data))
	copy(state, data)

	return state, nil
}

// Predict находит самое длинное совпадение суффикса input
// с обучающим текстом state и возвращает следующие байты.
func (a *HashAlgorithm) Predict(state []byte, input []byte) ([]byte, error) {
	if len(state) == 0 {
		return nil, ErrEmptyTraining
	}

	if len(input) == 0 {
		return nil, ErrNoMatch
	}

	length := longestSuffixLength(state, input)
	if length == 0 {
		return nil, ErrNoMatch
	}

	suffix := input[len(input)-length:]

	position := findLeftmost(state, suffix)
	if position == -1 {
		return nil, ErrNoMatch
	}

	start := position + length

	if start >= len(state) {
		return nil, ErrNoContinuation
	}

	end := start + predictionLength
	if end > len(state) {
		end = len(state)
	}

	result := make([]byte, end-start)
	copy(result, state[start:end])

	return result, nil
}

// longestSuffixLength возвращает максимальную длину суффикса input,
// который встречается в state.
//
// Используется бинарный поиск, поскольку если суффикс длины k
// встречается в state, то все его более короткие суффиксы
// также встречаются.
func longestSuffixLength(state []byte, input []byte) int {
	maxLength := len(input)
	if len(state) < maxLength {
		maxLength = len(state)
	}

	powers := make([]uint64, maxLength+1)
	powers[0] = 1

	for i := 1; i <= maxLength; i++ {
		powers[i] = powers[i-1] * hashBase
	}

	// Проверяем суффиксы input длиной от 1 до maxLength.
	// Для каждого размера проверка ищет вхождение в state.
	low := 1
	high := maxLength
	best := 0

	for low <= high {
		mid := low + (high-low)/2
		suffix := input[len(input)-mid:]

		if containsSubstring(state, suffix, powers[mid-1]) {
			best = mid
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	return best
}

// containsSubstring проверяет наличие pattern в text,
// используя rolling hash и дополнительное сравнение байтов.
func containsSubstring(text []byte, pattern []byte, highestPower uint64) bool {
	length := len(pattern)
	if length == 0 || length > len(text) {
		return false
	}

	targetHash := hashBytes(pattern)

	var windowHash uint64

	for i := 0; i < length; i++ {
		windowHash = windowHash*hashBase + uint64(text[i]) + 1
	}

	if windowHash == targetHash && bytes.Equal(text[:length], pattern) {
		return true
	}

	for i := length; i < len(text); i++ {
		outgoing := uint64(text[i-length]) + 1
		incoming := uint64(text[i]) + 1

		windowHash = (windowHash-outgoing*highestPower)*hashBase + incoming

		start := i - length + 1
		if windowHash == targetHash &&
			bytes.Equal(text[start:start+length], pattern) {
			return true
		}
	}

	return false
}

// findLeftmost возвращает индекс самого левого вхождения pattern в text.
// Возвращает -1, если совпадение не найдено.
func findLeftmost(text []byte, pattern []byte) int {
	if len(pattern) == 0 || len(pattern) > len(text) {
		return -1
	}

	return bytes.Index(text, pattern)
}

// hashBytes вычисляет полиномиальный хэш последовательности байтов.
func hashBytes(data []byte) uint64 {
	var hash uint64

	for _, b := range data {
		hash = hash*hashBase + uint64(b) + 1
	}

	return hash
}
