package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"wb-marketplace/internal/domain"
)

const manualAnalysisJobType = "manual"

// AnalysisService runs batch seller analysis jobs.
type AnalysisService struct {
	repos repositoryBundle
}

// NewAnalysisService creates an analysis service instance.
func NewAnalysisService(repos repositoryBundle) *AnalysisService {
	return &AnalysisService{repos: repos}
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

	sellersProcessed, runErr := s.processSellers(ctx)
	finishedAt := time.Now().UTC()

	finishJob := domain.AnalysisJob{
		ID:               jobID,
		SellersProcessed: sellersProcessed,
		FinishedAt:       &finishedAt,
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

func (s *AnalysisService) processSellers(ctx context.Context) (int, error) {
	sellers, err := s.repos.Analysis.ListSellersForAnalysis(ctx)
	if err != nil {
		return 0, fmt.Errorf("list sellers for analysis: %w", err)
	}

	processed := 0
	for _, seller := range sellers {
		snapshot, err := s.repos.Analysis.GetLatestSnapshot(ctx, seller.ID)
		if err != nil {
			return processed, fmt.Errorf("get latest snapshot for seller %d: %w", seller.ID, err)
		}
		if snapshot == nil {
			continue
		}

		log.Printf("processing seller_id=%d snapshot_id=%d", seller.ID, snapshot.ID)
		processed++
	}

	return processed, nil
}
