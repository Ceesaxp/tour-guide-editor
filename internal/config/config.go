// internal/config/config.go
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
    Server struct {
        Port     int    `yaml:"port"`
        Host     string `yaml:"host"`
    } `yaml:"server"`
    Auth struct {
        SecretKey string `yaml:"secret_key"`
        TokenTTL  int    `yaml:"token_ttl"`
    } `yaml:"auth"`
    S3 struct {
        MediaBucket string `yaml:"media_bucket"`
        TourBucket  string `yaml:"tour_bucket"`
        Region      string `yaml:"region"`
        Endpoint    string `yaml:"endpoint"`
    } `yaml:"s3"`
    Media struct {
        MaxFileSize    int64    `yaml:"max_file_size"`
        AllowedFormats []string `yaml:"allowed_formats"`
        ImageMaxWidth  int      `yaml:"image_max_width"`
        ImageMaxHeight int      `yaml:"image_max_height"`
    } `yaml:"media"`
}

func Load(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("reading config file: %w", err)
    }

    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parsing config file: %w", err)
    }

    return &cfg, nil
}
