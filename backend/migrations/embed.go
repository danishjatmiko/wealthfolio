// Package migrations embeds the goose SQL migration files so the API binary
// can apply them at startup without needing the directory present on disk
// (important for the Docker image, which only copies the compiled binary).
//
// This file intentionally lives alongside the .sql files themselves because
// Go's //go:embed directive cannot reference a parent directory (no "..").
// It does not modify 00001_init.sql or 00002_seed.sql.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
