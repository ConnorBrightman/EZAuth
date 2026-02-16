package handlers

import (
	"encoding/json"
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
			httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
			return
		}

		if err := service.Register(req.Email, req.Password); err != nil {
			httpx.Error(w, http.StatusBadRequest, err.Error())
			return
		}

		httpx.JSON(w, http.StatusCreated, map[string]string{
			"message": "user registered successfully",
		})
	})
}
