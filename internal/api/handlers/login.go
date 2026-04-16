package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ConnorBrightman/ezauth/internal/auth"
	"github.com/ConnorBrightman/ezauth/internal/httpx"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func LoginHandler(service *auth.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest

		// Parse JSON request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body")
			return
		}

		if !httpx.ValidEmail(req.Email) {
			httpx.WriteError(w, http.StatusBadRequest, "INVALID_EMAIL", "invalid email address")
			return
		}

		// Login and get both tokens
		accessToken, refreshToken, err := service.LoginWithTokens(auth.LoginInput{
			Email:    req.Email,
			Password: req.Password,
		})
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid credentials")
			return
		}

		// Respond with tokens
		httpx.WriteJSON(w, http.StatusOK, map[string]string{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		})
	})
}
