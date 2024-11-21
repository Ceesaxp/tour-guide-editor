// internal/handlers/auth.go
package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct {
	secretKey string
	tokenTTL  int
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func NewAuthHandler(cfg struct {
	SecretKey string `yaml:"secret_key"` // TODO: probably need to reference the actual config struct?
	TokenTTL  int    `yaml:"token_ttl"`
}) *AuthHandler {
	return &AuthHandler{
		secretKey: cfg.SecretKey,
		tokenTTL:  cfg.TokenTTL,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual user authentication
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": req.Username,
		"exp": time.Now().Add(time.Duration(h.tokenTTL) * time.Minute).Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.secretKey))
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(loginResponse{Token: tokenString})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Since we're using JWTs, logout is handled client-side
	w.WriteHeader(http.StatusOK)
}
