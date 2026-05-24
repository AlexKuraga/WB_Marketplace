package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"wb-marketplace/internal/service"
)

// AdminHandler handles admin API endpoints.
type AdminHandler struct {
	ruleSvc     *service.RuleService
	analysisSvc *service.AnalysisService
}

// NewAdminHandler creates an admin HTTP handler.
func NewAdminHandler(ruleSvc *service.RuleService, analysisSvc *service.AnalysisService) *AdminHandler {
	return &AdminHandler{
		ruleSvc:     ruleSvc,
		analysisSvc: analysisSvc,
	}
}

type createRuleRequest struct {
	RuleCode             string  `json:"rule_code"`
	RuleName             string  `json:"rule_name"`
	Description          *string `json:"description,omitempty"`
	RecommendationTypeID int64   `json:"recommendation_type_id"`
	Priority             *int    `json:"priority,omitempty"`
	CooldownDays         *int    `json:"cooldown_days,omitempty"`
	IsActive             *bool   `json:"is_active,omitempty"`
	ConditionExpression  string  `json:"condition_expression"`
	CreatedBy            *string `json:"created_by,omitempty"`
}

// ListRules handles GET /api/v1/admin/rules.
func (h *AdminHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	rules, err := h.ruleSvc.ListActive(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load rules")
		return
	}

	writeJSON(w, http.StatusOK, rules)
}

// CreateRule handles POST /api/v1/admin/rules.
func (h *AdminHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	var req createRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	input := service.CreateRuleInput{
		RuleCode:             req.RuleCode,
		RuleName:             req.RuleName,
		Description:          req.Description,
		RecommendationTypeID: req.RecommendationTypeID,
		ConditionExpression:  req.ConditionExpression,
		CreatedBy:            req.CreatedBy,
		Priority:             100,
		CooldownDays:         0,
		IsActive:             true,
	}
	if req.Priority != nil {
		input.Priority = *req.Priority
	}
	if req.CooldownDays != nil {
		input.CooldownDays = *req.CooldownDays
	}
	if req.IsActive != nil {
		input.IsActive = *req.IsActive
	}

	rule, err := h.ruleSvc.Create(r.Context(), input)
	if err != nil {
		if isValidationError(err) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create rule")
		return
	}

	writeJSON(w, http.StatusCreated, rule)
}

// RunAnalysis handles POST /api/v1/admin/run-analysis.
func (h *AdminHandler) RunAnalysis(w http.ResponseWriter, r *http.Request) {
	if err := h.analysisSvc.RunManualAnalysis(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to run analysis")
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "started"})
}

func isValidationError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "is required") || strings.Contains(msg, "must be")
}
