package handlers

import (
	"net/http"

	"ezauth/internal/httpx"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func LoginHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest

		if err := httpx.DecodeJSON(r, &req); err != nil {
			httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
			return
		}

		if !httpx.Required(req.Email) || !httpx.Required(req.Password) {
			httpx.Error(w, http.StatusBadRequest, "email and password are required")
			return
		}

		// Placeholder logic
		httpx.JSON(w, http.StatusOK, map[string]string{
			"message": "login input valid",
		})
	})
}
