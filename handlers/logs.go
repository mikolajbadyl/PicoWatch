package handlers

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"picowatch/pricing"
)

type LogEntry struct {
	ID           int64     `json:"id"`
	Model        string    `json:"model"`
	Provider     string    `json:"provider"`
	InputTokens  int       `json:"input_tokens"`
	OutputTokens int       `json:"output_tokens"`
	TotalTokens  int       `json:"total_tokens"`
	Cost         float64   `json:"cost"`
	DurationMs   int       `json:"duration_ms"`
	Status       string    `json:"status"`
	Error        string    `json:"error"`
	Metadata     string    `json:"metadata"`
	CreatedAt    time.Time `json:"created_at"`
}

func (h *Handler) LogsList(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 50
	}
	model := r.URL.Query().Get("model")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	offset := (page - 1) * limit

	sortCol := r.URL.Query().Get("sort")
	sortDir := r.URL.Query().Get("dir")
	allowed := map[string]string{
		"created_at":   "created_at",
		"model":        "model",
		"input_tokens": "input_tokens",
		"output_tokens": "output_tokens",
		"total_tokens": "input_tokens + output_tokens",
		"cost":         "cost",
		"duration_ms":  "duration_ms",
	}
	orderExpr, ok := allowed[sortCol]
	if !ok {
		orderExpr = "created_at"
	}
	if sortDir != "asc" {
		sortDir = "desc"
	}

	baseQuery := "FROM llm_logs"
	var filterArgs []any
	var conditions []string
	if model != "" {
		conditions = append(conditions, "model LIKE ?")
		filterArgs = append(filterArgs, "%"+model+"%")
	}
	if from != "" {
		conditions = append(conditions, "created_at >= ?")
		filterArgs = append(filterArgs, from)
	}
	if to != "" {
		conditions = append(conditions, "created_at <= ?")
		filterArgs = append(filterArgs, to+" 23:59:59")
	}
	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	h.db.QueryRow("SELECT COUNT(*) "+baseQuery, filterArgs...).Scan(&total)

	query := "SELECT id, model, provider, input_tokens, output_tokens, input_tokens + output_tokens, cost, duration_ms, status, error, metadata, created_at " + baseQuery
	query += " ORDER BY " + orderExpr + " " + sortDir + " LIMIT ? OFFSET ?"
	args := append(filterArgs, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	logs := []LogEntry{}
	for rows.Next() {
		var l LogEntry
		if err := rows.Scan(&l.ID, &l.Model, &l.Provider, &l.InputTokens, &l.OutputTokens,
			&l.TotalTokens, &l.Cost, &l.DurationMs, &l.Status, &l.Error,
			&l.Metadata, &l.CreatedAt); err != nil {
			continue
		}
		logs = append(logs, l)
	}

	h.writeJSON(w, http.StatusOK, map[string]any{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

type createLogRequest struct {
	Model        string  `json:"model"`
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	Cost         float64 `json:"cost"`
	DurationMs   int     `json:"duration_ms"`
	Status       string  `json:"status"`
	Error        string  `json:"error"`
	Metadata     any     `json:"metadata"`
}

func (h *Handler) LogsCreate(w http.ResponseWriter, r *http.Request) {
	var req createLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Model == "" {
		h.writeError(w, http.StatusBadRequest, "model is required")
		return
	}

	if req.Status == "" {
		req.Status = "success"
	}

	if req.Cost == 0 && req.InputTokens+req.OutputTokens > 0 {
		if cost, ok := pricing.Calculate(h.db, req.Model, req.InputTokens, req.OutputTokens); ok {
			req.Cost = cost
		}
	}

	provider := resolveProvider(h.db, req.Model)

	metadata := "{}"
	if req.Metadata != nil {
		if b, err := json.Marshal(req.Metadata); err == nil {
			metadata = string(b)
		}
	}

	result, err := h.db.Exec(
		`INSERT INTO llm_logs (model, provider, input_tokens, output_tokens, cost, duration_ms, status, error, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		req.Model, provider, req.InputTokens, req.OutputTokens, req.Cost,
		req.DurationMs, req.Status, req.Error, metadata,
	)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to create log")
		return
	}

	id, _ := result.LastInsertId()
	h.writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func resolveProvider(db *sql.DB, model string) string {
	if idx := strings.Index(model, "/"); idx > 0 {
		return model[:idx]
	}
	var fullID string
	err := db.QueryRow(
		"SELECT model_id FROM model_pricing WHERE model_id = ? OR model_id LIKE ? OR LOWER(model_id) LIKE ? LIMIT 1",
		model, "%/"+model, "%"+strings.ToLower(model)+"%",
	).Scan(&fullID)
	if err != nil {
		return ""
	}
	if idx := strings.Index(fullID, "/"); idx > 0 {
		return fullID[:idx]
	}
	return ""
}

func (h *Handler) LogsDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	h.db.Exec("DELETE FROM llm_logs WHERE id = ?", id)
	h.writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) LogsExport(w http.ResponseWriter, r *http.Request) {
	model := r.URL.Query().Get("model")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	var conditions []string
	var args []any
	if model != "" {
		conditions = append(conditions, "model LIKE ?")
		args = append(args, "%"+model+"%")
	}
	if from != "" {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, from)
	}
	if to != "" {
		conditions = append(conditions, "created_at <= ?")
		args = append(args, to+" 23:59:59")
	}

	query := "SELECT id, model, input_tokens, output_tokens, input_tokens+output_tokens, cost, status, error, metadata, created_at FROM llm_logs"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY created_at DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=\"logs.csv\"")

	cw := csv.NewWriter(w)
	cw.Write([]string{"id", "model", "input_tokens", "output_tokens", "total_tokens", "cost", "status", "error", "metadata", "created_at"})

	for rows.Next() {
		var l LogEntry
		rows.Scan(&l.ID, &l.Model, &l.InputTokens, &l.OutputTokens, &l.TotalTokens, &l.Cost, &l.Status, &l.Error, &l.Metadata, &l.CreatedAt)
		cw.Write([]string{
			fmt.Sprintf("%d", l.ID),
			l.Model,
			fmt.Sprintf("%d", l.InputTokens),
			fmt.Sprintf("%d", l.OutputTokens),
			fmt.Sprintf("%d", l.TotalTokens),
			fmt.Sprintf("%.8f", l.Cost),
			l.Status,
			l.Error,
			l.Metadata,
			l.CreatedAt.Format(time.RFC3339),
		})
	}
	cw.Flush()
}
