// internal/services/tour_service.go
package services

// import (
// 	"fmt"

// 	"github.com/ceesaxp/tour-guide-editor/internal/validators"
// 	"github.com/go-playground/validator/v10"
// )

// type TourService struct {
//     validator *validator.Validate
// }

// func NewTourService() (*TourService, error) {
//     validate := validator.New()

//     // Register custom validations
//     if err := validators.RegisterCustomValidations(validate); err != nil {
//         return nil, fmt.Errorf("registering custom validations: %w", err)
//     }

//     return &TourService{validator: validate}, nil
// }
