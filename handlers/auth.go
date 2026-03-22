package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"picowatch/db"
)

func (h *Handler) AuthStatus(w http.ResponseWriter, r *http.Request) {
	var count int
	h.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	h.writeJSON(w, http.StatusOK, map[string]bool{"hasAdmin": count > 0})
}

type setupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) AuthSetup(w http.ResponseWriter, r *http.Request) {
	var count int
	h.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if count > 0 {
		h.writeError(w, http.StatusForbidden, "Admin account already exists")
		return
	}

	var req setupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		h.writeError(w, http.StatusBadRequest, "Email and password are required")
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
		h.writeError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	userID, _ := result.LastInsertId()

	apiKey := db.GenerateAPIKey()
	h.db.Exec(
		"INSERT INTO api_keys (name, key, user_id) VALUES (?, ?, ?)",
		"Default", apiKey, userID,
	)

	token, err := h.generateToken(userID, req.Email)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	h.writeJSON(w, http.StatusCreated, map[string]string{
		"token":  token,
		"apiKey": apiKey,
	})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) AuthLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var id int64
	var hash string
	err := h.db.QueryRow(
		"SELECT id, password_hash FROM users WHERE email = ?",
		req.Email,
	).Scan(&id, &hash)
	if err == sql.ErrNoRows {
		h.writeError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		h.writeError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	token, err := h.generateToken(id, req.Email)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *Handler) generateToken(userID int64, email string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"exp":   time.Now().Add(30 * 24 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.jwtSecret)
}
