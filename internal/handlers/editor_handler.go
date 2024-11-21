// internal/handlers/editor_handler.go
package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/ceesaxp/tour-guide-editor/internal/models"
	"github.com/ceesaxp/tour-guide-editor/internal/services"
)

type EditorHandler struct {
	templates    *template.Template
	tourService  *services.TourService
	mediaService *services.MediaService
}

type TemplateData struct {
	Title string
	Tour  *models.Tour
	Node  *models.Node
	Error string
}

func NewEditorHandler(templateDir string, tourService *services.TourService, mediaService *services.MediaService) *EditorHandler {
	templates, err := template.ParseFiles(
		filepath.Join(templateDir, "layout.html"),
		filepath.Join(templateDir, "editor", "condition.html"),
		filepath.Join(templateDir, "editor", "index.html"),
		filepath.Join(templateDir, "editor", "node.html"),
	)
	if err != nil {
		log.Printf("ERR: error parsing templates: %v", err)
		return nil
	}

	return &EditorHandler{
		templates:    templates,
		tourService:  tourService,
		mediaService: mediaService,
	}
}

func (h *EditorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get current tour from session or create new one
	tour := h.tourService.GetCurrentTour(r.Context())
	if tour == nil {
		tour = models.NewTour()
	}

	// Save tour to session
	if err := h.tourService.SaveTourToSession(r.Context(), tour); err != nil {
		log.Printf("ERR: error saving tour to session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := TemplateData{
		Title: "Tour Editor",
		Tour:  tour,
	}

	if err := h.templates.ExecuteTemplate(w, "layout.html", data); err != nil {
		log.Printf("ERR: error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// Helper functions for node updates
func (h *EditorHandler) updateNodeFromForm(node *models.Node, r *http.Request) error {
	// Basic information
	if id, err := strconv.Atoi(r.FormValue("id")); err == nil {
		node.ID = id
	}
	node.ShortDesc = r.FormValue("short_description")
	node.Narrative = r.FormValue("narrative")

	// Location
	if lat, err := strconv.ParseFloat(r.FormValue("location.lat"), 64); err == nil {
		node.Location.Lat = lat
	}
	if lon, err := strconv.ParseFloat(r.FormValue("location.lon"), 64); err == nil {
		node.Location.Lon = lon
	}

	// Media files
	node.MediaFiles = nil
	for i := 0; ; i++ {
		uri := r.FormValue(fmt.Sprintf("media_files[%d].uri", i))
		if uri == "" {
			break
		}

		delay, _ := strconv.Atoi(r.FormValue(fmt.Sprintf("media_files[%d].send_delay", i)))
		mediaType := r.FormValue(fmt.Sprintf("media_files[%d].type", i))

		node.MediaFiles = append(node.MediaFiles, models.MediaFile{
			Type:      mediaType,
			URI:       uri,
			SendDelay: delay,
		})
	}

	// Update conditions
	if err := h.updateConditionFromForm(node.EntryCondition, "entry_condition", r); err != nil {
		return err
	}
	if err := h.updateConditionFromForm(node.ExitCondition, "exit_condition", r); err != nil {
		return err
	}

	return nil
}

func (h *EditorHandler) updateConditionFromForm(condition *models.Condition, prefix string, r *http.Request) error {
	condType := r.FormValue(prefix + ".type")
	if condType == "" {
		return nil
	}

	if condition == nil {
		condition = &models.Condition{}
	}

	condition.Type = condType
	condition.Question = r.FormValue(prefix + ".question")
	condition.CorrectAnswer = r.FormValue(prefix + ".correct_answer")
	condition.Strict = r.FormValue(prefix+".strict") == "on"
	condition.MediaLink = r.FormValue(prefix + ".media_link")

	// Options
	condition.Options = nil
	for i := 0; ; i++ {
		option := r.FormValue(fmt.Sprintf("%s.options[%d]", prefix, i))
		if option == "" {
			break
		}
		condition.Options = append(condition.Options, option)
	}

	// Hints
	condition.Hints = nil
	for i := 0; ; i++ {
		hint := r.FormValue(fmt.Sprintf("%s.hints[%d]", prefix, i))
		if hint == "" {
			break
		}
		condition.Hints = append(condition.Hints, hint)
	}

	return nil
}
