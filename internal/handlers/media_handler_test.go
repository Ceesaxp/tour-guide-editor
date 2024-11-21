// internal/handlers/media_handler_test.go
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ceesaxp/tour-guide-editor/internal/mocks"
	"github.com/ceesaxp/tour-guide-editor/internal/services"
)

func TestMediaHandler_Upload(t *testing.T) {
	// Create temporary test directory
    tempDir, err := os.MkdirTemp("", "media-test-*")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(tempDir)

    config := services.MediaConfig{
        MaxFileSize:    1024 * 1024,
        AllowedFormats: []string{"image/", "audio/", "video/"},
        ImageMaxWidth:  800,
        ImageMaxHeight: 600,
        S3Bucket:      "test-bucket",
    }

    mockS3 := &mocks.MockS3Client{
        PutObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
            return &s3.PutObjectOutput{}, nil
        },
    }

    mediaService := services.NewMediaService(config, mockS3)
    handler := NewMediaHandler(mediaService)

    tests := []struct {
        name         string
        createFile   func(t *testing.T) ([]byte, string)
        expectedCode int
    }{
        {
            name: "valid image upload",
            createFile: func(t *testing.T) ([]byte, string) {
                img := image.NewRGBA(image.Rect(0, 0, 100, 100))
                buf := new(bytes.Buffer)
                if err := jpeg.Encode(buf, img, nil); err != nil {
                    t.Fatal(err)
                }
                return buf.Bytes(), "test.jpg"
            },
            expectedCode: http.StatusOK,
        },
        {
            name: "invalid file type",
            createFile: func(t *testing.T) ([]byte, string) {
                return []byte("invalid file content"), "test.txt"
            },
            expectedCode: http.StatusBadRequest,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            content, filename := tt.createFile(t)

            // Create multipart form
            body := &bytes.Buffer{}
            writer := multipart.NewWriter(body)

            part, err := writer.CreateFormFile("file", filename)
            if err != nil {
                t.Fatal(err)
            }
            part.Write(content)
            writer.Close()

            // Create request
            req := httptest.NewRequest("POST", "/media/upload", body)
            req.Header.Set("Content-Type", writer.FormDataContentType())

            // Create response recorder
            rr := httptest.NewRecorder()

            // Handle request
            handler.Upload(rr, req)

            if rr.Code != tt.expectedCode {
                t.Errorf("handler returned wrong status code: got %v want %v",
                    rr.Code, tt.expectedCode)
            }
        })
    }
}

func TestMediaHandler_ValidateURL(t *testing.T) {
	mockS3 := &mocks.MockS3Client{
		PutObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			return &s3.PutObjectOutput{}, nil
		},
	}
	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always return success with valid image type
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Length", "1000")
		// Create and write a valid JPEG
		img := image.NewRGBA(image.Rect(0, 0, 100, 100))
		jpeg.Encode(w, img, nil)
	}))
	defer ts.Close()

	config := services.MediaConfig{
		MaxFileSize:    1024 * 1024,
		AllowedFormats: []string{"image/", "audio/", "video/"},
	}

	mediaService := services.NewMediaService(config, mockS3)
	handler := NewMediaHandler(mediaService)

	tests := []struct {
		name         string
		request      validateURLRequest
		expectedCode int
	}{
		{
			name: "valid URL",
			request: validateURLRequest{
				URL: ts.URL + "/test.jpg",
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "invalid URL",
			request: validateURLRequest{
				URL: "not-a-url",
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
	// Create a new 100x100 image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// Fill it with a solid color
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	// Encode to JPEG
	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, img, nil); err != nil {
		t.Fatalf("Failed to encode test image: %v", err)
	}

	return buf.Bytes()
}
