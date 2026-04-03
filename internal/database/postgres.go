package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/gatekeep/internal/audit"
	"github.com/yourusername/gatekeep/internal/repository"
)

// ConnectPostgres attempts to connect to PostgreSQL
// Returns nil if connection fails or DSN is not configured
func ConnectPostgres(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		log.Println("PostgreSQL not configured (POSTGRES_DSN not set) - audit logging disabled")
		return nil, nil
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres DSN: %w", err)
	}

	// Set connection pool settings
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	log.Println("✓ PostgreSQL connected - audit logging enabled")
	return pool, nil
}

// NewAuditLogger creates an audit logger based on PostgreSQL availability
// If PostgreSQL is not available, returns a no-op logger
func NewAuditLogger(ctx context.Context) (audit.AuditLogger, *pgxpool.Pool, error) {
	pool, err := ConnectPostgres(ctx)
	if err != nil {
		// Log error but return no-op logger (graceful degradation)
		log.Printf("⚠ PostgreSQL connection failed: %v - using no-op audit logger", err)
		return audit.NewNoOpLogger(), nil, nil
	}

	if pool == nil {
		// PostgreSQL not configured
		return audit.NewNoOpLogger(), nil, nil
	}

	// Create audit repository and logger
	repo := repository.NewAuditRepository(pool)
	logger := audit.NewLogger(repo)

	return logger, pool, nil
}
