// internal/services/tour_service_test.go
package services

import (
	"strings"
	"testing"

	"github.com/ceesaxp/tour-guide-editor/internal/models"
)

func TestTourService_ParseTour(t *testing.T) {
	service, _ := NewTourService()

	tests := []struct {
		name        string
		yaml        string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid tour",
			yaml: `
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
			wantErr: false,
		},
		{
			name: "invalid tour - missing required fields",
			yaml: `
id: "test_tour"
name: "Test Tour"
`,
			wantErr:     true,
			errContains: "validating tour",
		},
		{
			name: "invalid tour - invalid dates",
			yaml: `
id: "test_tour"
name: "Test Tour"
description: "A test tour"
start_date: "2024-12-31T23:59:59Z"
end_date: "2024-01-01T00:00:00Z"
version: "1.0"
hero_image: "http://example.com/image.jpg"
author:
  name: "Test Author"
  profile_link: "http://example.com/author"
price: 1000
nodes: []
edges: []
`,
			wantErr:     true,
			errContains: "EndDate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.yaml)
			tour, err := service.ParseTour(reader)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if tour == nil {
					t.Error("Expected tour, got nil")
				}
			}
		})
	}
}

func TestTourService_ValidateNode(t *testing.T) {
	service, _ := NewTourService()

	tests := []struct {
		name    string
		node    models.Node
		wantErr bool
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
			wantErr: false,
		},
		{
			name: "invalid node - missing required fields",
			node: models.Node{
				ID: 1,
			},
			wantErr: true,
		},
		{
			name: "invalid node - invalid location",
			node: models.Node{
				ID: 1,
				Location: models.Location{
					Lat: 91.0, // Invalid latitude
					Lon: 20.0,
				},
				ShortDesc: "Test Node",
				Narrative: "Test narrative",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateNode(&tt.node)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
