package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/interfaces/http/handlers"
	"github.com/FrostBitzX/smart-task-ai/internal/interfaces/http/middlewares"
	"github.com/FrostBitzX/smart-task-ai/internal/interfaces/http/routes"
	"gorm.io/gorm"

	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/config"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/database"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

// matchOriginPattern returns true if origin matches pattern. Pattern may
// contain a single '*' wildcard anywhere, so Vercel preview URLs like
// "https://p-mate-fontend-abc123.vercel.app" can match "https://p-mate-fontend-*.vercel.app".
func matchOriginPattern(pattern, origin string) bool {
	if !strings.Contains(pattern, "*") {
		return pattern == origin
	}
	parts := strings.SplitN(pattern, "*", 2)
	prefix, suffix := parts[0], parts[1]
	return len(origin) >= len(prefix)+len(suffix) &&
		strings.HasPrefix(origin, prefix) &&
		strings.HasSuffix(origin, suffix)
}

type Application struct {
	Server *fiber.App
	Config *config.Config
	DB     *gorm.DB
	Logger logger.Logger
}

func main() {
	_ = godotenv.Load()

	cfg := config.NewConfig()
	db := database.NewDB(cfg)
	dbConnector := database.NewDBConnector(db)

	app := fiber.New(fiber.Config{
		AppName: "smart-task-ai",
	})

	zapLogger := logger.NewZapLogger()

	// Panic recovery middleware (should be first)
	app.Use(middlewares.RecoverMiddleware(zapLogger))

	// Request timeout middleware
	app.Use(middlewares.TimeoutMiddleware(middlewares.TimeoutConfig{
		Timeout:      30 * time.Second,
		ErrorMessage: "Request timeout",
	}))

	// Logger middleware
	app.Use(middlewares.FiberLoggerMiddleware(zapLogger))

	// CORS middleware. CORS_ALLOW_ORIGINS is a comma-separated list; each entry
	// may contain a single '*' wildcard so Vercel preview URLs are all accepted
	// (e.g. "https://p-mate-fontend-*.vercel.app"). AllowCredentials rules out
	// plain "*", so we use AllowOriginsFunc when any pattern is present.
	allowOrigins := os.Getenv("CORS_ALLOW_ORIGINS")
	if allowOrigins == "" {
		allowOrigins = "http://localhost:3000"
	}

	rawPatterns := strings.Split(allowOrigins, ",")
	patterns := make([]string, 0, len(rawPatterns))
	hasWildcard := false
	for _, p := range rawPatterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		patterns = append(patterns, p)
		if strings.Contains(p, "*") {
			hasWildcard = true
		}
	}

	corsCfg := cors.Config{
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}
	if hasWildcard {
		corsCfg.AllowOriginsFunc = func(origin string) bool {
			for _, p := range patterns {
				if matchOriginPattern(p, origin) {
					return true
				}
			}
			return false
		}
	} else {
		corsCfg.AllowOrigins = strings.Join(patterns, ",")
	}

	app.Use(cors.New(corsCfg))

	// Health check route (public, no auth required)
	healthHandler := handlers.NewHealthHandler(dbConnector)
	app.Get("/health", healthHandler.Health)

	// Application routes
	routes.RegisterPublicRoutes(app, db, zapLogger)
	routes.RegisterPrivateRoutes(app, db, zapLogger)

	port := os.Getenv("PORT")
	addr := ":" + port

	// Graceful shutdown
	go func() {
		log.Printf("🚀 server running on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("❌ server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	if err := app.Shutdown(); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}
