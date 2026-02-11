package handlers

import (
	"encoding/json"
	"net/http"

	"ezauth/internal/auth"
	"ezauth/internal/httpx"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func LoginHandler(service *auth.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
			return
		}

		user, err := service.Login(auth.LoginInput{
			Email:    req.Email,
			Password: req.Password,
		})
		if err != nil {
			httpx.Error(w, http.StatusBadRequest, err.Error())
			return
		}

		// Generate JWT
		token, err := service.GenerateToken(user)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "failed to generate token")
			return
		}

		// Respond with token
		httpx.JSON(w, http.StatusOK, map[string]string{
			"token": token,
		})
	})
}
