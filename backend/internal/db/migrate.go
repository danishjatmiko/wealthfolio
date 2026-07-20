package db

import (
	"database/sql"
	"fmt"
	"io/fs"

	// Registers the "pgx" database/sql driver used only for running goose
	// migrations. Regular request handling uses the native pgxpool below.
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// RunMigrations applies all pending goose "up" migrations found at the root
// of migrationsFS. goose requires a database/sql connection rather than a
// pgx native pool, so this opens and closes its own short-lived connection
// via the pgx stdlib adapter.
func RunMigrations(dsn string, migrationsFS fs.FS) error {
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open migration db connection: %w", err)
	}
	defer sqlDB.Close()

	goose.SetBaseFS(migrationsFS)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	// migrationsFS is rooted directly at the migrations directory itself
	// (see migrations/embed.go), so migration files live at "." within it.
	if err := goose.Up(sqlDB, "."); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}
