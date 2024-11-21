// internal/middleware/auth.go
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string
const UserIDKey contextKey = "userID"

func RequireAuth(secretKey string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            tokenString := extractToken(r)
            if tokenString == "" {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
                return []byte(secretKey), nil
            })

            if err != nil || !token.Valid {
                http.Error(w, "Invalid token", http.StatusUnauthorized)
                return
            }

            if claims, ok := token.Claims.(jwt.MapClaims); ok {
                ctx := context.WithValue(r.Context(), UserIDKey, claims["sub"])
                next.ServeHTTP(w, r.WithContext(ctx))
            } else {
                http.Error(w, "Invalid token claims", http.StatusUnauthorized)
            }
        })
    }
}

func extractToken(r *http.Request) string {
    bearerToken := r.Header.Get("Authorization")
    if len(strings.Split(bearerToken, " ")) == 2 {
        return strings.Split(bearerToken, " ")[1]
    }
    return ""
}
