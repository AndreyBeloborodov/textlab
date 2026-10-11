package prediction

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/AndreyBeloborodov/textlab/auth"
	"github.com/AndreyBeloborodov/textlab/algorithm"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Predict(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	modelID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || modelID <= 0 {
		http.Error(w, "invalid model id", http.StatusBadRequest)
		return
	}

	var request struct {
		Text string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	prediction, err := h.service.Predict(ownerID, modelID, request.Text)
	if err != nil {
		switch {
		case errors.Is(err, ErrModelNotFound):
			http.Error(w, "model not found", http.StatusNotFound)

		case errors.Is(err, ErrModelNotTrained):
			http.Error(w, "model has not been trained", http.StatusConflict)

		case errors.Is(err, ErrEmptyInput):
			http.Error(w, err.Error(), http.StatusBadRequest)

		case errors.Is(err, algorithm.ErrNoMatch):
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)

		case errors.Is(err, algorithm.ErrNoContinuation):
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)

		case errors.Is(err, ErrAlgorithmUnavailable):
			http.Error(w, "algorithm unavailable", http.StatusInternalServerError)

		default:
			http.Error(w, "failed to predict text", http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusOK, struct {
		Prediction string `json:"prediction"`
	}{
		Prediction: prediction,
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(value)
}
