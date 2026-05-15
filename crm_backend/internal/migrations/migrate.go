// Package migrations applies all bundled .sql files in order, tracking
// applied versions in the schema_migrations table. Idempotent: safe to call
// on every server startup.
package migrations

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strings"

	"crm_backend/pkg/database"
)

//go:embed sql/*.sql
var migrationFS embed.FS

// Apply scans embedded SQL files (sql/000NNN_*.sql), executes any not yet
// recorded in schema_migrations, and records each on success.
func Apply(ctx context.Context, db *database.DB) error {
	if db == nil || db.Pool == nil {
		return fmt.Errorf("migrations: db is nil")
	}

	// Ensure tracking table exists
	if _, err := db.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("migrations: create schema_migrations: %w", err)
	}

	// Load already-applied versions
	applied := map[string]bool{}
	rows, err := db.Pool.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("migrations: read schema_migrations: %w", err)
	}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			rows.Close()
			return err
		}
		applied[v] = true
	}
	rows.Close()

	// Collect & sort migration files
	entries, err := fs.ReadDir(migrationFS, "sql")
	if err != nil {
		return fmt.Errorf("migrations: read embed: %w", err)
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		if applied[name] {
			continue
		}
		body, err := migrationFS.ReadFile("sql/" + name)
		if err != nil {
			return fmt.Errorf("migrations: read %s: %w", name, err)
		}

		log.Printf("[migrate] applying %s", name)
		tx, err := db.Pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("migrations: begin tx: %w", err)
		}
		if _, err := tx.Exec(ctx, string(body)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("migrations: exec %s: %w", name, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, name); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("migrations: record %s: %w", name, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("migrations: commit %s: %w", name, err)
		}
		log.Printf("[migrate] applied  %s", name)
	}

	log.Println("[migrate] all migrations up to date")
	return nil
}
