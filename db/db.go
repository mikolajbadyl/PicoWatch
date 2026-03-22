package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func Init() (*sql.DB, error) {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		dbPath = filepath.Join(home, ".picowatch", "picowatch.db")
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	db, err := sql.Open("sqlite", "file:"+dbPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(on)")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	db.SetMaxOpenConns(10)

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS api_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		key TEXT NOT NULL UNIQUE,
		user_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_used_at DATETIME,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS llm_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		model TEXT NOT NULL,
		provider TEXT DEFAULT '',
		input_tokens INTEGER DEFAULT 0,
		output_tokens INTEGER DEFAULT 0,
		cost REAL DEFAULT 0,
		duration_ms INTEGER DEFAULT 0,
		status TEXT DEFAULT 'success',
		error TEXT DEFAULT '',
		metadata TEXT DEFAULT '{}',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);


	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS model_pricing (
		model_id TEXT PRIMARY KEY,
		prompt_price REAL NOT NULL,
		completion_price REAL NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_llm_logs_created_at ON llm_logs(created_at);
	CREATE INDEX IF NOT EXISTS idx_llm_logs_model ON llm_logs(model);
	`

	if _, err := db.Exec(schema); err != nil {
		return err
	}

	db.Exec("ALTER TABLE llm_logs ADD COLUMN provider TEXT DEFAULT ''")

	var count int
	db.QueryRow("SELECT COUNT(*) FROM settings WHERE key = 'jwt_secret'").Scan(&count)
	if count == 0 {
		secret := generateRandom(32)
		db.Exec("INSERT INTO settings (key, value) VALUES ('jwt_secret', ?)", secret)
		log.Println("Generated new JWT secret")
	}

	return nil
}

func generateRandom(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func GetSetting(db *sql.DB, key string) (string, error) {
	var value string
	err := db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	return value, err
}

func GenerateAPIKey() string {
	b := make([]byte, 24)
	rand.Read(b)
	return "pk_" + hex.EncodeToString(b)
}
