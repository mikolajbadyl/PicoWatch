package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
	"picowatch/db"
	"picowatch/pricing"
)

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *Handler) UsersList(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("SELECT id, email, created_at FROM users ORDER BY created_at ASC")
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var u User
		rows.Scan(&u.ID, &u.Email, &u.CreatedAt)
		users = append(users, u)
	}
	h.writeJSON(w, http.StatusOK, users)
}

func (h *Handler) UsersCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		h.writeError(w, http.StatusBadRequest, "Email and password are required")
		return
	}
	if len(req.Password) < 8 {
		h.writeError(w, http.StatusBadRequest, "Password must be at least 8 characters")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	result, err := h.db.Exec(
		"INSERT INTO users (email, password_hash) VALUES (?, ?)",
		req.Email, string(hash),
	)
	if err != nil {
		h.writeError(w, http.StatusConflict, "Email already exists")
		return
	}

	id, _ := result.LastInsertId()
	h.writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func (h *Handler) UsersDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	callerID := r.Context().Value(userIDKey).(int64)
	var targetID int64
	h.db.QueryRow("SELECT id FROM users WHERE id = ?", id).Scan(&targetID)
	if targetID == callerID {
		h.writeError(w, http.StatusBadRequest, "Cannot delete your own account")
		return
	}

	var exists int
	h.db.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", id).Scan(&exists)
	if exists == 0 {
		h.writeError(w, http.StatusNotFound, "User not found")
		return
	}

	h.db.Exec("DELETE FROM api_keys WHERE user_id = ?", id)
	h.db.Exec("DELETE FROM users WHERE id = ?", id)
	h.writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) UsersChangePassword(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if len(req.Password) < 8 {
		h.writeError(w, http.StatusBadRequest, "Password must be at least 8 characters")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	res, err := h.db.Exec("UPDATE users SET password_hash = ? WHERE id = ?", string(hash), id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Database error")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		h.writeError(w, http.StatusNotFound, "User not found")
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

type APIKey struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	KeyPreview string     `json:"key_preview"`
	UserID     int64      `json:"user_id"`
	UserEmail  string     `json:"user_email"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
}

func (h *Handler) APIKeysList(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT ak.id, ak.name, ak.key, ak.user_id, u.email, ak.created_at, ak.last_used_at
		FROM api_keys ak
		JOIN users u ON u.id = ak.user_id
		ORDER BY ak.created_at DESC
	`)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	keys := []APIKey{}
	for rows.Next() {
		var k APIKey
		var fullKey string
		rows.Scan(&k.ID, &k.Name, &fullKey, &k.UserID, &k.UserEmail, &k.CreatedAt, &k.LastUsedAt)
		if len(fullKey) > 12 {
			k.KeyPreview = fullKey[:12] + "••••••••"
		} else {
			k.KeyPreview = fullKey
		}
		keys = append(keys, k)
	}
	h.writeJSON(w, http.StatusOK, keys)
}

func (h *Handler) APIKeysCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Name == "" {
		req.Name = "API Key"
	}

	callerID := r.Context().Value(userIDKey).(int64)
	apiKey := db.GenerateAPIKey()

	result, err := h.db.Exec(
		"INSERT INTO api_keys (name, key, user_id) VALUES (?, ?, ?)",
		req.Name, apiKey, callerID,
	)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to create API key")
		return
	}

	id, _ := result.LastInsertId()

	h.writeJSON(w, http.StatusCreated, map[string]any{
		"id":  id,
		"key": apiKey,
	})
}

func (h *Handler) APIKeysDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var exists int
	h.db.QueryRow("SELECT COUNT(*) FROM api_keys WHERE id = ?", id).Scan(&exists)
	if exists == 0 {
		h.writeError(w, http.StatusNotFound, "API key not found")
		return
	}

	h.db.Exec("DELETE FROM api_keys WHERE id = ?", id)
	h.writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) PricingStatus(w http.ResponseWriter, r *http.Request) {
	var count int
	var updatedAt *time.Time
	h.db.QueryRow("SELECT COUNT(*), MAX(updated_at) FROM model_pricing").Scan(&count, &updatedAt)
	h.writeJSON(w, http.StatusOK, map[string]any{
		"model_count": count,
		"updated_at":  updatedAt,
	})
}

func (h *Handler) PricingSync(w http.ResponseWriter, r *http.Request) {
	if err := pricing.Sync(h.db); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Sync failed: "+err.Error())
		return
	}
	var count int
	h.db.QueryRow("SELECT COUNT(*) FROM model_pricing").Scan(&count)
	h.writeJSON(w, http.StatusOK, map[string]any{"ok": true, "model_count": count})
}

func (h *Handler) RetentionGet(w http.ResponseWriter, r *http.Request) {
	var days string
	h.db.QueryRow("SELECT value FROM settings WHERE key = 'retention_days'").Scan(&days)
	h.writeJSON(w, http.StatusOK, map[string]string{"retention_days": days})
}

func (h *Handler) RetentionSet(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Days string `json:"retention_days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	h.db.Exec("INSERT INTO settings (key, value) VALUES ('retention_days', ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value", req.Days)
	h.writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) WhoAmI(w http.ResponseWriter, r *http.Request) {
	callerID := r.Context().Value(userIDKey).(int64)
	var email string
	err := h.db.QueryRow("SELECT email FROM users WHERE id = ?", callerID).Scan(&email)
	if err == sql.ErrNoRows {
		h.writeError(w, http.StatusNotFound, "User not found")
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]any{"id": callerID, "email": email})
}
