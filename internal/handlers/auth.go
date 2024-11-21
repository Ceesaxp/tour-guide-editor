// internal/handlers/auth.go
package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"time"

	"github.com/ceesaxp/tour-guide-editor/internal/config"
	"github.com/ceesaxp/tour-guide-editor/internal/services"
	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct {
	secretKey   string
	tokenTTL    int
	templates   *template.Template
	authService *services.AuthService
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func NewAuthHandler(cfg config.Auth, templates *template.Template, authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		templates:   templates,
		authService: authService,
		secretKey:   cfg.SecretKey,
		tokenTTL:    cfg.TokenTTL,
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

func (h *AuthHandler) ServeLogin(w http.ResponseWriter, r *http.Request) {
	h.templates.ExecuteTemplate(w, "login", nil)
}

func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	token, err := h.authService.Authenticate(username, password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
	})

	// Redirect to editor
	w.Header().Set("HX-Redirect", "/")
}
