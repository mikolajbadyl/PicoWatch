package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"picowatch/db"
	"picowatch/handlers"
	"picowatch/pricing"
)

//go:embed all:ui/build
var staticFiles embed.FS

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	database, err := db.Init()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer database.Close()

	pricing.StartSync(database)

	go func() {
		for {
			var daysStr string
			if err := database.QueryRow("SELECT value FROM settings WHERE key = 'retention_days'").Scan(&daysStr); err == nil {
				if days, err := strconv.Atoi(daysStr); err == nil && days > 0 {
					res, _ := database.Exec("DELETE FROM llm_logs WHERE created_at < datetime('now', ? || ' days')", strconv.Itoa(-days))
					if n, _ := res.RowsAffected(); n > 0 {
						log.Printf("Retention: deleted %d old log entries", n)
					}
				}
			}
			time.Sleep(24 * time.Hour)
		}
	}()

	h := handlers.New(database)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/auth/status", h.AuthStatus)
	mux.HandleFunc("POST /api/auth/setup", h.AuthSetup)
	mux.HandleFunc("POST /api/auth/login", h.AuthLogin)

	mux.Handle("GET /api/logs", h.RequireAuth(http.HandlerFunc(h.LogsList)))
	mux.Handle("GET /api/logs/export", h.RequireAuth(http.HandlerFunc(h.LogsExport)))
	mux.Handle("DELETE /api/logs/{id}", h.RequireAuth(http.HandlerFunc(h.LogsDelete)))
	mux.Handle("GET /api/stats", h.RequireAuth(http.HandlerFunc(h.StatsGet)))
	mux.Handle("GET /api/stats/models", h.RequireAuth(http.HandlerFunc(h.StatsModels)))
	mux.Handle("GET /api/stats/daily", h.RequireAuth(http.HandlerFunc(h.StatsDaily)))

	mux.Handle("POST /api/logs", h.RequireAnyAuth(http.HandlerFunc(h.LogsCreate)))

	mux.Handle("GET /api/users", h.RequireAuth(http.HandlerFunc(h.UsersList)))
	mux.Handle("POST /api/users", h.RequireAuth(http.HandlerFunc(h.UsersCreate)))
	mux.Handle("DELETE /api/users/{id}", h.RequireAuth(http.HandlerFunc(h.UsersDelete)))
	mux.Handle("PATCH /api/users/{id}/password", h.RequireAuth(http.HandlerFunc(h.UsersChangePassword)))

	mux.Handle("GET /api/apikeys", h.RequireAuth(http.HandlerFunc(h.APIKeysList)))
	mux.Handle("POST /api/apikeys", h.RequireAuth(http.HandlerFunc(h.APIKeysCreate)))
	mux.Handle("DELETE /api/apikeys/{id}", h.RequireAuth(http.HandlerFunc(h.APIKeysDelete)))

	mux.Handle("GET /api/me", h.RequireAuth(http.HandlerFunc(h.WhoAmI)))

	mux.Handle("GET /api/pricing/status", h.RequireAuth(http.HandlerFunc(h.PricingStatus)))
	mux.Handle("POST /api/pricing/sync", h.RequireAuth(http.HandlerFunc(h.PricingSync)))

	mux.Handle("GET /api/retention", h.RequireAuth(http.HandlerFunc(h.RetentionGet)))
	mux.Handle("POST /api/retention", h.RequireAuth(http.HandlerFunc(h.RetentionSet)))

	subFS, err := fs.Sub(staticFiles, "ui/build")
	if err != nil {
		log.Fatal("Failed to create sub filesystem:", err)
	}

	fileServer := http.FileServer(http.FS(subFS))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path != "" {
			if _, err := fs.Stat(subFS, path); err != nil {
				r2 := r.Clone(r.Context())
				r2.URL.Path = "/"
				fileServer.ServeHTTP(w, r2)
				return
			}
		}
		fileServer.ServeHTTP(w, r)
	})

	log.Printf("PicoWatch running on http://0.0.0.0:%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
