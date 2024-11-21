// internal/services/media_service.go
package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gabriel-vasile/mimetype"
	"github.com/h2non/bimg"

	"github.com/ceesaxp/tour-guide-editor/internal/mocks"
)

type MediaConfig struct {
	MaxFileSize    int64    `yaml:"max_file_size"`
	AllowedFormats []string `yaml:"allowed_formats"`
	ImageMaxWidth  int      `yaml:"image_max_width"`
	ImageMaxHeight int      `yaml:"image_max_height"`
	S3Bucket       string   `yaml:"s3_bucket"`
}

type MediaService struct {
	config   MediaConfig
	s3Client mocks.S3Client
}

func NewMediaService(config MediaConfig, s3Client mocks.S3Client) *MediaService {
	return &MediaService{
		config:   config,
		s3Client: s3Client,
	}
}

type ProcessedMedia struct {
	URL      string
	Hash     string
	MimeType string
	Size     int64
}

func (s *MediaService) ProcessAndUpload(file multipart.File, header *multipart.FileHeader) (*ProcessedMedia, error) {
	// Read file into memory for processing
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	// Check file size
	if int64(len(data)) > s.config.MaxFileSize {
		return nil, fmt.Errorf("file too large: %d > %d", len(data), s.config.MaxFileSize)
	}

	// Detect MIME type
	mime := mimetype.Detect(data)
	if !s.isAllowedFormat(mime.String()) {
		return nil, fmt.Errorf("unsupported file format: %s", mime.String())
	}

	// Process media based on type
	processed, err := s.processMedia(data, mime.String())
	if err != nil {
		return nil, fmt.Errorf("processing media: %w", err)
	}

	// Generate hash
	hash := sha256.Sum256(processed)
	hashString := hex.EncodeToString(hash[:])

	// Generate S3 key
	extension := filepath.Ext(header.Filename)
	s3Key := fmt.Sprintf("%s/%s%s", time.Now().Format("2006/01/02"), hashString, extension)

	// Check if file already exists
	exists, existingURL, err := s.checkFileExists(hashString)
	if err != nil {
		return nil, fmt.Errorf("checking file existence: %w", err)
	}
	if exists {
		return &ProcessedMedia{
			URL:      existingURL,
			Hash:     hashString,
			MimeType: mime.String(),
			Size:     int64(len(processed)),
		}, nil
	}

	// Upload to S3
	url, err := s.uploadToS3(processed, s3Key, mime.String())
	if err != nil {
		return nil, fmt.Errorf("uploading to S3: %w", err)
	}

	return &ProcessedMedia{
		URL:      url,
		Hash:     hashString,
		MimeType: mime.String(),
		Size:     int64(len(processed)),
	}, nil
}

func (s *MediaService) ProcessURL(url string) (*ProcessedMedia, error) {
	// Validate URL
	if err := s.ValidateURL(url); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Download file
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("downloading file: %w", err)
	}
	defer resp.Body.Close()

	// Read the data
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	// Create a temporary file
	tempFile, err := os.CreateTemp("", "download-*")
	if err != nil {
		return nil, fmt.Errorf("creating temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Write the data to the temporary file
	if _, err := tempFile.Write(data); err != nil {
		return nil, fmt.Errorf("writing to temporary file: %w", err)
	}

	// Seek back to the beginning
	if _, err := tempFile.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("seeking temporary file: %w", err)
	}

	header := &multipart.FileHeader{
		Filename: filepath.Base(url),
		Size:     int64(len(data)),
	}

	return s.ProcessAndUpload(tempFile, header)
}

func (s *MediaService) processMedia(data []byte, mimeType string) ([]byte, error) {
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return s.processImage(data)
	case strings.HasPrefix(mimeType, "audio/"):
		return s.processAudio(data)
	case strings.HasPrefix(mimeType, "video/"):
		return s.processVideo(data)
	default:
		return data, nil
	}
}

func (s *MediaService) processImage(data []byte) ([]byte, error) {
	img := bimg.NewImage(data)

	// Get image size
	size, err := img.Size()
	if err != nil {
		return nil, fmt.Errorf("getting image size: %w", err)
	}

	// Check if resize is needed
	if size.Width > s.config.ImageMaxWidth || size.Height > s.config.ImageMaxHeight {
		options := bimg.Options{
			Width:  s.config.ImageMaxWidth,
			Height: s.config.ImageMaxHeight,
			Embed:  true,
		}

		processed, err := img.Process(options)
		if err != nil {
			return nil, fmt.Errorf("processing image: %w", err)
		}
		return processed, nil
	}

	return data, nil
}

func (s *MediaService) processAudio(data []byte) ([]byte, error) {
	// TODO: Implement audio processing with ffmpeg
	// For now, return original data
	return data, nil
}

func (s *MediaService) processVideo(data []byte) ([]byte, error) {
	// TODO: Implement video processing with ffmpeg
	// For now, return original data
	return data, nil
}

func (s *MediaService) uploadToS3(data []byte, key string, contentType string) (string, error) {
	ctx := context.Background()

	_, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.config.S3Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("uploading to S3: %w", err)
	}

	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.config.S3Bucket, key), nil
}

func (s *MediaService) checkFileExists(hash string) (bool, string, error) {
	// TODO: Implement file existence check in S3 or database
	return false, hash, nil // File does not exist
}

func (s *MediaService) ValidateURL(url string) error {
	resp, err := http.Head(url)
	if err != nil {
		return fmt.Errorf("checking URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid URL status: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !s.isAllowedFormat(contentType) {
		return fmt.Errorf("unsupported content type: %s", contentType)
	}

	contentLength := resp.ContentLength
	if contentLength > s.config.MaxFileSize {
		return fmt.Errorf("file too large: %d > %d", contentLength, s.config.MaxFileSize)
	}

	return nil
}

func (s *MediaService) isAllowedFormat(mimeType string) bool {
	for _, format := range s.config.AllowedFormats {
		if strings.HasPrefix(mimeType, format) {
			return true
		}
	}
	return false
}
