package handlers

import (
	"encoding/json"
	"fmt"
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
			httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
			return
		}

		// Login and get both tokens
		accessToken, refreshToken, err := service.LoginWithTokens(auth.LoginInput{
			Email:    req.Email,
			Password: req.Password,
		})
		if err != nil {
			httpx.Error(w, http.StatusBadRequest, err.Error())
			return
		}

		// Respond with tokens
		httpx.JSON(w, http.StatusOK, map[string]string{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		})
		fmt.Println("Logged in")
	})
}
