package main

import (
	"log"
	"net/http"
	"os"
	"time"

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

	// Keep-alive: prevent Render free tier from sleeping
	if selfURL := os.Getenv("RENDER_EXTERNAL_URL"); selfURL != "" {
		go keepAlive(selfURL + "/api/v1/health")
	}

	// Start server
	addr := ":" + cfg.Port
	log.Printf("3D Store API starting on http://localhost%s", addr)
	log.Printf("Health check: http://localhost%s/api/v1/health", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// keepAlive pings the health endpoint every 14 minutes to prevent Render free tier sleep.
func keepAlive(url string) {
	time.Sleep(10 * time.Second)
	log.Printf("Keep-alive started: pinging %s every 14 min", url)

	ticker := time.NewTicker(14 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Keep-alive ping failed: %v", err)
			continue
		}
		resp.Body.Close()
		log.Printf("Keep-alive ping: %d", resp.StatusCode)
	}
}
