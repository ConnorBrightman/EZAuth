package handlers

import (
	"net/http"

	"github.com/ConnorBrightman/ezauth/internal/httpx"
	"github.com/ConnorBrightman/ezauth/internal/middleware"
)

// MeHandler returns the currently authenticated user's info
func MeHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := middleware.GetUserFromContext(r)
		if !ok {
			httpx.Error(w, http.StatusUnauthorized, "user not found in context")
			return
		}

		userID, _ := claims["user_id"].(string)
		email, _ := claims["email"].(string)

		httpx.JSON(w, http.StatusOK, map[string]string{
			"user_id": userID,
			"email":   email,
		})
	})
}
