// internal/handlers/editor.go
package handlers

import (
	"net/http"
)

type EditorHandler struct{}

func NewEditorHandler() *EditorHandler {
	return &EditorHandler{}
}

func (h *EditorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement editor handling
	w.Write([]byte("Editor placeholder"))
}
