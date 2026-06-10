package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"wb-marketplace/internal/domain"
	"wb-marketplace/internal/rules"
)

// AIDemoHandler serves a simple demo page and a direct recommendation preview endpoint.
type AIDemoHandler struct {
	engine *rules.Engine
}

// NewAIDemoHandler creates a demo handler.
func NewAIDemoHandler() *AIDemoHandler {
	return &AIDemoHandler{
		engine: rules.New(),
	}
}

type aiDemoRequest struct {
	SellerID                  int64   `json:"seller_id"`
	ActiveProductsCount       int     `json:"active_products_count"`
	PublishedProductsCount    int     `json:"published_products_count"`
	ProductsWithoutStockCount int     `json:"products_without_stock_count"`
	CategoriesCount           int     `json:"categories_count"`
	ActiveCategoriesCount     int     `json:"active_categories_count"`
	RegionsCount              int     `json:"regions_count"`
	Orders7d                  int     `json:"orders_7d"`
	Orders30d                 int     `json:"orders_30d"`
	Revenue7d                 float64 `json:"revenue_7d"`
	Revenue30d                float64 `json:"revenue_30d"`
	Margin30d                 float64 `json:"margin_30d"`
	LastLoginDays             int     `json:"last_login_days"`
	NoSalesDays               int     `json:"no_sales_days"`
	CurrentPrimaryModelCode   *string `json:"current_primary_model_code,omitempty"`
}

type aiDemoMatch struct {
	TriggerCode          string `json:"trigger_code"`
	RuleID               int64  `json:"rule_id"`
	RecommendationTypeID int64  `json:"recommendation_type_id"`
	Title                string `json:"title"`
	Description          string `json:"description"`
	ReasonText           string `json:"reason_text"`
}

type aiDemoResponse struct {
	Matches []aiDemoMatch `json:"matches"`
}

