package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"

	"qr-code-generator/internal/handler"
	"qr-code-generator/internal/qrcode"
	"qr-code-generator/internal/storage"
)

//go:embed templates/*
var templatesFS embed.FS

func main() {
	// Configuration from environment
	port := getEnv("PORT", "8080")
	dbPath := getEnv("DB_PATH", "/data/qrcodes.db")

	// Initialize storage
	store, err := storage.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Initialize QR generator
	generator := qrcode.New()

	// Parse templates
	templates, err := template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	// Initialize handler
	h := handler.New(store, generator, templates)

	// Setup routes
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Start server
	log.Printf("Starting server on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
