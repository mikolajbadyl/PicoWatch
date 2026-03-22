package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"picowatch/db"
)

type Handler struct {
	db        *sql.DB
	jwtSecret []byte
}

func New(database *sql.DB) *Handler {
	secret, err := db.GetSetting(database, "jwt_secret")
	if err != nil {
		panic("Failed to get JWT secret: " + err.Error())
	}
	return &Handler{
		db:        database,
		jwtSecret: []byte(secret),
	}
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, msg string) {
	h.writeJSON(w, status, map[string]string{"error": msg})
}
