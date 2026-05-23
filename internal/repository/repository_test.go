package repository

import "testing"

func TestRepositoryInterfacesCompile(t *testing.T) {
	var _ RecommendationRepository
	var _ RuleRepository
	var _ AnalysisRepository
}
