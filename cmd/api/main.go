package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/ashparshp/mentormatch-backend/internal/config"
	"github.com/ashparshp/mentormatch-backend/internal/database"
	"github.com/ashparshp/mentormatch-backend/internal/server"
	"github.com/ashparshp/mentormatch-backend/pkg/logger"
)

func main() {
	// Load config
	cfg := config.LoadConfig()

	// Init logger
	logger.Init(cfg.LogLevel)

	// Init database
	db, err := database.NewPostgresPool(cfg.DatabaseURL)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Init server
	srv := server.NewServer(cfg, db)

	// Init and Start Notification Hearbeat (Attendance Reminders)
	notificationService := server.SetupNotificationService(db)
	go notificationService.StartReminderTicker(context.Background(), 5*time.Minute)

	// Start server
	if err := srv.Start(); err != nil && err.Error() != "http: Server closed" {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
