package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ConnorBrightman/ezauth/internal/httpx"
	"github.com/golang-jwt/jwt/v5"
)

// key type for storing user in context
type contextKey string

const UserContextKey contextKey = "user"

// JWTMiddleware validates JWT tokens and stores user info in request context
func JWTMiddleware(secretKey []byte, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid authorization header")
			return
		}

		tokenStr := parts[1]

		// Parse and verify token
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, http.ErrAbortHandler
			}
			return secretKey, nil
		})

		if err != nil || !token.Valid {
			httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
			return
		}

		// Store claims in context
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token claims")
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext extracts the JWT claims from request context
func GetUserFromContext(r *http.Request) (jwt.MapClaims, bool) {
	claims, ok := r.Context().Value(UserContextKey).(jwt.MapClaims)
	return claims, ok
}
