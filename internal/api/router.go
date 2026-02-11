package api

import (
	"net/http"

	"ezauth/internal/api/handlers"
	"ezauth/internal/auth"
	"ezauth/internal/httpx"
)

// NewRouter creates the HTTP router using the provided auth service
func NewRouter(service *auth.Service) http.Handler {
	mux := http.NewServeMux()

	// Health check
	healthHandler := handlers.HealthHandler("1.0.0")
	healthHandler = httpx.AllowMethod(http.MethodGet, healthHandler)
	mux.Handle("/health", healthHandler)

	// Login
	loginHandler := handlers.LoginHandler(service)
	loginHandler = httpx.AllowMethod(http.MethodPost, loginHandler)
	mux.Handle("/auth/login", loginHandler)

	// Register
	registerHandler := handlers.RegisterHandler(service)
	registerHandler = httpx.AllowMethod(http.MethodPost, registerHandler)
	mux.Handle("/auth/register", registerHandler)

	return mux
}
