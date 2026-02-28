package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/thucdx/todovibe/internal/config"
	"github.com/thucdx/todovibe/internal/db"
	"github.com/thucdx/todovibe/internal/handlers"
	"github.com/thucdx/todovibe/internal/middleware"
	"github.com/thucdx/todovibe/internal/models"
	"github.com/thucdx/todovibe/internal/repository"
	"github.com/thucdx/todovibe/internal/services"
)

func main() {
	cfg := config.Load()

	// Connect to database
	database, err := db.Connect(cfg.DBDSN)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}

	// Run migrations
	if err := db.Migrate(database); err != nil {
		slog.Error("migration failed", "err", err)
		os.Exit(1)
	}
	slog.Info("migrations applied successfully")

	// Repositories
	taskRepo     := repository.NewTaskRepo(database)
	sessionRepo  := repository.NewSessionRepo(database)
	settingsRepo := repository.NewSettingsRepo(database)

	// Services
	authSvc     := services.NewAuthService(settingsRepo, sessionRepo)
	taskSvc     := services.NewTaskService(taskRepo)
	calendarSvc := services.NewCalendarService(taskRepo)
	statsSvc    := services.NewStatsService(taskRepo)

	// Handle PIN reset from env var (best-effort on startup)
	if cfg.PINReset != "" {
		if err := authSvc.ResetPIN(context.Background(), cfg.PINReset); err != nil {
			slog.Error("PIN reset failed", "err", err)
		} else {
			slog.Info("PIN reset from APP_PIN_RESET env var")
		}
	}

	// Handlers
	authH     := handlers.NewAuthHandler(authSvc, sessionRepo)
	taskH     := handlers.NewTaskHandler(taskSvc)
	calendarH := handlers.NewCalendarHandler(calendarSvc)
	statsH    := handlers.NewStatsHandler(statsSvc)

	// Router
	r := gin.New()
	r.Use(middleware.Logger(), gin.Recovery())

	// Health check — unauthenticated
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")

	// Auth routes — no session required
	auth := api.Group("/auth")
	auth.GET("/status", authH.Status)
	auth.POST("/setup", authH.Setup)
	auth.POST("/login", authH.Login)
	auth.POST("/logout", authH.Logout)

	// Protected routes — require valid session cookie
	protected := api.Group("/")
	protected.Use(middleware.Auth(sessionRepo))
	protected.GET("/tasks", taskH.List)
	protected.POST("/tasks", taskH.Create)
	protected.PUT("/tasks/:id", taskH.Update)
	protected.PATCH("/tasks/:id/done", taskH.ToggleDone)
	protected.DELETE("/tasks/:id", taskH.Delete)
	protected.GET("/calendar", calendarH.Summary)
	protected.GET("/stats", statsH.Chart)

	if os.Getenv("APP_ENV") == "test" {
		api.POST("/test/reset", func(c *gin.Context) {
			ctx := c.Request.Context()
			_ = taskRepo.DeleteAll(ctx)
			_ = sessionRepo.DeleteAll(ctx)
			_ = settingsRepo.Delete(ctx, models.PinHashKey)
			c.Status(http.StatusNoContent)
		})
	}

	slog.Info("server starting", "port", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}
