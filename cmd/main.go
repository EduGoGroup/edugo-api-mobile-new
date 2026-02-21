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

	"github.com/gin-gonic/gin"

	"github.com/EduGoGroup/edugo-shared/logger"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/config"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/container"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/infrastructure/http/router"
)

func main() {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	// Create logger
	appLogger := createLogger(cfg)
	defer func() { _ = appLogger.Sync() }()

	// Configure Gin mode
	if os.Getenv("APP_ENV") == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create dependency container
	c, err := container.New(ctx, cfg, appLogger)
	if err != nil {
		appLogger.Fatal("failed to initialize container", "error", err)
	}
	defer c.Close()

	appLogger.Info("dependency container initialized")

	// Setup router
	r := router.Setup(c)

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		appLogger.Info("HTTP server starting", "address", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("HTTP server failed", "error", err)
		}
	}()

	// Wait for interrupt signal
	sig := <-quit
	appLogger.Info("shutdown signal received", "signal", sig.String())

	// Give outstanding requests time to finish
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("server forced to shutdown", "error", err)
	}

	appLogger.Info("server exited gracefully")
}

func createLogger(cfg *config.Config) logger.Logger {
	return logger.NewZapLogger(cfg.Logging.Level, cfg.Logging.Format)
}
