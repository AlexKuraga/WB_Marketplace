package rules

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"wb-marketplace/internal/domain"
)

// Match is a rule evaluation result for a seller snapshot.
type Match struct {
	TriggerCode          string
	RuleID               int64
	RecommendationTypeID int64
	Title                string
	Description          string
	ReasonText           string
}

// Engine evaluates seller snapshots using Yandex AI with a safe local fallback.
type Engine struct {
	apiKey   string
	folderID string
	promptID string
	client   *http.Client
	timeout  time.Duration
}

const (
	yandexResponsesURL = ""

	defaultYandexAPIKey   = ""
	defaultYandexFolderID = ""
	defaultYandexPromptID = ""

	defaultAIRequestTimeout = 20 * time.Second
)

// New creates an engine instance.
func New() *Engine {
	return &Engine{
		apiKey:   envString("YANDEX_API_KEY", defaultYandexAPIKey),
		folderID: envString("YANDEX_FOLDER_ID", defaultYandexFolderID),
		promptID: envString("YANDEX_PROMPT_ID", defaultYandexPromptID),
		client: &http.Client{
			Timeout: defaultAIRequestTimeout,
		},
		timeout: defaultAIRequestTimeout,
	}
}

// EvaluateSnapshot returns all recommendations for the given snapshot.
// It first tries Yandex AI. If the AI call or parsing fails, it falls back to local rules.
func (e *Engine) EvaluateSnapshot(snapshot domain.SellerMetricsSnapshot) []Match {
	if e == nil {
		e = New()
	}

	matches, err := e.evaluateWithAI(snapshot)
	if err != nil {
		log.Printf("rules: AI evaluation failed, using local fallback: %v", err)
		return e.evaluateLocal(snapshot)
	}

	return matches
}

type promptRequest struct {
	Prompt promptRef `json:"prompt"`
	Input  string    `json:"input"`
}

type promptRef struct {
	ID string `json:"id"`
}

type yandexResponse struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Error   *yandexError `json:"error"`
	Output  []yandexOutputItem `json:"output"`
	Text    any `json:"text"`
}

type yandexError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type yandexOutputItem struct {
	ID      string `json:"id"`
	Role    string `json:"role"`
	Status  string `json:"status"`
	Content []yandexContentItem `json:"content"`
}

type yandexContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type aiMatchesEnvelope struct {
	Matches []aiMatch `json:"matches"`
}