const demoPageHTML = `<!DOCTYPE html>
<html lang="ru">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>AI Recommendation Demo</title>
	<style>
		:root { color-scheme: light; }
		body {
			margin: 0;
			padding: 24px;
			font-family: Arial, sans-serif;
			background: #f3ede4;
			color: #1f2937;
		}
		.container {
			max-width: 1100px;
			margin: 0 auto;
		}
		h1 {
			margin: 0 0 8px 0;
			font-size: 28px;
		}
		p {
			margin: 0 0 20px 0;
			color: #4b5563;
		}
		.grid {
			display: grid;
			grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
			gap: 12px;
			margin-bottom: 16px;
		}
		.field {
			background: #fff;
			border: 1px solid #e5e7eb;
			border-radius: 14px;
			padding: 12px;
			box-shadow: 0 1px 2px rgba(0,0,0,.03);
		}
		label {
			display: block;
			font-size: 13px;
			color: #374151;
			margin-bottom: 6px;
		}
		input {
			width: 100%;
			box-sizing: border-box;
			border: 1px solid #d1d5db;
			border-radius: 10px;
			padding: 10px 12px;
			font-size: 14px;
			background: #fff;
		}
		input:focus {
			outline: none;
			border-color: #ae2573;
			box-shadow: 0 0 0 3px rgba(99,102,241,.12);
		}
		.actions {
			display: flex;
			gap: 12px;
			margin: 16px 0 22px 0;
			flex-wrap: wrap;
		}
		button {
			border: 0;
			border-radius: 12px;
			padding: 12px 16px;
			cursor: pointer;
			font-size: 14px;
			font-weight: 600;
		}
		.primary {
			background: #ae2573;
			color: #fff;
		}
		.secondary {
			background: #e5e7eb;
			color: #111827;
		}
		.banner {
			margin: 0 0 16px 0;
			padding: 14px 16px;
			border-radius: 14px;
			border: 1px solid transparent;
		}
		.banner.ok {
			background: #ecfdf5;
			border-color: #a7f3d0;
			color: #065f46;
		}
		.banner.error {
			background: #fef2f2;
			border-color: #fecaca;
			color: #991b1b;
		}
		.results {
			display: grid;
			grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
			gap: 12px;
		}
		.card {
			background: #fff;
			border: 1px solid #e5e7eb;
			border-radius: 16px;
			padding: 16px;
			box-shadow: 0 2px 4px rgba(0,0,0,.04);
		}
		.badge {
			display: inline-block;
			font-size: 12px;
			font-weight: 700;
			background: #eef2ff;
			color: #ae2573;
			padding: 5px 8px;
			border-radius: 999px;
			margin-bottom: 10px;
		}
		.card h3 {
			margin: 0 0 8px 0;
			font-size: 18px;
		}
		.card .meta {
			font-size: 13px;
			color: #6b7280;
			margin-top: 10px;
		}
		.code {
			background: #0f172a;
			color: #e2e8f0;
			padding: 12px;
			border-radius: 12px;
			overflow: auto;
			font-size: 13px;
			line-height: 1.45;
		}
	</style>
</head>
<body>
	<div class="container">
		<h1>AI Recommendation Demo</h1>
		<p>Заполни метрики продавца и нажми кнопку. Ответ придёт через текущий rules.Engine.</p>

		<div class="grid">
			<div class="field">
				<label for="seller_id">Seller ID</label>
				<input id="seller_id" type="number" value="10">
			</div>
			<div class="field">
				<label for="active_products_count">Active Products Count</label>
				<input id="active_products_count" type="number" value="0">
			</div>
			<div class="field">
				<label for="published_products_count">Published Products Count</label>
				<input id="published_products_count" type="number" value="0">
			</div>
			<div class="field">
				<label for="products_without_stock_count">Products Without Stock Count</label>
				<input id="products_without_stock_count" type="number" value="0">
			</div>
			<div class="field">
				<label for="categories_count">Categories Count</label>
				<input id="categories_count" type="number" value="1">
			</div>
			<div class="field">
				<label for="active_categories_count">Active Categories Count</label>
				<input id="active_categories_count" type="number" value="1">
			</div>
			<div class="field">
				<label for="regions_count">Regions Count</label>
				<input id="regions_count" type="number" value="1">
			</div>
			<div class="field">
				<label for="orders_7d">Orders 7d</label>
				<input id="orders_7d" type="number" value="0">
			</div>
			<div class="field">
				<label for="orders_30d">Orders 30d</label>
				<input id="orders_30d" type="number" value="0">
			</div>
			<div class="field">
				<label for="revenue_7d">Revenue 7d</label>
				<input id="revenue_7d" type="number" step="0.01" value="0">
			</div>
			<div class="field">
				<label for="revenue_30d">Revenue 30d</label>
				<input id="revenue_30d" type="number" step="0.01" value="0">
			</div>
			<div class="field">
				<label for="margin_30d">Margin 30d</label>
				<input id="margin_30d" type="number" step="0.01" value="0">
			</div>
			<div class="field">
				<label for="last_login_days">Last Login Days</label>
				<input id="last_login_days" type="number" value="10">
			</div>
			<div class="field">
				<label for="no_sales_days">No Sales Days</label>
				<input id="no_sales_days" type="number" value="30">
			</div>
			<div class="field">
				<label for="current_primary_model_code">Current Primary Model Code</label>
				<input id="current_primary_model_code" type="text" placeholder="optional">
			</div>
		</div>

		<div class="actions">
			<button class="primary" type="button" onclick="generateRecommendation()">Generate recommendation</button>
			<button class="secondary" type="button" onclick="fillExample()">Fill example</button>
			<button class="secondary" type="button" onclick="clearResults()">Clear output</button>
		</div>

		<div id="banner"></div>
		<div id="results" class="results"></div>
		<pre id="raw" class="code" style="display:none;"></pre>
	</div>

	<script>
		function num(id) {
			var value = document.getElementById(id).value;
			if (value === '' || value === null || value === undefined) {
				return 0;
			}
			return Number(value);
		}

		function text(id) {
			var value = document.getElementById(id).value;
			return value ? String(value).trim() : '';
		}

		function escapeHtml(value) {
			return String(value)
				.replace(/&/g, '&amp;')
				.replace(/</g, '&lt;')
				.replace(/>/g, '&gt;')
				.replace(/"/g, '&quot;')
				.replace(/'/g, '&#39;');
		}

		function setBanner(message, kind) {
			var banner = document.getElementById('banner');
			banner.className = 'banner ' + kind;
			banner.textContent = message;
		}

		function clearResults() {
			document.getElementById('banner').innerHTML = '';
			document.getElementById('results').innerHTML = '';
			document.getElementById('raw').style.display = 'none';
			document.getElementById('raw').textContent = '';
		}

		function fillExample() {
			document.getElementById('seller_id').value = 10;
			document.getElementById('active_products_count').value = 0;
			document.getElementById('published_products_count').value = 0;
			document.getElementById('products_without_stock_count').value = 2;
			document.getElementById('categories_count').value = 1;
			document.getElementById('active_categories_count').value = 1;
			document.getElementById('regions_count').value = 1;
			document.getElementById('orders_7d').value = 0;
			document.getElementById('orders_30d').value = 0;
			document.getElementById('revenue_7d').value = 0;
			document.getElementById('revenue_30d').value = 0;
			document.getElementById('margin_30d').value = 0;
			document.getElementById('last_login_days').value = 10;
			document.getElementById('no_sales_days').value = 30;
			document.getElementById('current_primary_model_code').value = '';
			clearResults();
		}

		async function generateRecommendation() {
			clearResults();

			var payload = {
				seller_id: num('seller_id'),
				active_products_count: num('active_products_count'),
				published_products_count: num('published_products_count'),
				products_without_stock_count: num('products_without_stock_count'),
				categories_count: num('categories_count'),
				active_categories_count: num('active_categories_count'),
				regions_count: num('regions_count'),
				orders_7d: num('orders_7d'),
				orders_30d: num('orders_30d'),
				revenue_7d: num('revenue_7d'),
				revenue_30d: num('revenue_30d'),
				margin_30d: num('margin_30d'),
				last_login_days: num('last_login_days'),
				no_sales_days: num('no_sales_days'),
				current_primary_model_code: text('current_primary_model_code') || null
			};

			setBanner('Запрос отправлен...', 'ok');

			var response;
			var data;
			try {
				response = await fetch('/api/v1/admin/generate-ai-recommendation', {
					method: 'POST',
					headers: {
						'Content-Type': 'application/json'
					},
					body: JSON.stringify(payload)
				});
				data = await response.json();
			} catch (err) {
				setBanner('Ошибка запроса: ' + err.message, 'error');
				return;
			}

			document.getElementById('raw').style.display = 'block';
			document.getElementById('raw').textContent = JSON.stringify(data, null, 2);

			if (!response.ok) {
				setBanner((data && data.error) ? data.error : 'Request failed', 'error');
				return;
			}

			var matches = data.matches || [];
			if (matches.length === 0) {
				setBanner('Критичных отклонений не обнаружено.', 'ok');
				document.getElementById('results').innerHTML = '';
				return;
			}

			setBanner('Найдено рекомендаций: ' + matches.length, 'ok');

			var html = '';
			for (var i = 0; i < matches.length; i++) {
				var m = matches[i];
				html += ''
					+ '<div class="card">'
					+   '<div class="badge">' + escapeHtml(m.trigger_code || '') + '</div>'
					+   '<h3>' + escapeHtml(m.title || '') + '</h3>'
					+   '<div>' + escapeHtml(m.description || '') + '</div>'
					+   '<div class="meta"><strong>Reason:</strong> ' + escapeHtml(m.reason_text || '') + '</div>'
					+ '</div>';
			}
			document.getElementById('results').innerHTML = html;
		}
	</script>
</body>
</html>`
	
