package rules

import "wb-marketplace/internal/domain"

// Match is a rule evaluation result for a seller snapshot.
type Match struct {
	TriggerCode          string
	RuleID               int64
	RecommendationTypeID int64
	Title                string
	Description          string
	ReasonText           string
}

// Engine evaluates seller snapshots against MVP business rules.
type Engine struct{}

// New creates a rule engine instance.
func New() *Engine {
	return &Engine{}
}

// EvaluateSnapshot returns all rules that match the given snapshot.
func (e *Engine) EvaluateSnapshot(snapshot domain.SellerMetricsSnapshot) []Match {
	_ = e

	matches := make([]Match, 0)

	if snapshot.ActiveProductsCount == 0 {
		matches = append(matches, Match{
			TriggerCode:          "NO_PRODUCTS",
			RuleID:               1,
			RecommendationTypeID: 1,
			Title:                "Добавьте первые товары",
			Description:          "У вас пока нет активных товаров. Добавьте товары, чтобы начать продажи.",
			ReasonText:           "У продавца нет активных товаров",
		})
	}

	if snapshot.NoSalesDays >= 14 && snapshot.ActiveProductsCount > 0 {
		matches = append(matches, Match{
			TriggerCode:          "NO_SALES",
			RuleID:               2,
			RecommendationTypeID: 3,
			Title:                "Нет продаж 14 дней",
			Description:          "Рекомендуем проверить ассортимент, цены или модель продаж.",
			ReasonText:           "У продавца нет продаж 14 или более дней",
		})
	}

	if snapshot.LastLoginDays >= 7 {
		matches = append(matches, Match{
			TriggerCode:          "INACTIVE_SELLER",
			RuleID:               3,
			RecommendationTypeID: 6,
			Title:                "Вернитесь к управлению продажами",
			Description:          "Вы давно не заходили в кабинет. Проверьте товары, остатки и продажи.",
			ReasonText:           "Продавец не заходил 7 или более дней",
		})
	}

	if snapshot.ProductsWithoutStockCount > 0 {
		matches = append(matches, Match{
			TriggerCode:          "OUT_OF_STOCK",
			RuleID:               4,
			RecommendationTypeID: 5,
			Title:                "Пополните остатки",
			Description:          "У части товаров закончились остатки. Пополните их, чтобы не терять продажи.",
			ReasonText:           "Есть товары без остатков",
		})
	}

	return matches
}
