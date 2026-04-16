package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ConnorBrightman/ezauth/internal/auth"
	"github.com/ConnorBrightman/ezauth/internal/httpx"
)

type RefreshRequest struct {
	Email        string `json:"email"`
	RefreshToken string `json:"refresh_token"`
}

func RefreshHandler(service *auth.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req RefreshRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body")
			return
		}

		accessToken, refreshToken, err := service.RefreshTokens(req.Email, req.RefreshToken)
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Could not refresh tokens")
			return
		}

		httpx.WriteJSON(w, http.StatusOK, map[string]string{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		})
	})
}
