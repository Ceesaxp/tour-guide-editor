// internal/models/tour_test.go
package models

import (
	"testing"
	"time"

	"github.com/ceesaxp/tour-guide-editor/internal/validators"
	"github.com/go-playground/validator/v10"
)

func setupValidator(t *testing.T) *validator.Validate {
	validate := validator.New()
	err := validators.RegisterCustomValidations(validate)
	if err != nil {
		t.Fatalf("Failed to register custom validations: %v", err)
	}
	return validate
}

func TestTourValidation(t *testing.T) {
	validate := setupValidator(t)

	tests := []struct {
		name    string
		tour    Tour
		wantErr bool
	}{
		{
			name: "valid tour",
			tour: Tour{
				ID:          "test_tour",
				Name:        "Test Tour",
				Description: "A test tour",
				StartDate:   time.Now(),
				EndDate:     time.Now().AddDate(0, 1, 0),
				Version:     "1.0",
				HeroImage:   "http://example.com/image.jpg",
				Author: Author{
					Name:        "Test Author",
					ProfileLink: "http://example.com/author",
				},
				Price: 1000,
				Nodes: []Node{},
				Edges: []Edge{},
			},
			wantErr: false,
		},
		{
			name: "invalid dates",
			tour: Tour{
				ID:          "test_tour",
				Name:        "Test Tour",
				Description: "A test tour",
				StartDate:   time.Now(),
				EndDate:     time.Now().AddDate(0, -1, 0), // End date before start date
				Version:     "1.0",
				HeroImage:   "http://example.com/image.jpg",
				Author: Author{
					Name:        "Test Author",
					ProfileLink: "http://example.com/author",
				},
				Price: 1000,
				Nodes: []Node{},
				Edges: []Edge{},
			},
			wantErr: true,
		},
		{
			name: "invalid price",
			tour: Tour{
				ID:          "test_tour",
				Name:        "Test Tour",
				Description: "A test tour",
				StartDate:   time.Now(),
				EndDate:     time.Now().AddDate(0, 1, 0),
				Version:     "1.0",
				HeroImage:   "http://example.com/image.jpg",
				Author: Author{
					Name:        "Test Author",
					ProfileLink: "http://example.com/author",
				},
				Price: -1000, // Negative price
				Nodes: []Node{},
				Edges: []Edge{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.tour)
			if (err != nil) != tt.wantErr {
				t.Errorf("Tour.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeValidation(t *testing.T) {
	validate := setupValidator(t)

	tests := []struct {
		name    string
		node    Node
		wantErr bool
	}{
		{
			name: "valid node with all fields",
			node: Node{
				ID: 1,
				Location: Location{
					Lat: 45.0,
					Lon: 20.0,
				},
				ShortDesc:      "Test Node",
				Narrative:      "Test narrative",
				AudioNarrative: "http://example.com/audio.ogg",
				MediaFiles: []MediaFile{
					{
						Type:      "image",
						URI:       "http://example.com/image.jpg",
						SendDelay: 0,
					},
				},
				EntryCondition: &Condition{
					Type:          "quiz",
					Strict:        true,
					Question:      "Test question?",
					CorrectAnswer: "Answer",
					Options:       []string{"Answer", "Wrong", "Wrong2"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid location coordinates",
			node: Node{
				ID: 1,
				Location: Location{
					Lat: 91.0,  // Invalid latitude
					Lon: 180.1, // Invalid longitude
				},
				ShortDesc: "Test Node",
				Narrative: "Test narrative",
			},
			wantErr: true,
		},
		{
			name: "invalid quiz condition - not enough options",
			node: Node{
				ID: 1,
				Location: Location{
					Lat: 45.0,
					Lon: 20.0,
				},
				ShortDesc: "Test Node",
				Narrative: "Test narrative",
				EntryCondition: &Condition{
					Type:          "quiz",
					Question:      "Test question?",
					CorrectAnswer: "Answer",
					Options:       []string{"Answer"}, // Need at least 2 options
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("Node.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
