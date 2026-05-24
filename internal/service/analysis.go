package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"wb-marketplace/internal/domain"
	"wb-marketplace/internal/rules"
)

const manualAnalysisJobType = "manual"

const (
	recommendationPriority = 100
	recommendationScore    = 0.8
	notificationChannel    = "in_app"
)

// AnalysisService runs batch seller analysis jobs.
type AnalysisService struct {
	repos      repositoryBundle
	ruleEngine *rules.Engine
}

// NewAnalysisService creates an analysis service instance.
func NewAnalysisService(repos repositoryBundle) *AnalysisService {
	return &AnalysisService{
		repos:      repos,
		ruleEngine: rules.New(),
	}
}

type analysisRunStats struct {
	sellersProcessed       int
	triggersCreated        int
	recommendationsCreated int
}

// RunManualAnalysis executes a synchronous manual analysis run.
func (s *AnalysisService) RunManualAnalysis(ctx context.Context) error {
	if s.repos.Analysis == nil {
		return errors.New("analysis repository not configured")
	}

	now := time.Now().UTC()
	jobID, err := s.repos.Analysis.CreateAnalysisJob(ctx, domain.AnalysisJob{
		JobType:   manualAnalysisJobType,
		Status:    "running",
		StartedAt: &now,
	})
	if err != nil {
		return fmt.Errorf("create analysis job: %w", err)
	}

	stats, runErr := s.processSellers(ctx)
	finishedAt := time.Now().UTC()

	finishJob := domain.AnalysisJob{
		ID:                     jobID,
		SellersProcessed:       stats.sellersProcessed,
		TriggersCreated:        stats.triggersCreated,
		RecommendationsCreated: stats.recommendationsCreated,
		FinishedAt:             &finishedAt,
	}

	if runErr != nil {
		finishJob.Status = "failed"
		errMsg := runErr.Error()
		finishJob.ErrorMessage = &errMsg
	} else {
		finishJob.Status = "success"
	}

	if err := s.repos.Analysis.UpdateAnalysisJob(ctx, finishJob); err != nil {
		return fmt.Errorf("finish analysis job: %w", err)
	}

	return runErr
}

func (s *AnalysisService) processSellers(ctx context.Context) (analysisRunStats, error) {
	stats := analysisRunStats{}

	sellers, err := s.repos.Analysis.ListSellersForAnalysis(ctx)
	if err != nil {
		return stats, fmt.Errorf("list sellers for analysis: %w", err)
	}

	for _, seller := range sellers {
		snapshot, err := s.repos.Analysis.GetLatestSnapshot(ctx, seller.ID)
		if err != nil {
			return stats, fmt.Errorf("get latest snapshot for seller %d: %w", seller.ID, err)
		}
		if snapshot == nil {
			continue
		}

		log.Printf("processing seller_id=%d snapshot_id=%d", seller.ID, snapshot.ID)
		stats.sellersProcessed++

		matches := s.ruleEngine.EvaluateSnapshot(*snapshot)
		for _, match := range matches {
			triggered, recommended, err := s.applyMatch(ctx, seller, *snapshot, match)
			if err != nil {
				return stats, fmt.Errorf("apply rule %s for seller %d: %w", match.TriggerCode, seller.ID, err)
			}
			stats.triggersCreated += triggered
			stats.recommendationsCreated += recommended
		}
	}

	return stats, nil
}

func (s *AnalysisService) applyMatch(
	ctx context.Context,
	seller domain.Seller,
	snapshot domain.SellerMetricsSnapshot,
	match rules.Match,
) (triggersCreated int, recommendationsCreated int, err error) {
	now := time.Now().UTC()
	snapshotID := snapshot.ID
	periodKey := snapshot.SnapshotDate.Format("2006-01-02")

	triggerID, err := s.repos.Analysis.CreateTrigger(ctx, domain.SellerTriggerLog{
		SellerID:    seller.ID,
		RuleID:      match.RuleID,
		TriggerCode: match.TriggerCode,
		TriggeredAt: now,
		PeriodKey:   periodKey,
		SnapshotID:  &snapshotID,
		Status:      "detected",
	})
	if err != nil {
		if isDuplicateTriggerError(err) {
			return 0, 0, nil
		}
		return 0, 0, err
	}
	triggersCreated++

	score := recommendationScore
	expiresAt := now.Add(14 * 24 * time.Hour)
	reasonText := match.ReasonText

	recommendationID, err := s.repos.Analysis.CreateRecommendation(ctx, domain.Recommendation{
		SellerID:             seller.ID,
		TriggerID:            triggerID,
		RecommendationTypeID: match.RecommendationTypeID,
		Title:                match.Title,
		Description:          match.Description,
		ReasonText:           &reasonText,
		Priority:             recommendationPriority,
		Score:                &score,
		Status:               "created",
		ExpiresAt:            &expiresAt,
	})
	if err != nil {
		return triggersCreated, 0, err
	}
	recommendationsCreated++

	err = s.repos.Analysis.CreateNotificationLog(ctx, domain.SellerNotificationLog{
		SellerID:         seller.ID,
		RecommendationID: recommendationID,
		ChannelCode:      notificationChannel,
		Status:           "ready_to_send",
	})
	if err != nil {
		return triggersCreated, recommendationsCreated, err
	}

	return triggersCreated, recommendationsCreated, nil
}

func isDuplicateTriggerError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique")
}
