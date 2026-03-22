package pricing

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type modelInfo struct {
	ID      string `json:"id"`
	Pricing struct {
		Prompt     string `json:"prompt"`
		Completion string `json:"completion"`
	} `json:"pricing"`
}

func Sync(db *sql.DB) error {
	resp, err := http.Get("https://openrouter.ai/api/v1/models")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Data []modelInfo `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, m := range result.Data {
		prompt, _ := strconv.ParseFloat(m.Pricing.Prompt, 64)
		completion, _ := strconv.ParseFloat(m.Pricing.Completion, 64)

		tx.Exec(`
			INSERT INTO model_pricing (model_id, prompt_price, completion_price, updated_at)
			VALUES (?, ?, ?, CURRENT_TIMESTAMP)
			ON CONFLICT(model_id) DO UPDATE SET
				prompt_price = excluded.prompt_price,
				completion_price = excluded.completion_price,
				updated_at = excluded.updated_at
		`, m.ID, prompt, completion)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("Pricing synced: %d models from OpenRouter", len(result.Data))
	return nil
}

func StartSync(db *sql.DB) {
	go func() {
		if err := Sync(db); err != nil {
			log.Printf("Pricing sync failed: %v", err)
		}
		ticker := time.NewTicker(24 * time.Hour)
		for range ticker.C {
			if err := Sync(db); err != nil {
				log.Printf("Pricing refresh failed: %v", err)
			}
		}
	}()
}

func Calculate(db *sql.DB, modelID string, inputTokens, outputTokens int) (float64, bool) {
	var prompt, completion float64

	err := db.QueryRow(
		"SELECT prompt_price, completion_price FROM model_pricing WHERE model_id = ?",
		modelID,
	).Scan(&prompt, &completion)

	if err != nil {
		err = db.QueryRow(
			"SELECT prompt_price, completion_price FROM model_pricing WHERE model_id = ? OR model_id LIKE ?",
			modelID, "%/"+modelID,
		).Scan(&prompt, &completion)
	}

	if err != nil {
		needle := strings.ToLower(modelID)
		err = db.QueryRow(
			"SELECT prompt_price, completion_price FROM model_pricing WHERE LOWER(model_id) LIKE ? LIMIT 1",
			"%"+needle+"%",
		).Scan(&prompt, &completion)
	}

	if err != nil {
		return 0, false
	}

	cost := float64(inputTokens)*prompt + float64(outputTokens)*completion
	return cost, true
}
