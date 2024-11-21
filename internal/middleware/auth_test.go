// internal/middleware/auth_test.go

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestRequireAuth(t *testing.T) {
    secretKey := "test-secret"

    // Create a test handler
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })

    // Create middleware
    authMiddleware := RequireAuth(secretKey)(handler)

    tests := []struct {
        name           string
        token          string
        expectedStatus int
    }{
        {
            name:           "no token",
            token:          "",
            expectedStatus: http.StatusUnauthorized,
        },
        {
            name:           "invalid token",
            token:          "Bearer invalid-token",
            expectedStatus: http.StatusUnauthorized,
        },
        {
            name:           "valid token",
            token:          createValidToken(secretKey),
            expectedStatus: http.StatusOK,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("GET", "/", nil)
            if tt.token != "" {
                req.Header.Set("Authorization", tt.token)
            }

            rr := httptest.NewRecorder()
            authMiddleware.ServeHTTP(rr, req)

            if rr.Code != tt.expectedStatus {
                t.Errorf("handler returned wrong status code: got %v want %v",
                    rr.Code, tt.expectedStatus)
            }
        })
    }
}

func createValidToken(secretKey string) string {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "sub": "testuser",
        "exp": time.Now().Add(time.Hour).Unix(),
    })

    tokenString, _ := token.SignedString([]byte(secretKey))
    return "Bearer " + tokenString
}
