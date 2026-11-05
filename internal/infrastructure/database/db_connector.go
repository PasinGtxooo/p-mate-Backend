package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DBConnector holds the database connection and provides health check functionality
type DBConnector struct {
	DB *gorm.DB
}

// NewDB creates a new database connection with connection pooling configured
func NewDB(cfg *config.Config) *gorm.DB {
	// Default to sslmode=require (needed by Supabase/managed Postgres in prod).
	// Override by setting DB_SSLMODE env (e.g. "disable" for local dev).
	sslMode := os.Getenv("DB_SSLMODE")
	if sslMode == "" {
		sslMode = "require"
	}
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBUsername,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
		sslMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ failed to connect PostgreSQL: %v", err)
	}

	// Configure connection pooling
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("❌ failed to get underlying sql.DB: %v", err)
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns sets the maximum number of open connections to the database
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused
	sqlDB.SetConnMaxLifetime(time.Hour)

	// SetConnMaxIdleTime sets the maximum amount of time a connection may be idle
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	log.Println("✅ Database connection pool configured successfully")

	return db
}

// NewDBConnector creates a new DBConnector instance
func NewDBConnector(db *gorm.DB) *DBConnector {
	return &DBConnector{DB: db}
}

// HealthCheck performs a health check on the database connection
func (c *DBConnector) HealthCheck(ctx context.Context) error {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Use context with timeout for ping
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// GetStats returns database connection pool statistics
func (c *DBConnector) GetStats() map[string]interface{} {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}
