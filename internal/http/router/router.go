package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"wb-marketplace/internal/http/handlers"
	"wb-marketplace/internal/service"
)

// New builds the HTTP router with registered routes.
func New(services *service.Services) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", handlers.Health)

	if services == nil {
		return r
	}

	r.Route("/api/v1", func(r chi.Router) {
		if services.Recommendation != nil {
			recHandler := handlers.NewRecommendationHandler(services.Recommendation)
			r.Get("/sellers/{sellerId}/recommendations", recHandler.ListBySeller)

			r.Route("/recommendations", func(r chi.Router) {
				r.Post("/{id}/view", recHandler.View)
				r.Post("/{id}/accept", recHandler.Accept)
				r.Post("/{id}/reject", recHandler.Reject)
			})
		}

		if services.Rule != nil && services.Analysis != nil {
			adminHandler := handlers.NewAdminHandler(services.Rule, services.Analysis)
			r.Route("/admin", func(r chi.Router) {
				r.Get("/rules", adminHandler.ListRules)
				r.Post("/rules", adminHandler.CreateRule)
				r.Post("/run-analysis", adminHandler.RunAnalysis)
			})
		}
	})

	if services.Notification != nil {
		notificationHandler := handlers.NewNotificationHandler(services.Notification)
		r.Post("/internal/notifications/process-outbox", notificationHandler.ProcessOutbox)
	}

	return r
}
