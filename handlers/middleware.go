package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userIDKey contextKey = "userID"

func (h *Handler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := h.parseJWT(r)
		if err != nil || !token.Valid {
			h.writeError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			h.writeError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		userID := int64(claims["sub"].(float64))
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) RequireAnyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			h.writeError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")

		if strings.HasPrefix(tokenStr, "pk_") {
			var userID int64
			err := h.db.QueryRow("SELECT user_id FROM api_keys WHERE key = ?", tokenStr).Scan(&userID)
			if err != nil {
				h.writeError(w, http.StatusUnauthorized, "Invalid API key")
				return
			}
			h.db.Exec("UPDATE api_keys SET last_used_at = CURRENT_TIMESTAMP WHERE key = ?", tokenStr)
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		token, err := h.parseJWT(r)
		if err != nil || !token.Valid {
			h.writeError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		userID := int64(claims["sub"].(float64))
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) parseJWT(r *http.Request) (*jwt.Token, error) {
	auth := r.Header.Get("Authorization")
	tokenStr := strings.TrimPrefix(auth, "Bearer ")
	if tokenStr == "" || tokenStr == auth {
		return nil, fmt.Errorf("no token")
	}

	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return h.jwtSecret, nil
	})
}
