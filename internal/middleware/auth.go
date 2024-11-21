// internal/middleware/auth.go
package middleware

import (
	"context"
	"fmt"
	"log"
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
				if cookie, err := r.Cookie("auth_token"); err == nil {
					tokenString = cookie.Value
				}
			}

			if tokenString == "" {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			// Parse token with explicit validation
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(secretKey), nil
			}, jwt.WithValidMethods([]string{"HS256"}))

			if err != nil {
				log.Printf("Token validation error: %v", err)
				// Token is expired, clear the cookie
				log.Printf("Force-clearing cookie")
				http.SetCookie(w, &http.Cookie{
					Name:     "auth_token",
					Value:    "",
					Path:     "/",
					MaxAge:   -1,
					HttpOnly: true,
				})

				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok || !token.Valid {
				log.Printf("Invalid token claims")
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			// Log claims for debugging
			log.Printf("Token claims: %+v", claims)

			ctx := context.WithValue(r.Context(), UserIDKey, claims["sub"])
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	if bearerToken != "" {
		if len(strings.Split(bearerToken, " ")) == 2 {
			return strings.Split(bearerToken, " ")[1]
		}
	}
	return ""
}
