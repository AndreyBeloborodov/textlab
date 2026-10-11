package users

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

type updateUserRequest struct {
	Login string `json:"login"`
}

type userResponse struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	user, err := h.service.Get(id)
	if err != nil {
		handleUserError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, userResponseFromUser(user))
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.GetAll()
	if err != nil {
		handleUserError(w, err)
		return
	}

	result := make([]userResponse, 0, len(users))

	for _, user := range users {
		result = append(result, userResponseFromUser(user))
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var request updateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.service.Update(id, request.Login)
	if err != nil {
		handleUserError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, userResponseFromUser(user))
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := h.service.Delete(id); err != nil {
		handleUserError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseID(r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	return id, err == nil && id > 0
}

func userResponseFromUser(user User) userResponse {
	return userResponse{
		ID:    user.ID,
		Login: user.Login,
	}
}

func handleUserError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "user not found")

	case errors.Is(err, ErrLoginTaken):
		writeError(w, http.StatusConflict, "login already taken")

	case errors.Is(err, ErrInvalidUser):
		writeError(w, http.StatusBadRequest, "invalid user")

	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}
