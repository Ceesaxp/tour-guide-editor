// internal/services/tour_service.go
package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/ceesaxp/tour-guide-editor/internal/models"
	"github.com/ceesaxp/tour-guide-editor/internal/validators"
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type TourService struct {
	validator  *validator.Validate
	tour       *models.Tour
	tourStore  sync.Map // For demo purposes, using in-memory storage
	activeTour sync.Map // Maps session ID to active tour
}

func NewTourService() *TourService {
	validate := validator.New()

	// Register custom validations
	if err := validators.RegisterCustomValidations(validate); err != nil {
		log.Printf("ERR: registering custom validations error: %v", err)
		return nil
	}

	return &TourService{
		validator: validate,
	}
}

func (s *TourService) GetCurrentTour(ctx context.Context) *models.Tour {
	sessionID := ctx.Value("sessionID").(string)
	if tour, ok := s.activeTour.Load(sessionID); ok {
		return tour.(*models.Tour)
	}
	return nil
}

func (s *TourService) ValidateTour(tour *models.Tour) error {
	return s.validator.Struct(tour)
}

func (s *TourService) ValidateNode(node *models.Node) error {
	return s.validator.Struct(node)
}

func (s *TourService) SaveTour(ctx context.Context, tour *models.Tour) error {
	if err := s.ValidateTour(tour); err != nil {
		return err
	}

	sessionID := ctx.Value("sessionID").(string)
	s.activeTour.Store(sessionID, tour)
	s.tourStore.Store(tour.ID, tour)

	return nil
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

func (s *TourService) ValidateEdge(edge *models.Edge) error {
	return s.validator.Struct(edge)
}

func (s *TourService) ExportTour(tour *models.Tour) ([]byte, error) {
	if err := s.validator.Struct(tour); err != nil {
		return nil, fmt.Errorf("validating tour before export: %w", err)
	}

	return yaml.Marshal(tour)
}

func (s *TourService) SaveNode(ctx context.Context, tour *models.Tour, node *models.Node) error {
	if err := s.ValidateNode(node); err != nil {
		return err
	}

	// Update or add node
	found := false
	for i, n := range tour.Nodes {
		if n.ID == node.ID {
			tour.Nodes[i] = *node
			found = true
			break
		}
	}

	if !found {
		tour.Nodes = append(tour.Nodes, *node)
	}

	return s.SaveTour(ctx, tour)
}

func (s *TourService) DeleteNode(ctx context.Context, tour *models.Tour, nodeID int) error {
	// Remove node
	for i, node := range tour.Nodes {
		if node.ID == nodeID {
			tour.Nodes = append(tour.Nodes[:i], tour.Nodes[i+1:]...)
			break
		}
	}

	// Remove associated edges
	var newEdges []models.Edge
	for _, edge := range tour.Edges {
		if edge.From != nodeID && edge.To != nodeID {
			newEdges = append(newEdges, edge)
		}
	}
	tour.Edges = newEdges

	return s.SaveTour(ctx, tour)
}