// DemoPage serves the test page.
func (h *AIDemoHandler) DemoPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(demoPageHTML))
}

// Generate handles POST /api/v1/admin/generate-ai-recommendation.
func (h *AIDemoHandler) Generate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req aiDemoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	snapshot := domain.SellerMetricsSnapshot{
		SellerID:                  req.SellerID,
		SnapshotDate:              time.Now().UTC(),
		ActiveProductsCount:       req.ActiveProductsCount,
		PublishedProductsCount:    req.PublishedProductsCount,
		ProductsWithoutStockCount: req.ProductsWithoutStockCount,
		CategoriesCount:           req.CategoriesCount,
		ActiveCategoriesCount:     req.ActiveCategoriesCount,
		RegionsCount:              req.RegionsCount,
		Orders7d:                  req.Orders7d,
		Orders30d:                 req.Orders30d,
		Revenue7d:                 req.Revenue7d,
		Revenue30d:                req.Revenue30d,
		Margin30d:                 req.Margin30d,
		LastLoginDays:             req.LastLoginDays,
		NoSalesDays:               req.NoSalesDays,
		CurrentPrimaryModelCode:   req.CurrentPrimaryModelCode,
	}

	matches := h.engine.EvaluateSnapshot(snapshot)

	out := make([]aiDemoMatch, 0, len(matches))
	for _, m := range matches {
		out = append(out, aiDemoMatch{
			TriggerCode:          m.TriggerCode,
			RuleID:               m.RuleID,
			RecommendationTypeID: m.RecommendationTypeID,
			Title:                m.Title,
			Description:          m.Description,
			ReasonText:           m.ReasonText,
		})
	}

	writeJSON(w, http.StatusOK, aiDemoResponse{
		Matches: out,
	})
}