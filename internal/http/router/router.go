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

	if services != nil && services.Recommendation != nil {
		recHandler := handlers.NewRecommendationHandler(services.Recommendation)

		r.Route("/api/v1", func(r chi.Router) {
			r.Get("/sellers/{sellerId}/recommendations", recHandler.ListBySeller)

			r.Route("/recommendations", func(r chi.Router) {
				r.Post("/{id}/view", recHandler.View)
				r.Post("/{id}/accept", recHandler.Accept)
				r.Post("/{id}/reject", recHandler.Reject)
			})
		})
	}

	return r
}
