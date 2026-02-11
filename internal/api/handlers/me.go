package handlers

import (
	"ezauth/internal/httpx"
	"ezauth/internal/middleware"
	"net/http"
)

func MeHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims, ok := middleware.GetUserFromContext(r)
		if !ok {
			httpx.Error(w, http.StatusUnauthorized, "invalid token context")
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
