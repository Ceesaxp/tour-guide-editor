// internal/services/auth_service.go
package services

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	users    map[string]string
	jwtKey   []byte
	tokenTTL time.Duration
	mu       sync.RWMutex
}

func NewAuthService(jwtKey string, tokenTTL time.Duration) *AuthService {
	// For demo purposes, hardcoded users
	users := map[string]string{
		"admin": "password123",
		"user":  "user12345",
	}

	return &AuthService{
		users:    users,
		jwtKey:   []byte(jwtKey),
		tokenTTL: tokenTTL,
	}
}

func (s *AuthService) Authenticate(username, password string) (string, error) {
	s.mu.RLock()
	storedPassword, exists := s.users[username]
	s.mu.RUnlock()

	if !exists || storedPassword != password {
		return "", fmt.Errorf("invalid credentials")
	}

	// Set expiration time explicitly
	expirationTime := time.Now().Add(s.tokenTTL * time.Minute)

	claims := jwt.MapClaims{
		"sub": username, // Use "sub" consistently
		"exp": time.Now().Add(s.tokenTTL).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	log.Printf("Creating token for user: %s", username)
	tokenString, err := token.SignedString(s.jwtKey)
	if err != nil {
		return "", fmt.Errorf("error creating token: %w", err)
	}

	log.Printf("Token created with expiration: %v", expirationTime)
	return tokenString, nil
}
