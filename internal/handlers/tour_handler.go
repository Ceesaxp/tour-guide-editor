// internal/handlers/tour_handler.go
package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

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

func (h *EditorHandler) HandleTourMetadata(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get current tour
	tour := h.tourService.GetCurrentTour(r.Context())
	if tour == nil {
		http.Error(w, "No active tour", http.StatusBadRequest)
		return
	}

	// Update tour metadata
	tour.ID = r.FormValue("id")
	tour.Name = r.FormValue("name")
	tour.Description = r.FormValue("description")

	if price, err := strconv.Atoi(r.FormValue("price")); err == nil {
		tour.Price = price
	}

	// Parse dates
	if startDate, err := parseDate(r.FormValue("start_date")); err == nil {
		tour.StartDate = startDate
	}
	if endDate, err := parseDate(r.FormValue("end_date")); err == nil {
		tour.EndDate = endDate
	}

	// Validate and save
	if err := h.tourService.ValidateTour(tour); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.tourService.SaveTour(r.Context(), tour); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success toast message
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Tour metadata updated successfully",
		"type":    "success",
	})
}

func (h *EditorHandler) HandleNodesList(w http.ResponseWriter, r *http.Request) {
	tour := h.tourService.GetCurrentTour(r.Context())
	if tour == nil {
		http.Error(w, "No active tour", http.StatusBadRequest)
		return
	}

	if err := h.templates.ExecuteTemplate(w, "nodes-list", tour.Nodes); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *EditorHandler) HandleNodeEditor(w http.ResponseWriter, r *http.Request) {
	nodeID, _ := strconv.Atoi(r.PathValue("id"))

	tour := h.tourService.GetCurrentTour(r.Context())
	if tour == nil {
		http.Error(w, "No active tour", http.StatusBadRequest)
		return
	}

	var node *models.Node
	if nodeID > 0 {
		node = tour.GetNode(nodeID)
	} else {
		node = models.NewNode()
	}

	if node == nil {
		http.Error(w, "Node not found", http.StatusNotFound)
		return
	}

	data := TemplateData{
		Node: node,
	}

	h.templates.ExecuteTemplate(w, "node-editor", data)
}

func (h *EditorHandler) HandleNodeSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tour := h.tourService.GetCurrentTour(r.Context())
	if tour == nil {
		http.Error(w, "No active tour", http.StatusBadRequest)
		return
	}

	nodeID, _ := strconv.Atoi(r.PathValue("id"))
	node := tour.GetNode(nodeID)
	if node == nil {
		node = models.NewNode()
		node.ID = nodeID
	}

	// Update node data
	if err := h.updateNodeFromForm(node, r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate and save
	if err := h.tourService.ValidateNode(node); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.tourService.SaveNode(r.Context(), tour, node); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Trigger node list update
	w.Header().Set("HX-Trigger", "nodeListChanged")

	// Return success toast message
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Node saved successfully",
		"type":    "success",
	})
}

// Helper functions

func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}
