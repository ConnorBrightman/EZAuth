package httpx

import (
	"encoding/json"
	"net/http"
)

func Error(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := Response{
		Success: false,
		Error:   message,
	}

	_ = json.NewEncoder(w).Encode(resp)
}
