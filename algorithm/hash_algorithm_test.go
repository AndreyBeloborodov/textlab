package algorithm

import (
	"bytes"
	"errors"
	"testing"
)

func TestHashAlgorithmTrain(t *testing.T) {
	algo := NewHashAlgorithm()

	input := []byte("hello world")

	state, err := algo.Train(input)
	if err != nil {
		t.Fatalf("Train() error = %v", err)
	}

	if !bytes.Equal(state, input) {
		t.Fatalf("Train() = %q, want %q", state, input)
	}

	// Состояние не должно зависеть от последующих изменений
	// исходного массива байтов.
	input[0] = 'X'

	if !bytes.Equal(state, []byte("hello world")) {
		t.Fatalf("state unexpectedly changed: %q", state)
	}
}

func TestHashAlgorithmTrainEmpty(t *testing.T) {
	algo := NewHashAlgorithm()

	_, err := algo.Train(nil)
	if !errors.Is(err, ErrEmptyTraining) {
		t.Fatalf("Train() error = %v, want %v", err, ErrEmptyTraining)
	}
}

func TestHashAlgorithmPredictUsesLongestSuffix(t *testing.T) {
	algo := NewHashAlgorithm()

	state := []byte("abcXbcYcZ")
	input := []byte("qqabc")

	got, err := algo.Predict(state, input)
	if err != nil {
		t.Fatalf("Predict() error = %v", err)
	}

	want := []byte("XbcYcZ")
	if !bytes.Equal(got, want) {
		t.Fatalf("Predict() = %q, want %q", got, want)
	}
}

func TestHashAlgorithmPredictUsesLeftmostOccurrence(t *testing.T) {
	algo := NewHashAlgorithm()

	state := []byte("abc111abc222")
	input := []byte("xyzabc")

	got, err := algo.Predict(state, input)
	if err != nil {
		t.Fatalf("Predict() error = %v", err)
	}

	want := []byte("111abc222")
	if !bytes.Equal(got, want) {
		t.Fatalf("Predict() = %q, want %q", got, want)
	}
}

func TestHashAlgorithmPredictNoMatch(t *testing.T) {
	algo := NewHashAlgorithm()

	state := []byte("abcdefgh")
	input := []byte("xyz")

	_, err := algo.Predict(state, input)
	if !errors.Is(err, ErrNoMatch) {
		t.Fatalf("Predict() error = %v, want %v", err, ErrNoMatch)
	}
}

func TestHashAlgorithmPredictNoContinuation(t *testing.T) {
	algo := NewHashAlgorithm()

	state := []byte("abc")
	input := []byte("abc")

	_, err := algo.Predict(state, input)
	if !errors.Is(err, ErrNoContinuation) {
		t.Fatalf("Predict() error = %v, want %v", err, ErrNoContinuation)
	}
}

func TestHashAlgorithmPredictShortContinuation(t *testing.T) {
	algo := NewHashAlgorithm()

	state := []byte("abc123")
	input := []byte("xyzabc")

	got, err := algo.Predict(state, input)
	if err != nil {
		t.Fatalf("Predict() error = %v", err)
	}

	want := []byte("123")
	if !bytes.Equal(got, want) {
		t.Fatalf("Predict() = %q, want %q", got, want)
	}
}

func TestMemoryStore(t *testing.T) {
	store := NewMemoryStore()

	key := "model-1"
	data := []byte("trained state")

	if err := store.Save(key, data); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Изменение исходного среза не должно менять данные в хранилище.
	data[0] = 'X'

	got, err := store.Get(key)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	want := []byte("trained state")
	if !bytes.Equal(got, want) {
		t.Fatalf("Get() = %q, want %q", got, want)
	}

	// Изменение полученного среза тоже не должно менять хранилище.
	got[0] = 'Y'

	got, err = store.Get(key)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if !bytes.Equal(got, want) {
		t.Fatalf("Get() after mutation = %q, want %q", got, want)
	}

	if err := store.Delete(key); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = store.Get(key)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get() error = %v, want %v", err, ErrNotFound)
	}
}
