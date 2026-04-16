package handlers

import (
	"net/http"

	"github.com/ConnorBrightman/ezauth/internal/auth"
	"github.com/ConnorBrightman/ezauth/internal/httpx"
	"github.com/ConnorBrightman/ezauth/internal/middleware"
)

func LogoutHandler(service *auth.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := middleware.GetUserFromContext(r)
		if !ok {
			httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not found in context")
			return
		}

		email, _ := claims["email"].(string)

		if err := service.Logout(email); err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not log out")
			return
		}

		httpx.WriteJSON(w, http.StatusOK, map[string]string{
			"message": "logged out successfully",
		})
	})
}