type aiMatch struct {
	TriggerCode string `json:"trigger_code"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ReasonText  string `json:"reason_text"`
}

type ruleMeta struct {
	RuleID               int64
	RecommendationTypeID int64
	FallbackTitle        string
	FallbackDescription  string
	FallbackReasonText   string
}

var allowedRules = map[string]ruleMeta{
	"NO_PRODUCTS": {
		RuleID:               1,
		RecommendationTypeID: 1,
		FallbackTitle:        "Добавьте первые товары",
		FallbackDescription:  "У вас пока нет активных товаров. Добавьте товары, чтобы начать продажи.",
		FallbackReasonText:   "У продавца нет активных товаров",
	},
	"NO_SALES": {
		RuleID:               2,
		RecommendationTypeID: 3,
		FallbackTitle:        "Нет продаж",
		FallbackDescription:  "Проверьте ассортимент, цены, карточки товаров и каналы привлечения трафика.",
		FallbackReasonText:   "У продавца отсутствуют продажи в течение длительного периода",
	},
	"INACTIVE_SELLER": {
		RuleID:               3,
		RecommendationTypeID: 6,
		FallbackTitle:        "Вернитесь к управлению продажами",
		FallbackDescription:  "Вы давно не заходили в кабинет. Проверьте товары, остатки, заказы и динамику продаж.",
		FallbackReasonText:   "Продавец давно не заходил в кабинет",
	},
	"OUT_OF_STOCK": {
		RuleID:               4,
		RecommendationTypeID: 5,
		FallbackTitle:        "Пополните остатки",
		FallbackDescription:  "У части товаров закончились остатки. Пополните их, чтобы не терять продажи и показы.",
		FallbackReasonText:   "У части товаров отсутствуют остатки",
	},

	"LOW_ASSORTMENT": {
		RuleID:               5,
		RecommendationTypeID: 1,
		FallbackTitle:        "Расширьте ассортимент",
		FallbackDescription:  "Текущий ассортимент выглядит слишком узким для устойчивого роста продаж. Добавьте новые товарные позиции в близких категориях.",
		FallbackReasonText:   "Ассортимент продавца слишком узкий",
	},
	"NARROW_CATEGORY_FOCUS": {
		RuleID:               6,
		RecommendationTypeID: 1,
		FallbackTitle:        "Расширьте категории продаж",
		FallbackDescription:  "Продавец работает слишком узко по категориям. Попробуйте добавить товары в смежные категории, чтобы увеличить охват аудитории.",
		FallbackReasonText:   "Пониженная диверсификация по категориям",
	},
	"LOW_REGION_COVERAGE": {
		RuleID:               7,
		RecommendationTypeID: 1,
		FallbackTitle:        "Расширьте географию продаж",
		FallbackDescription:  "Товары представлены в слишком малом числе регионов. Проверьте доступность отгрузок и расширьте географию продаж там, где это возможно.",
		FallbackReasonText:   "Слишком низкое покрытие регионов",
	},
	"LOW_PUBLISHED_SHARE": {
		RuleID:               8,
		RecommendationTypeID: 1,
		FallbackTitle:        "Опубликуйте больше товаров",
		FallbackDescription:  "Часть товаров создана, но не опубликована. Завершите публикацию карточек, чтобы увеличить видимость ассортимента.",
		FallbackReasonText:   "Низкая доля опубликованных товаров",
	},
	"HIGH_STOCKOUT_RATE": {
		RuleID:               9,
		RecommendationTypeID: 5,
		FallbackTitle:        "Снизьте долю товаров без остатков",
		FallbackDescription:  "Слишком много товаров находится без остатков. Это ограничивает продажи и ухудшает доступность ассортимента.",
		FallbackReasonText:   "Высокая доля товаров без остатков",
	},
	"LOW_ORDER_VOLUME": {
		RuleID:               10,
		RecommendationTypeID: 3,
		FallbackTitle:        "Увеличьте количество заказов",
		FallbackDescription:  "За последний период объём заказов низкий. Проверьте карточки товаров, цены, акции и источники трафика.",
		FallbackReasonText:   "Низкий объем заказов",
	},
	"LOW_REVENUE": {
		RuleID:               11,
		RecommendationTypeID: 3,
		FallbackTitle:        "Увеличьте выручку",
		FallbackDescription:  "Выручка за период выглядит недостаточной. Проверьте ассортимент, ценовую политику и конверсию карточек.",
		FallbackReasonText:   "Низкая выручка за период",
	},
	"LOW_MARGIN": {
		RuleID:               12,
		RecommendationTypeID: 4,
		FallbackTitle:        "Повысите маржинальность",
		FallbackDescription:  "Маржинальность выглядит слабой. Стоит пересмотреть цены, скидки, себестоимость и структуру ассортимента.",
		FallbackReasonText:   "Низкая маржа за период",
	},
	"NO_RECENT_ACTIVITY": {
		RuleID:               13,
		RecommendationTypeID: 6,
		FallbackTitle:        "Повышайте операционную активность",
		FallbackDescription:  "В кабинете давно не было заметной активности. Проверьте заказы, остатки, карточки товаров и актуальность ассортимента.",
		FallbackReasonText:   "Длительное отсутствие активности",
	},
	"STALE_CATALOG": {
		RuleID:               14,
		RecommendationTypeID: 1,
		FallbackTitle:        "Обновите каталог",
		FallbackDescription:  "Каталог выглядит устаревшим или слабо обновляемым. Добавьте новые товары и пересмотрите актуальность старых позиций.",
		FallbackReasonText:   "Каталог давно не обновлялся",
	},
	"LOW_CATEGORIES_COUNT": {
		RuleID:               15,
		RecommendationTypeID: 1,
		FallbackTitle:        "Добавьте больше категорий",
		FallbackDescription:  "Продавец работает в слишком малом числе категорий. Расширение категорий может улучшить охват и снизить зависимость от одного сегмента.",
		FallbackReasonText:   "Слишком малое число категорий",
	},
	"LOW_ACTIVE_CATEGORIES": {
		RuleID:               16,
		RecommendationTypeID: 1,
		FallbackTitle:        "Активируйте категории",
		FallbackDescription:  "Не все категории, в которых есть товары, сейчас дают полноценную активность. Проверьте наполнение и публикацию по категориям.",
		FallbackReasonText:   "Недостаточно активных категорий",
	},
	"WEAK_7D_DYNAMICS": {
		RuleID:               17,
		RecommendationTypeID: 3,
		FallbackTitle:        "Улучшите краткосрочную динамику",
		FallbackDescription:  "Последние 7 дней показывают слабую динамику. Проверьте ассортимент, остатки и краткосрочные действия для роста продаж.",
		FallbackReasonText:   "Слабая динамика за 7 дней",
	},
	"WEAK_30D_DYNAMICS": {
		RuleID:               18,
		RecommendationTypeID: 3,
		FallbackTitle:        "Улучшите динамику за месяц",
		FallbackDescription:  "По итогам 30 дней динамика продаж выглядит слабой. Стоит пересмотреть ассортимент, карточки и ценовую стратегию.",
		FallbackReasonText:   "Слабая динамика за 30 дней",
	},
}

// evaluateWithAI calls Yandex AI and maps the response to []Match.
func (e *Engine) evaluateWithAI(snapshot domain.SellerMetricsSnapshot) ([]Match, error) {
	if e.apiKey == "" {
		return nil, fmt.Errorf("YANDEX_API_KEY is empty")
	}
	if e.folderID == "" {
		return nil, fmt.Errorf("YANDEX_FOLDER_ID is empty")
	}
	if e.promptID == "" {
		return nil, fmt.Errorf("YANDEX_PROMPT_ID is empty")
	}

	payload, err := json.Marshal(snapshot)
	if err != nil {
		return nil, fmt.Errorf("marshal snapshot: %w", err)
	}

	reqBody := promptRequest{
		Prompt: promptRef{ID: e.promptID},
		Input:  string(payload),
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, yandexResponsesURL, bytes.NewReader(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Api-Key "+e.apiKey)
	req.Header.Set("OpenAI-Project", e.folderID)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("yandex ai returned %s: %s", resp.Status, string(body))
	}

	text, err := extractModelText(body)
	if err != nil {
		return nil, err
	}

	parsed, err := parseAIText(text)
	if err != nil {
		return nil, err
	}

	return normalizeMatches(parsed), nil
}

// evaluateLocal is the exact old fallback logic.
func (e *Engine) evaluateLocal(snapshot domain.SellerMetricsSnapshot) []Match {
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

func extractModelText(body []byte) (string, error) {
	var resp yandexResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("unmarshal ai response: %w", err)
	}

	if resp.Error != nil && (resp.Error.Code != "" || resp.Error.Message != "") {
		return "", fmt.Errorf("ai error %s: %s", resp.Error.Code, resp.Error.Message)
	}

	if len(resp.Output) > 0 {
		for _, item := range resp.Output {
			for _, c := range item.Content {
				if strings.TrimSpace(c.Text) != "" {
					return c.Text, nil
				}
			}
		}
	}

	return "", fmt.Errorf("empty ai output")
}

func parseAIText(text string) ([]aiMatch, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("empty model text")
	}

	// 1) Preferred format: JSON {"matches":[...]}
	if matches, err := parseJSONEnvelope(text); err == nil {
		return matches, nil
	}

	// 2) Plain text fallback:
	// Match:
	// TriggerCode: ...
	// Title: ...
	// Description: ...
	// ReasonText: ...
	matches := parseLegacyPlainText(text)
	if len(matches) > 0 {
		return matches, nil
	}

	return nil, fmt.Errorf("unable to parse model output")
}

func parseJSONEnvelope(text string) ([]aiMatch, error) {
	trimmed := strings.TrimSpace(text)

	// Direct JSON.
	var env aiMatchesEnvelope
	if err := json.Unmarshal([]byte(trimmed), &env); err == nil {
		return env.Matches, nil
	}

	// JSON inside a larger text blob.
	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start >= 0 && end > start {
		candidate := trimmed[start : end+1]
		if err := json.Unmarshal([]byte(candidate), &env); err == nil {
			return env.Matches, nil
		}
	}

	return nil, fmt.Errorf("not json")
}

func parseLegacyPlainText(text string) []aiMatch {
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	normalized = strings.TrimSpace(normalized)

	parts := strings.Split(normalized, "\n\nMatch:")
	if len(parts) == 0 {
		return nil
	}

	matches := make([]aiMatch, 0, len(parts))

	for _, part := range parts {
		block := strings.TrimSpace(part)
		block = strings.TrimPrefix(block, "Match:")
		block = strings.TrimSpace(block)

		if block == "" {
			continue
		}

		lines := splitNonEmptyLines(block)
		m := aiMatch{
			TriggerCode: valueAfterPrefix(lines, "TriggerCode:"),
			Title:       valueAfterPrefix(lines, "Title:"),
			Description: valueAfterPrefix(lines, "Description:"),
			ReasonText:  valueAfterPrefix(lines, "ReasonText:"),
		}

		if strings.TrimSpace(m.TriggerCode) != "" {
			matches = append(matches, m)
		}
	}

	return matches
}

func splitNonEmptyLines(s string) []string {
	raw := strings.Split(s, "\n")
	out := make([]string, 0, len(raw))
	for _, line := range raw {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func valueAfterPrefix(lines []string, prefix string) string {
	for _, line := range lines {
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}

func normalizeMatches(items []aiMatch) []Match {
	matches := make([]Match, 0, len(items))

	seen := make(map[string]struct{}, len(items))

	for _, item := range items {
		code := strings.ToUpper(strings.TrimSpace(item.TriggerCode))
		meta, ok := allowedRules[code]
		if !ok {
			continue
		}

		if _, exists := seen[code]; exists {
			continue
		}
		seen[code] = struct{}{}

		title := strings.TrimSpace(item.Title)
		if title == "" {
			title = meta.FallbackTitle
		}

		description := strings.TrimSpace(item.Description)
		if description == "" {
			description = meta.FallbackDescription
		}

		reasonText := strings.TrimSpace(item.ReasonText)
		if reasonText == "" {
			reasonText = meta.FallbackReasonText
		}

		matches = append(matches, Match{
			TriggerCode:          code,
			RuleID:               meta.RuleID,
			RecommendationTypeID: meta.RecommendationTypeID,
			Title:                title,
			Description:          description,
			ReasonText:           reasonText,
		})
	}

	return matches
}

func envString(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}