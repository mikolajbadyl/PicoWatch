package handlers

import (
	"net/http"
)

type Stats struct {
	TotalRequests     int     `json:"total_requests"`
	TotalInputTokens  int64   `json:"total_input_tokens"`
	TotalOutputTokens int64   `json:"total_output_tokens"`
	TotalTokens       int64   `json:"total_tokens"`
	TotalCost         float64 `json:"total_cost"`
	CostToday         float64 `json:"cost_today"`
	CostMonth         float64 `json:"cost_month"`
	AvgLatencyMs      float64 `json:"avg_latency_ms"`
	RequestsToday     int     `json:"requests_today"`
	SuccessRate       float64 `json:"success_rate"`
}

func (h *Handler) StatsGet(w http.ResponseWriter, r *http.Request) {
	var stats Stats
	h.db.QueryRow(`
		SELECT
			COUNT(*),
			COALESCE(SUM(input_tokens), 0),
			COALESCE(SUM(output_tokens), 0),
			COALESCE(SUM(input_tokens + output_tokens), 0),
			COALESCE(SUM(cost), 0),
			COALESCE(AVG(duration_ms), 0)
		FROM llm_logs
	`).Scan(
		&stats.TotalRequests,
		&stats.TotalInputTokens,
		&stats.TotalOutputTokens,
		&stats.TotalTokens,
		&stats.TotalCost,
		&stats.AvgLatencyMs,
	)

	h.db.QueryRow(`SELECT COUNT(*) FROM llm_logs WHERE date(created_at) = date('now')`).Scan(&stats.RequestsToday)
	h.db.QueryRow(`SELECT COALESCE(SUM(cost),0) FROM llm_logs WHERE date(created_at) = date('now')`).Scan(&stats.CostToday)
	h.db.QueryRow(`SELECT COALESCE(SUM(cost),0) FROM llm_logs WHERE strftime('%Y-%m', created_at) = strftime('%Y-%m', 'now')`).Scan(&stats.CostMonth)

	var total, successful int
	h.db.QueryRow("SELECT COUNT(*), COUNT(CASE WHEN status = 'success' THEN 1 END) FROM llm_logs").
		Scan(&total, &successful)
	if total > 0 {
		stats.SuccessRate = float64(successful) / float64(total) * 100
	}

	h.writeJSON(w, http.StatusOK, stats)
}

type ModelStat struct {
	Model        string  `json:"model"`
	Count        int     `json:"count"`
	InputTokens  int64   `json:"input_tokens"`
	OutputTokens int64   `json:"output_tokens"`
	TotalCost    float64 `json:"total_cost"`
	AvgLatency   float64 `json:"avg_latency_ms"`
}

func (h *Handler) StatsModels(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT model, COUNT(*),
			COALESCE(SUM(input_tokens), 0),
			COALESCE(SUM(output_tokens), 0),
			COALESCE(SUM(cost), 0),
			COALESCE(AVG(duration_ms), 0)
		FROM llm_logs
		GROUP BY model
		ORDER BY COUNT(*) DESC
	`)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	models := []ModelStat{}
	for rows.Next() {
		var m ModelStat
		rows.Scan(&m.Model, &m.Count, &m.InputTokens, &m.OutputTokens, &m.TotalCost, &m.AvgLatency)
		models = append(models, m)
	}

	h.writeJSON(w, http.StatusOK, models)
}

type DailyStat struct {
	Date   string  `json:"date"`
	Count  int     `json:"count"`
	Tokens int64   `json:"tokens"`
	Cost   float64 `json:"cost"`
}

func (h *Handler) StatsDaily(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT
			date(created_at) as day,
			COUNT(*),
			COALESCE(SUM(input_tokens + output_tokens), 0),
			COALESCE(SUM(cost), 0)
		FROM llm_logs
		WHERE created_at >= date('now', '-30 days')
		GROUP BY day
		ORDER BY day ASC
	`)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	stats := []DailyStat{}
	for rows.Next() {
		var s DailyStat
		rows.Scan(&s.Date, &s.Count, &s.Tokens, &s.Cost)
		stats = append(stats, s)
	}

	h.writeJSON(w, http.StatusOK, stats)
}
