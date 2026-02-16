package handlers

import (
	"net/http"

	"github.com/ConnorBrightman/ezauth/internal/httpx"
)

type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

func HealthHandler(version string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpx.JSON(w, http.StatusOK, HealthResponse{
			Status:  "healthy",
			Version: version,
		})
	})
}
