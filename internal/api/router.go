package api

import (
	"net/http"

	"ezauth/internal/api/handlers"
	"ezauth/internal/auth"
	"ezauth/internal/httpx"
	"ezauth/internal/middleware"
)

func NewRouter(service *auth.Service, secret []byte) http.Handler {
	mux := http.NewServeMux()

	// Health
	healthHandler := handlers.HealthHandler("1.0.0")
	healthHandler = httpx.AllowMethod(http.MethodGet, healthHandler)
	mux.Handle("/health", healthHandler)

	// Register
	registerHandler := handlers.RegisterHandler(service)
	registerHandler = httpx.AllowMethod(http.MethodPost, registerHandler)
	mux.Handle("/auth/register", registerHandler)

	// Login
	loginHandler := handlers.LoginHandler(service)
	loginHandler = httpx.AllowMethod(http.MethodPost, loginHandler)
	mux.Handle("/auth/login", loginHandler)

	// Protected route
	meHandler := handlers.MeHandler()
	meHandler = middleware.JWTMiddleware(secret, meHandler)
	meHandler = httpx.AllowMethod(http.MethodGet, meHandler)
	mux.Handle("/auth/me", meHandler)

	return mux
}
