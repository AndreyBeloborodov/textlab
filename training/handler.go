package training

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
    return &Handler{service: service}
}

func (h *Handler) Start(w http.ResponseWriter, r *http.Request) {
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

    if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, "invalid JSON body", http.StatusBadRequest)
        return
    }

    t, err := h.service.Start(ownerID, modelID, request.Text)
    if err != nil {
        switch {
        case errors.Is(err, ErrModelNotFound):
            http.Error(w, "model not found", http.StatusNotFound)
        case errors.Is(err, ErrEmptyText):
            http.Error(w, err.Error(), http.StatusBadRequest)
        case errors.Is(err, ErrTrainingInProgress):
            http.Error(w, err.Error(), http.StatusConflict)
        default:
            http.Error(w, "failed to start training", http.StatusInternalServerError)
        }
        return
    }

    writeJSON(w, http.StatusAccepted, t)
}

func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
    ownerID, ok := auth.UserIDFromContext(r.Context())
    if !ok {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    trainingID, err := strconv.Atoi(r.PathValue("id"))
    if err != nil || trainingID <= 0 {
        http.Error(w, "invalid training id", http.StatusBadRequest)
        return
    }

    t, err := h.service.GetStatus(ownerID, trainingID)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            http.Error(w, "training not found", http.StatusNotFound)
            return
        }

        http.Error(w, "failed to get training status", http.StatusInternalServerError)
        return
    }

    writeJSON(w, http.StatusOK, t)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)

    _ = json.NewEncoder(w).Encode(value)
}
