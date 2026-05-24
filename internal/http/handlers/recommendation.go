package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"wb-marketplace/internal/service"
)

// RecommendationHandler handles recommendation API endpoints.
type RecommendationHandler struct {
	svc *service.RecommendationService
}

// NewRecommendationHandler creates a recommendation HTTP handler.
func NewRecommendationHandler(svc *service.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{svc: svc}
}

// ListBySeller handles GET /api/v1/sellers/{sellerId}/recommendations.
func (h *RecommendationHandler) ListBySeller(w http.ResponseWriter, r *http.Request) {
	sellerID, err := parsePositiveID(chi.URLParam(r, "sellerId"), "seller id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	recommendations, err := h.svc.GetActiveBySeller(r.Context(), sellerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load recommendations")
		return
	}

	writeJSON(w, http.StatusOK, recommendations)
}

// View handles POST /api/v1/recommendations/{id}/view.
func (h *RecommendationHandler) View(w http.ResponseWriter, r *http.Request) {
	h.recordRecommendationAction(w, r, h.svc.RecordView)
}

// Accept handles POST /api/v1/recommendations/{id}/accept.
func (h *RecommendationHandler) Accept(w http.ResponseWriter, r *http.Request) {
	h.recordRecommendationAction(w, r, h.svc.RecordAccept)
}

// Reject handles POST /api/v1/recommendations/{id}/reject.
func (h *RecommendationHandler) Reject(w http.ResponseWriter, r *http.Request) {
	h.recordRecommendationAction(w, r, h.svc.RecordReject)
}

func (h *RecommendationHandler) recordRecommendationAction(
	w http.ResponseWriter,
	r *http.Request,
	action func(context.Context, int64) error,
) {
	recommendationID, err := parsePositiveID(chi.URLParam(r, "id"), "recommendation id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := action(r.Context(), recommendationID); err != nil {
		switch {
		case errors.Is(err, service.ErrRecommendationNotFound):
			writeError(w, http.StatusNotFound, "recommendation not found")
		default:
			if err.Error() == "recommendation id must be positive" {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to process recommendation action")
		}
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "ok"})
}

func parsePositiveID(raw, field string) (int64, error) {
	if raw == "" {
		return 0, errors.New(field + " is required")
	}

	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New(field + " must be a positive integer")
	}

	return id, nil
}
