// internal/config/config_test.go
package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
    // Create temporary config file
    content := []byte(`
server:
  port: 8080
  host: "localhost"
auth:
  secret_key: "test-key"
  token_ttl: 60
s3:
  media_bucket: "test-media"
  tour_bucket: "test-tours"
  region: "us-west-2"
  endpoint: "http://localhost:4566"
media:
  max_file_size: 10485760
  allowed_formats:
    - "image/jpeg"
    - "image/png"
  image_max_width: 2048
  image_max_height: 2048
`)

    tmpfile, err := os.CreateTemp("", "config-*.yaml")
    if err != nil {
        t.Fatal(err)
    }
    defer os.Remove(tmpfile.Name())

    if _, err := tmpfile.Write(content); err != nil {
        t.Fatal(err)
    }
    if err := tmpfile.Close(); err != nil {
        t.Fatal(err)
    }

    // Test loading config
    cfg, err := Load(tmpfile.Name())
    if err != nil {
        t.Fatalf("Failed to load config: %v", err)
    }

    // Verify config values
    if cfg.Server.Port != 8080 {
        t.Errorf("Expected port 8080, got %d", cfg.Server.Port)
    }
    if cfg.Auth.SecretKey != "test-key" {
        t.Errorf("Expected secret_key 'test-key', got %s", cfg.Auth.SecretKey)
    }
}
