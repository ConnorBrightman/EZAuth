package api

import (
	"net/http"

	"ezauth/internal/api/handlers"
	"ezauth/internal/auth"
	"ezauth/internal/httpx"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()

	healthHandler := handlers.HealthHandler("1.0.0")
	healthHandler = httpx.AllowMethod(http.MethodGet, healthHandler)

	mux.Handle("/health", healthHandler)

	authService := auth.NewService()
	loginHandler := handlers.LoginHandler(authService)
	loginHandler = httpx.AllowMethod(http.MethodPost, loginHandler)

	mux.Handle("/auth/login", loginHandler)

	return mux
}
