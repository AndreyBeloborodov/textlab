package models

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/AndreyBeloborodov/textlab/auth"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

type createModelRequest struct {
	Name string `json:"name"`
}

type updateModelRequest struct {
	Name string `json:"name"`
}

type modelResponse struct {
	ID      int    `json:"id"`
	OwnerID int    `json:"owner_id"`
	Name    string `json:"name"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeModelError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var request createModelRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeModelError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	model, err := h.service.Create(userID, request.Name)
	if err != nil {
		handleModelError(w, err)
		return
	}

	writeModelJSON(w, http.StatusCreated, modelResponseFromModel(model))
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeModelError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, ok := parseModelID(r)
	if !ok {
		writeModelError(w, http.StatusBadRequest, "invalid model id")
		return
	}

	model, err := h.service.Get(userID, id)
	if err != nil {
		handleModelError(w, err)
		return
	}

	writeModelJSON(w, http.StatusOK, modelResponseFromModel(model))
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeModelError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	models, err := h.service.GetAll(userID)
	if err != nil {
		handleModelError(w, err)
		return
	}

	result := make([]modelResponse, 0, len(models))

	for _, model := range models {
		result = append(result, modelResponseFromModel(model))
	}

	writeModelJSON(w, http.StatusOK, result)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeModelError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, ok := parseModelID(r)
	if !ok {
		writeModelError(w, http.StatusBadRequest, "invalid model id")
		return
	}

	var request updateModelRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeModelError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	model, err := h.service.Update(userID, id, request.Name)
	if err != nil {
		handleModelError(w, err)
		return
	}

	writeModelJSON(w, http.StatusOK, modelResponseFromModel(model))
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeModelError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, ok := parseModelID(r)
	if !ok {
		writeModelError(w, http.StatusBadRequest, "invalid model id")
		return
	}

	if err := h.service.Delete(userID, id); err != nil {
		handleModelError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseModelID(r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))

	return id, err == nil && id > 0
}

func modelResponseFromModel(model Model) modelResponse {
	return modelResponse{
		ID:      model.ID,
		OwnerID: model.OwnerID,
		Name:    model.Name,
	}
}

func handleModelError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeModelError(w, http.StatusNotFound, "model not found")

	case errors.Is(err, ErrInvalidName):
		writeModelError(w, http.StatusBadRequest, "invalid model name")

	case errors.Is(err, ErrInvalidOwner):
		writeModelError(w, http.StatusBadRequest, "invalid owner")

	default:
		writeModelError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeModelJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(value)
}

func writeModelError(w http.ResponseWriter, status int, message string) {
	writeModelJSON(w, status, map[string]string{
		"error": message,
	})
}
