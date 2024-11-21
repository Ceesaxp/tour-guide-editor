// cmd/server/main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/ceesaxp/tour-guide-editor/internal/config"
	"github.com/ceesaxp/tour-guide-editor/internal/handlers"
	"github.com/ceesaxp/tour-guide-editor/internal/middleware"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	server := setupServer(cfg)
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s", addr)
	log.Fatal(http.ListenAndServe(addr, server))
}

func setupServer(cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	// Authentication endpoints
	auth := handlers.NewAuthHandler(cfg.Auth)
	mux.HandleFunc("/auth/login", auth.Login)
	mux.HandleFunc("/auth/logout", auth.Logout)

	// Protected routes
	//protected := middleware.RequireAuth(cfg.Auth.SecretKey)
	//mux.Handle("/editor/", protected(handlers.NewEditorHandler()))
	mux.Handle("/editor/", handlers.NewEditorHandler())

	return middleware.Logger(mux)
}
