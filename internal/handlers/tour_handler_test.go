// internal/handlers/tour_handler_test.go
package handlers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ceesaxp/tour-guide-editor/internal/models"
	"github.com/ceesaxp/tour-guide-editor/internal/services"
)

func TestTourHandler_Upload(t *testing.T) {
	tourService := services.NewTourService()
	handler := NewTourHandler(tourService, nil)

	tests := []struct {
		name           string
		fileContent    string
		expectedStatus int
	}{
		{
			name: "valid tour file",
			fileContent: `
id: "test_tour"
name: "Test Tour"
description: "A test tour"
start_date: "2024-01-01T00:00:00Z"
end_date: "2024-12-31T23:59:59Z"
version: "1.0"
hero_image: "http://example.com/image.jpg"
author:
  name: "Test Author"
  profile_link: "http://example.com/author"
price: 1000
nodes: []
edges: []
`,
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid tour file",
			fileContent: `
id: "test_tour"
invalid_yaml: [
`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create multipart form
			var b bytes.Buffer
			writer := multipart.NewWriter(&b)

			part, err := writer.CreateFormFile("tour_file", "test.yaml")
			if err != nil {
				t.Fatal(err)
			}
			part.Write([]byte(tt.fileContent))
			writer.Close()

			// Create request
			req := httptest.NewRequest("POST", "/tour/upload", &b)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			// Create response recorder
			rr := httptest.NewRecorder()

			// Handle request
			handler.Upload(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					rr.Code, tt.expectedStatus)
			}
		})
	}
}

func TestTourHandler_ValidateNode(t *testing.T) {
	tourService := services.NewTourService()
	handler := NewTourHandler(tourService, nil)

	tests := []struct {
		name           string
		node           models.Node
		expectedStatus int
		expectedHeader string // HX-Trigger header value
	}{
		{
			name: "valid node",
			node: models.Node{
				ID: 1,
				Location: models.Location{
					Lat: 45.0,
					Lon: 20.0,
				},
				ShortDesc: "Test Node",
				Narrative: "Test narrative",
				MediaFiles: []models.MediaFile{
					{
						Type:      "image",
						URI:       "http://example.com/image.jpg",
						SendDelay: 0,
					},
				},
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "showSuccess",
		},
		{
			name: "invalid node",
			node: models.Node{
				ID: 1,
				// Missing required fields
			},
			expectedStatus: http.StatusBadRequest,
			expectedHeader: "showError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			body, err := json.Marshal(tt.node)
			if err != nil {
				t.Fatal(err)
			}

			// Create request
			req := httptest.NewRequest("POST", "/tour/validate-node", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Handle request
			handler.ValidateNode(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					rr.Code, tt.expectedStatus)
			}

			// Check HTMX trigger header
			triggerHeader := rr.Header().Get("HX-Trigger")
			if triggerHeader != tt.expectedHeader {
				t.Errorf("handler returned wrong HX-Trigger header: got %v want %v",
					triggerHeader, tt.expectedHeader)
			}
		})
	}
}
