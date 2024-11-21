// internal/handlers/tour_handler.go
package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/ceesaxp/tour-guide-editor/internal/models"
	"github.com/ceesaxp/tour-guide-editor/internal/services"
)

type TourHandler struct {
    tourService *services.TourService
    templates   *template.Template
}

func NewTourHandler(tourService *services.TourService, templates *template.Template) *TourHandler {
    return &TourHandler{
        tourService: tourService,
        templates:   templates,
    }
}

func (h *TourHandler) Upload(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    file, _, err := r.FormFile("tour_file")
    if err != nil {
        http.Error(w, "Invalid file upload", http.StatusBadRequest)
        return
    }
    defer file.Close()

    tour, err := h.tourService.ParseTour(file)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(tour)
}

func (h *TourHandler) ValidateNode(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var node models.Node
    if err := json.NewDecoder(r.Body).Decode(&node); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    if err := h.tourService.ValidateNode(&node); err != nil {
        w.Header().Set("HX-Trigger", "showError")
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    w.Header().Set("HX-Trigger", "showSuccess")
    w.WriteHeader(http.StatusOK)
}
