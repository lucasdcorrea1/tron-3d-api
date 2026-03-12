package main

import (
	"log"
	"net/http"

	"github.com/tron-legacy/tron-3d-api/internal/config"
	"github.com/tron-legacy/tron-3d-api/internal/database"
	"github.com/tron-legacy/tron-3d-api/internal/router"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to MongoDB
	if err := database.Connect(cfg.MongoURI, cfg.DBName); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer database.Disconnect()

	// Ensure indexes
	if err := database.EnsureIndexes(); err != nil {
		log.Printf("Warning: failed to ensure indexes: %v", err)
	}

	// Create router
	r := router.New()

	// Start server
	addr := ":" + cfg.Port
	log.Printf("3D Store API starting on http://localhost%s", addr)
	log.Printf("Health check: http://localhost%s/api/v1/health", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
