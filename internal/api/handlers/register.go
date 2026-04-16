package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ConnorBrightman/ezauth/internal/auth"
	"github.com/ConnorBrightman/ezauth/internal/httpx"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func RegisterHandler(service *auth.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body")
			return
		}

		if !httpx.ValidEmail(req.Email) {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_EMAIL", "invalid email address")
			return
		}

		if ok, reason := httpx.ValidPassword(req.Password); !ok {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_PASSWORD", reason)
			return
		}

		if err := service.Register(req.Email, req.Password); err != nil {
			if errors.Is(err, auth.ErrUserExists) {
				httpx.WriteError(w, http.StatusConflict, "EMAIL_ALREADY_EXISTS", "a user with that email already exists")
				return
			}
			httpx.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}

		httpx.WriteJSON(w, http.StatusCreated, map[string]string{
			"message": "user registered successfully",
		})
	})
}
