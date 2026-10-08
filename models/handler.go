package models

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
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

type modelResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var request createModelRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	model, err := h.service.Create(request.Name)
	if err != nil {
		http.Error(w, "failed to create model", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, modelResponseFromModel(model))
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	models, err := h.service.GetAll()
	if err != nil {
		http.Error(w, "failed to get models", http.StatusInternalServerError)
		return
	}

    modelResponses := make([]modelResponse, len(models))
    for i, model := range models {
        modelResponses[i] = modelResponseFromModel(model)
    }

	writeJSON(w, http.StatusOK, modelResponses)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		http.Error(w, "invalid model id", http.StatusBadRequest)
		return
	}

	model, err := h.service.Get(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "model not found", http.StatusNotFound)
			return
		}

		http.Error(w, "failed to get model", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, modelResponseFromModel(model))
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		http.Error(w, "invalid model id", http.StatusBadRequest)
		return
	}

	var request createModelRequest

	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	model, err := h.service.Update(id, request.Name)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "model not found", http.StatusNotFound)
			return
		}

		http.Error(w, "failed to update model", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, modelResponseFromModel(model))
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getID(r)
	if err != nil {
		http.Error(w, "invalid model id", http.StatusBadRequest)
		return
	}

	err = h.service.Delete(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "model not found", http.StatusNotFound)
			return
		}

		http.Error(w, "failed to delete model", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getID(r *http.Request) (int, error) {
	return strconv.Atoi(r.PathValue("id"))
}

func modelResponseFromModel(model Model) modelResponse {
	return modelResponse{
		ID:   model.ID,
		Name: model.Name,
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(value)
}