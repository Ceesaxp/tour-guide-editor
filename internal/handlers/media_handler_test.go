// internal/handlers/media_handler_test.go
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ceesaxp/tour-guide-editor/internal/services"
	"github.com/h2non/bimg"
)

func TestMediaHandler_Upload(t *testing.T) {
	// Create mock media service
	config := services.MediaConfig{
		MaxFileSize:    1024 * 1024,
		AllowedFormats: []string{"image/", "audio/", "video/"},
		ImageMaxWidth:  800,
		ImageMaxHeight: 600,
		S3Bucket:       "test-bucket",
	}

	mockS3 := &services.MockS3Client{
		PutObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			return &s3.PutObjectOutput{}, nil
		},
	}

	mediaService := services.NewMediaService(config, mockS3)
	handler := NewMediaHandler(mediaService)

	tests := []struct {
		name         string
		fileContent  []byte
		filename     string
		contentType  string
		expectedCode int
	}{
		{
			name:         "valid image upload",
			fileContent:  createTestImage(t),
			filename:     "test.jpg",
			contentType:  "image/jpeg",
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid file type",
			fileContent:  []byte("invalid file content"),
			filename:     "test.txt",
			contentType:  "text/plain",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create multipart form
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			part, err := writer.CreateFormFile("file", tt.filename)
			if err != nil {
				t.Fatal(err)
			}
			part.Write(tt.fileContent)
			writer.Close()

			// Create request
			req := httptest.NewRequest("POST", "/media/upload", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			// Create response recorder
			rr := httptest.NewRecorder()

			// Handle request
			handler.Upload(rr, req)

			// Check status code
			if rr.Code != tt.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					rr.Code, tt.expectedCode)
			}

			if tt.expectedCode == http.StatusOK {
				var response services.ProcessedMedia
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if response.URL == "" {
					t.Error("Expected URL in response, got empty string")
				}
			}
		})
	}
}

func TestMediaHandler_ValidateURL(t *testing.T) {
	config := services.MediaConfig{
		MaxFileSize:    1024 * 1024,
		AllowedFormats: []string{"image/", "audio/", "video/"},
	}

	mediaService := services.NewMediaService(config, nil)
	handler := NewMediaHandler(mediaService)

	tests := []struct {
		name         string
		request      validateURLRequest
		expectedCode int
	}{
		{
			name: "valid URL",
			request: validateURLRequest{
				URL: "http://example.com/valid.jpg",
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "invalid URL",
			request: validateURLRequest{
				URL: "invalid-url",
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			body, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatal(err)
			}

			// Create request
			req := httptest.NewRequest("POST", "/media/validate-url", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Handle request
			handler.ValidateURL(rr, req)

			// Check status code
			if rr.Code != tt.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					rr.Code, tt.expectedCode)
			}
		})
	}
}

// Helper functions for testing
func createTestImage(t *testing.T) []byte {
	// Create a small test JPEG image using bimg
	img := bimg.NewImage(make([]byte, 100*100*3))
	options := bimg.Options{
		Width:  100,
		Height: 100,
		Type:   bimg.JPEG,
	}

	processed, err := img.Process(options)
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	return processed
}
