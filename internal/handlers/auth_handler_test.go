// internal/handlers/auth_test.go
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ceesaxp/tour-guide-editor/internal/config"
)

func TestAuthHandler_Login(t *testing.T) {
	handler := NewAuthHandler(config.Auth{
		SecretKey: "test-secret",
		TokenTTL:  60,
	}, nil, nil)

	tests := []struct {
		name           string
		request        loginRequest
		expectedStatus int
	}{
		{
			name: "empty credentials",
			request: loginRequest{
				Username: "",
				Password: "",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "valid credentials",
			request: loginRequest{
				Username: "testuser",
				Password: "testpass",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.Login(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					rr.Code, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusOK {
				var response loginResponse
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}
				if response.Token == "" {
					t.Error("Expected token in response, got empty string")
				}
			}
		})
	}
}
