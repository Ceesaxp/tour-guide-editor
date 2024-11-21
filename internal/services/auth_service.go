// internal/services/auth_service.go
package services

import (
	"fmt"
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

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "username": username,
        "exp":      time.Now().Add(s.tokenTTL).Unix(),
    })

    return token.SignedString(s.jwtKey)
}
