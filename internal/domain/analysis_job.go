package domain

import "time"

// AnalysisJob maps to the analysis_jobs table.
type AnalysisJob struct {
	ID                     int64      `json:"id"`
	JobType                string     `json:"job_type"`
	Status                 string     `json:"status"`
	StartedAt              *time.Time `json:"started_at,omitempty"`
	FinishedAt             *time.Time `json:"finished_at,omitempty"`
	SellersProcessed       int        `json:"sellers_processed"`
	RecommendationsCreated int        `json:"recommendations_created"`
	TriggersCreated        int        `json:"triggers_created"`
	ErrorMessage           *string    `json:"error_message,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
}
