// internal/services/media_service_test.go
package services

import (
	"bytes"
	"context"
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
	"github.com/ceesaxp/tour-guide-editor/internal/types"
	"github.com/h2non/bimg"
)

// S3ClientAPI interface for mocking
type S3ClientAPI interface {
	PutObject(ctx context.Context,
		params *s3.PutObjectInput,
		optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

// MockS3Client implements S3ClientAPI
type MockS3Client struct {
	PutObjectFunc func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

func (m *MockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	if m.PutObjectFunc != nil {
		return m.PutObjectFunc(ctx, params, optFns...)
	}
	return &s3.PutObjectOutput{}, nil
}

func TestMediaService_ProcessAndUpload(t *testing.T) {
	mockS3 := &mocks.MockS3Client{
		PutObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			return &s3.PutObjectOutput{}, nil
		},
	}

	config := MediaConfig{
		MaxFileSize:    1024 * 1024,
		AllowedFormats: []string{"image/", "audio/", "video/"},
		ImageMaxWidth:  800,
		ImageMaxHeight: 600,
		S3Bucket:       "test-bucket",
	}

	service := NewMediaService(config, mockS3)

	tests := []struct {
		name     string
		fileData []byte
		filename string
		wantErr  bool
	}{
		{
			name:     "valid image file",
			fileData: createTestImage(t),
			filename: "test.jpg",
			wantErr:  false,
		},
		{
			name:     "file too large",
			fileData: make([]byte, 2*1024*1024), // 2MB
			filename: "large.jpg",
			wantErr:  true,
		},
		{
			name:     "invalid format",
			fileData: []byte("invalid data"),
			filename: "test.txt",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tempFile, err := os.CreateTemp("", "test-*")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tempFile.Name())
			defer tempFile.Close()

			// Write test data
			if _, err := tempFile.Write(tt.fileData); err != nil {
				t.Fatal(err)
			}

			// Seek back to beginning
			if _, err := tempFile.Seek(0, 0); err != nil {
				t.Fatal(err)
			}

			multipartFile := types.NewMultipartFile(tempFile)

			file := &multipart.FileHeader{
				Filename: tt.filename,
				Size:     int64(len(tt.fileData)),
			}

			_, err = service.ProcessAndUpload(multipartFile, file)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessAndUpload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMediaService_ValidateURL(t *testing.T) {
	config := MediaConfig{
		MaxFileSize:    1024 * 1024,
		AllowedFormats: []string{"image/", "audio/", "video/"},
	}

	service := NewMediaService(config, nil)

	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/valid.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			w.Header().Set("Content-Length", "1000")
			// Write actual JPEG data
			img := createTestImage(t)
			w.Write(img)
		case "/invalid.txt":
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("test content"))
		case "/large.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			w.Header().Set("Content-Length", "2000000")
			// Write large image data
			img := createLargeTestImage(t, 2000, 2000)
			w.Write(img)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid image URL",
			url:     ts.URL + "/valid.jpg",
			wantErr: false,
		},
		{
			name:    "invalid format",
			url:     ts.URL + "/invalid.txt",
			wantErr: true,
		},
		{
			name:    "file too large",
			url:     ts.URL + "/large.jpg",
			wantErr: true,
		},
		{
			name:    "non-existent URL",
			url:     ts.URL + "/notfound.jpg",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMediaService_ProcessImage(t *testing.T) {
	config := MediaConfig{
		MaxFileSize:    1024 * 1024,
		AllowedFormats: []string{"image/"},
		ImageMaxWidth:  800,
		ImageMaxHeight: 600,
	}

	service := NewMediaService(config, nil)

	// Create test image data
	//testImage := createTestImage(t)

	tests := []struct {
		name       string
		imageData  []byte
		wantWidth  int
		wantHeight int
		wantErr    bool
	}{
		{
			name:       "resize large image",
			imageData:  createLargeTestImage(t, 1200, 900),
			wantWidth:  800,
			wantHeight: 600,
			wantErr:    false,
		},
		{
			name:       "keep small image",
			imageData:  createLargeTestImage(t, 400, 300),
			wantWidth:  400,
			wantHeight: 300,
			wantErr:    false,
		},
		{
			name:      "invalid image data",
			imageData: []byte("invalid image data"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processed, err := service.processImage(tt.imageData)
			if (err != nil) != tt.wantErr {
				t.Errorf("processImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify processed image dimensions
				img := bimg.NewImage(processed)
				size, err := img.Size()
				if err != nil {
					t.Errorf("Failed to get processed image size: %v", err)
					return
				}

				if size.Width != tt.wantWidth || size.Height != tt.wantHeight {
					t.Errorf("Processed image size = %dx%d, want %dx%d",
						size.Width, size.Height, tt.wantWidth, tt.wantHeight)
				}
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

func createLargeTestImage(t *testing.T, width, height int) []byte {
	// Create a new image with specified dimensions
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill it with a pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8((x * 255) / width),
				G: uint8((y * 255) / height),
				B: 100,
				A: 255,
			})
		}
	}

	// Encode to JPEG
	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, img, nil); err != nil {
		t.Fatalf("Failed to encode large test image: %v", err)
	}

	return buf.Bytes()
}
