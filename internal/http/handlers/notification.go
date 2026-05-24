package handlers

import (
	"net/http"
	"strconv"

	"wb-marketplace/internal/service"
)

// NotificationHandler handles internal notification processing endpoints.
type NotificationHandler struct {
	svc *service.NotificationService
}

// NewNotificationHandler creates a notification HTTP handler.
func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

type processOutboxResponse struct {
	Status    string `json:"status"`
	Processed int    `json:"processed"`
}

// ProcessOutbox handles POST /internal/notifications/process-outbox.
func (h *NotificationHandler) ProcessOutbox(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			writeError(w, http.StatusBadRequest, "limit must be a positive integer")
			return
		}
		limit = parsed
	}

	processed, err := h.svc.ProcessReadyNotifications(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to process notifications")
		return
	}

	writeJSON(w, http.StatusOK, processOutboxResponse{
		Status:    "ok",
		Processed: processed,
	})
}
