package service

import "wb-marketplace/internal/repository"

// Services groups application use-case services.
type Services struct {
	Recommendation *RecommendationService
	Rule           *RuleService
	Analysis       *AnalysisService
	Notification   *NotificationService
}

// New creates the application service layer.
func New(repos repositoryBundle) *Services {
	return &Services{
		Recommendation: NewRecommendationService(repos),
		Rule:           NewRuleService(repos),
		Analysis:       NewAnalysisService(repos),
		Notification:   NewNotificationService(repos),
	}
}

type repositoryBundle struct {
	Recommendations repository.RecommendationRepository
	Rules           repository.RuleRepository
	Analysis        repository.AnalysisRepository
	Notifications   repository.NotificationRepository
}

// Repositories groups data access dependencies for the service layer.
func Repositories(
	recommendations repository.RecommendationRepository,
	rules repository.RuleRepository,
	analysis repository.AnalysisRepository,
	notifications repository.NotificationRepository,
) repositoryBundle {
	return repositoryBundle{
		Recommendations: recommendations,
		Rules:           rules,
		Analysis:        analysis,
		Notifications: notifications,
	}
}
