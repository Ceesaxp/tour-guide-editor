// internal/services/tour_service.go
package services

import (
	"fmt"
	"io"

	"github.com/ceesaxp/tour-guide-editor/internal/models"
	"github.com/ceesaxp/tour-guide-editor/internal/validators"
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type TourService struct {
	validator *validator.Validate
}

func NewTourService() (*TourService, error) {
	validate := validator.New()

	// Register custom validations
	if err := validators.RegisterCustomValidations(validate); err != nil {
		return nil, fmt.Errorf("registering custom validations: %w", err)
	}

	return &TourService{validator: validate}, nil
}

func (s *TourService) ParseTour(r io.Reader) (*models.Tour, error) {
	var tour models.Tour
	decoder := yaml.NewDecoder(r)
	if err := decoder.Decode(&tour); err != nil {
		return nil, fmt.Errorf("parsing tour YAML: %w", err)
	}

	if err := s.validator.Struct(tour); err != nil {
		return nil, fmt.Errorf("validating tour: %w", err)
	}

	return &tour, nil
}

func (s *TourService) ValidateNode(node *models.Node) error {
	return s.validator.Struct(node)
}

func (s *TourService) ValidateEdge(edge *models.Edge) error {
	return s.validator.Struct(edge)
}

func (s *TourService) ExportTour(tour *models.Tour) ([]byte, error) {
	if err := s.validator.Struct(tour); err != nil {
		return nil, fmt.Errorf("validating tour before export: %w", err)
	}

	return yaml.Marshal(tour)
}
