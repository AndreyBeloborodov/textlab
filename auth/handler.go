package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/AndreyBeloborodov/textlab/users"
)

const refreshCookieName = "refresh_token"

type Handler struct {
	service    Service
	refreshTTL time.Duration
	secure     bool
}

func NewHandler(
	service Service,
	refreshTTL time.Duration,
	secure bool,
) *Handler {
	return &Handler{
		service:    service,
		refreshTTL: refreshTTL,
		secure:     secure,
	}
}

type credentialsRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type refreshResponse struct {
	AccessToken string `json:"access_token"`
}

type authResponse struct {
	User        authUserResponse `json:"user"`
	AccessToken string           `json:"access_token"`
}

type authUserResponse struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var request credentialsRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAuthError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, accessToken, refreshToken, err :=
		h.service.Register(request.Login, request.Password)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.setRefreshCookie(w, refreshToken)

	writeAuthJSON(w, http.StatusCreated, authResponse{
		User: authUserResponse{
			ID:    user.ID,
			Login: user.Login,
		},
		AccessToken: accessToken,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var request credentialsRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAuthError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, accessToken, refreshToken, err :=
		h.service.Login(request.Login, request.Password)
	if err != nil {
		h.handleError(w, err)
		return
	}

	h.setRefreshCookie(w, refreshToken)

	writeAuthJSON(w, http.StatusOK, authResponse{
		User: authUserResponse{
			ID:    user.ID,
			Login: user.Login,
		},
		AccessToken: accessToken,
	})
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil {
		writeAuthError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	accessToken, refreshToken, err := h.service.Refresh(cookie.Value)
	if err != nil {
		h.clearRefreshCookie(w)
		h.handleError(w, err)
		return
	}

	h.setRefreshCookie(w, refreshToken)

	writeAuthJSON(w, http.StatusOK, refreshResponse{
		AccessToken: accessToken,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshCookieName)
    if err != nil {
        h.clearRefreshCookie(w)
	    w.WriteHeader(http.StatusNoContent)
        return
    }

    err = h.service.Logout(cookie.Value)
    h.clearRefreshCookie(w)

    if err != nil && !errors.Is(err, ErrRefreshTokenInvalid) {
        writeAuthError(
            w,
            http.StatusInternalServerError,
            "internal server error",
        )
        return
    }
	
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) setRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Path:     "/api/auth",
		MaxAge:   int(h.refreshTTL.Seconds()),
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/api/auth",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidRequest):
		writeAuthError(w, http.StatusBadRequest, "invalid request")

	case errors.Is(err, ErrInvalidCredentials):
		writeAuthError(w, http.StatusUnauthorized, "invalid credentials")

	case errors.Is(err, ErrRefreshTokenInvalid):
		writeAuthError(w, http.StatusUnauthorized, "unauthorized")

	case errors.Is(err, users.ErrLoginTaken):
		writeAuthError(w, http.StatusConflict, "login already taken")

	case errors.Is(err, users.ErrInvalidUser):
		writeAuthError(w, http.StatusBadRequest, "invalid user")

	default:
		writeAuthError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeAuthJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(value)
}

func writeAuthError(w http.ResponseWriter, status int, message string) {
	writeAuthJSON(w, status, map[string]string{
		"error": message,
	})
}
