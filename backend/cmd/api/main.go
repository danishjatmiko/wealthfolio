// Command api is the Wealthfolio backend HTTP server. On startup it applies
// any pending goose migrations, opens a pgx connection pool, and serves the
// REST API described in the project spec.
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"wealthfolio/backend/internal/config"
	"wealthfolio/backend/internal/db"
	"wealthfolio/backend/internal/httpapi"
	"wealthfolio/backend/internal/service"
	"wealthfolio/backend/migrations"
)

func main() {
	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}
	if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" {
		log.Fatal("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET environment variables are required")
	}

	log.Println("running database migrations...")
	if err := db.RunMigrations(cfg.DatabaseURL, migrations.FS); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}
	log.Println("migrations up to date")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	repos := db.NewRepos(pool)
	services := service.NewServices(repos, cfg)
	router := httpapi.NewRouter(cfg, repos, services)

	addr := ":" + cfg.Port
	log.Printf("wealthfolio api listening on %s (cors origin: %s)", addr, cfg.CORSOrigin)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
