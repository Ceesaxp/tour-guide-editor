package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ceesaxp/tour-guide-editor/internal/config"
	"github.com/ceesaxp/tour-guide-editor/internal/handlers"
	"github.com/ceesaxp/tour-guide-editor/internal/middleware"
	"github.com/ceesaxp/tour-guide-editor/internal/mocks"
	"github.com/ceesaxp/tour-guide-editor/internal/services"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("TTL is %d", cfg.Auth.TokenTTL)

	// Initialize services
	tourService := services.NewTourService()

	// For development, use mock S3 client
	mockS3 := &mocks.MockS3Client{
		PutObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			return &s3.PutObjectOutput{}, nil
		},
	}

	mediaService := services.NewMediaService(services.MediaConfig{
		MaxFileSize:    cfg.Media.MaxFileSize,
		AllowedFormats: cfg.Media.AllowedFormats,
		ImageMaxWidth:  cfg.Media.ImageMaxWidth,
		ImageMaxHeight: cfg.Media.ImageMaxHeight,
		S3Bucket:       cfg.S3.MediaBucket,
	}, mockS3)

	// Initialize auth handler
	authTemplates, err := template.ParseGlob(filepath.Join("templates", "*.html"))
	if err != nil {
		log.Printf("ERR: error parsing templates: %v", err)
	}
	authService := services.NewAuthService(
		cfg.Auth.SecretKey,
		time.Duration(cfg.Auth.TokenTTL)*time.Minute,
	)
	authHandler := handlers.NewAuthHandler(cfg.Auth, authTemplates, authService)

	// Initialize editor handler
	editorHandler := handlers.NewEditorHandler("templates", tourService, mediaService)
	if err != nil {
		log.Fatalf("Failed to create editor handler: %v", err)
	}

	// Setup routes with middleware
	router := setupRoutes(editorHandler, authHandler, cfg.Auth)

	// Add global middleware
	handler := middleware.Chain(
		router,
		middleware.Logger,
		//middleware.RequireAuth(cfg.Auth.SecretKey),
		middleware.SessionID, // Add this middleware to generate session IDs
	)

	// Create server
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: handler,
	}

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	// Start the server
	go func() {
		log.Printf("Server listening on %s", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	// Channel to listen for an interrupt or terminate signal from the OS.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		log.Fatalf("Error starting server: %v", err)

	case <-shutdown:
		log.Println("Starting shutdown")

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Asking listener to shut down and shed load.
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Graceful shutdown did not complete in %v: %v", 5*time.Second, err)
			if err := server.Close(); err != nil {
				log.Printf("Error killing server: %v", err)
			}
		}
	}
}

// Update cmd/server/main.go setupRoutes
func setupRoutes(e *handlers.EditorHandler, a *handlers.AuthHandler, cfg config.Auth) http.Handler {
	mux := http.NewServeMux()

	// Auth routes (unprotected)
	mux.HandleFunc("/login", a.ServeLogin)
	mux.HandleFunc("/auth/login", a.HandleLogin)
	mux.HandleFunc("/logout", a.Logout)

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Protected routes
	protected := middleware.RequireAuth(cfg.SecretKey)

	// Editor routes
	mux.Handle("/", protected(http.HandlerFunc(e.ServeHTTP)))
	mux.Handle("/tour/metadata", protected(http.HandlerFunc(e.HandleTourMetadata)))
	//mux.Handle("/tour/preview", protected(http.HandlerFunc(e.HandleTourPreview)))
	//mux.Handle("/tour/export", protected(http.HandlerFunc(e.HandleTourExport)))
	//mux.Handle("/nodes/new", protected(http.HandlerFunc(e.HandleNewNode)))
	mux.Handle("/nodes", protected(http.HandlerFunc(e.HandleNodesList)))
	mux.Handle("/nodes/{id}/edit", protected(http.HandlerFunc(e.HandleNodeEditor)))
	mux.Handle("/nodes/{id}", protected(http.HandlerFunc(e.HandleNodeSave)))
	mux.Handle("/media/upload", protected(http.HandlerFunc(e.HandleMediaUpload)))
	mux.Handle("/media/validate-url", protected(http.HandlerFunc(e.HandleMediaValidation)))

	return mux
}
