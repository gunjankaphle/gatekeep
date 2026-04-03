package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/gatekeep/internal/api"
	"github.com/yourusername/gatekeep/internal/config"
	"github.com/yourusername/gatekeep/internal/database"
	"github.com/yourusername/gatekeep/internal/repository"
)

func main() {
	log.Println("🚀 Starting GateKeep API Server...")

	// Load configuration
	cfg := loadConfig()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize components
	configParser := config.NewParser()
	log.Println("✓ Config parser initialized")

	// Optionally connect to PostgreSQL for audit history
	var auditRepo *repository.AuditRepository
	pgPool, err := database.ConnectPostgres(ctx)
	if err != nil {
		log.Printf("⚠ PostgreSQL connection failed: %v - history endpoints will be unavailable", err)
	} else if pgPool != nil {
		defer pgPool.Close()
		auditRepo = repository.NewAuditRepository(pgPool)
		log.Println("✓ PostgreSQL connected - audit history available")
	} else {
		log.Println("ℹ PostgreSQL not configured - history endpoints will be unavailable")
	}

	// Create API router (read-only mode)
	routerConfig := api.RouterConfig{
		AuditRepo:    auditRepo,
		ConfigParser: configParser,
		ConfigPath:   cfg.ConfigPath,
	}

	router := api.NewRouter(routerConfig)
	log.Println("✓ API router configured (read-only mode)")

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Printf("✓ GateKeep API server listening on %s:%s", cfg.Host, cfg.Port)
		log.Printf("  Health check: http://localhost:%s/api/health", cfg.Port)
		log.Printf("  Roles: http://localhost:%s/api/roles", cfg.Port)
		log.Printf("  History: http://localhost:%s/api/sync/history", cfg.Port)
		log.Println()
		log.Println("📖 Read-only mode: YAML files are the source of truth")
		log.Println("   Use the CLI for sync operations: gatekeep sync --config <file>")
		log.Println()

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println()
	log.Println("🛑 Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✓ Server stopped gracefully")
}

// Config holds server configuration
type Config struct {
	Host         string
	Port         string
	ConfigPath   string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// loadConfig loads configuration from environment variables
func loadConfig() Config {
	return Config{
		Host:         getEnv("SERVER_HOST", "0.0.0.0"),
		Port:         getEnv("SERVER_PORT", "8080"),
		ConfigPath:   getEnv("GATEKEEP_CONFIG_PATH", "configs/example.yaml"),
		ReadTimeout:  parseDuration(getEnv("SERVER_READ_TIMEOUT", "30s"), 30*time.Second),
		WriteTimeout: parseDuration(getEnv("SERVER_WRITE_TIMEOUT", "30s"), 30*time.Second),
		IdleTimeout:  parseDuration(getEnv("SERVER_IDLE_TIMEOUT", "60s"), 60*time.Second),
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseDuration parses a duration string with a fallback default
func parseDuration(s string, defaultValue time.Duration) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return defaultValue
	}
	return d
}
