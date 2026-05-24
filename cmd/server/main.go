package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"wb-marketplace/internal/config"
	"wb-marketplace/internal/db"
	"wb-marketplace/internal/http/router"
	"wb-marketplace/internal/notifications"
	"wb-marketplace/internal/repository"
	"wb-marketplace/internal/rules"
	"wb-marketplace/internal/service"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	_ = rules.New()
	_ = notifications.New()

	services := service.New(service.Repositories(
		repository.NewRecommendationRepository(pool),
		repository.NewRuleRepository(pool),
		repository.NewAnalysisRepository(pool),
	))
	handler := router.New(services)
	server := &http.Server{
		Addr:    cfg.HTTPAddr(),
		Handler: handler,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("listening on %s", cfg.HTTPAddr())
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-stop:
		log.Printf("shutdown signal received: %s", sig)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	return server.Shutdown(shutdownCtx)
}
