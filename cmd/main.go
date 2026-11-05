package main

import (
	"log"
	"os"
	"os/signal"
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

	// CORS middleware
	allowOrigins := os.Getenv("CORS_ALLOW_ORIGINS")
	if allowOrigins == "" {
		allowOrigins = "http://localhost:3000"
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

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
