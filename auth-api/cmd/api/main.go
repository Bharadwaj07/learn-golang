package main

import (
	"auth-api/internal/domain"
	"auth-api/internal/handler"
	"auth-api/internal/repository"
	"auth-api/internal/router"
	"auth-api/internal/service"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1. load env
	if err := godotenv.Load(); err != nil {
		slog.Warn("no .env file found")
	}

	// 2. setup logger
	setupLogger(os.Getenv("ENV"))

	// 3. connect db
	db, err := openDB()
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	slog.Info("database connected")

	// 4. wire everything together
	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo, os.Getenv("JWT_SECRET"))
	userHandler := handler.NewUserHandler(userSvc)

	// 5. setup router
	r := router.Setup(userHandler, os.Getenv("JWT_SECRET"))

	// 6. create server
	server := &http.Server{
		Addr:         ":" + os.Getenv("PORT"),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 7. start server in goroutine
	go func() {
		slog.Info("server started", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	// 8. wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutdown signal received")

	// 9. graceful shutdown — 30s for in-flight requests
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("forced shutdown", "err", err)
		os.Exit(1)
	}

	slog.Info("server stopped cleanly")
}

func setupLogger(env string) {
	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}

	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}

func openDB() (*gorm.DB, error) {
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DB_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("opening db: %w", err)
	}

	if err := db.AutoMigrate(&domain.User{}); err != nil {
		return nil, fmt.Errorf("migrating db: %w", err)
	}

	return db, nil
}
