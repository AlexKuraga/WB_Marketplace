package domain

import "testing"

func TestDomainModelsCompile(t *testing.T) {
	_ = Seller{}
	_ = Recommendation{}
	_ = RecommendationType{}
	_ = RecommendationRule{}
	_ = SellerMetricsSnapshot{}
	_ = SellerTriggerLog{}
	_ = SellerNotificationLog{}
	_ = RecommendationFeedback{}
	_ = AnalysisJob{}
}
