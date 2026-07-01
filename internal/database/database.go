// Package database owns the PostgreSQL connection pool and a tiny, dependency-
// free migration runner that applies the embedded *.up.sql files in order.
package database

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kusl/GoTunnels/internal/config"
	"github.com/kusl/GoTunnels/migrations"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect opens a pgx connection pool using the resolved configuration and
// verifies connectivity, retrying briefly so the API can win the startup race
// against a not-yet-ready Postgres even when compose healthchecks are bypassed.
func Connect(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("database: parse DATABASE_URL: %w", err)
	}
	if cfg.DBMaxConns > 0 {
		poolCfg.MaxConns = cfg.DBMaxConns
	}
	if cfg.DBMinConns > 0 {
		poolCfg.MinConns = cfg.DBMinConns
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("database: create pool: %w", err)
	}

	deadline := time.Now().Add(cfg.DBConnectTimeout)
	var lastErr error
	for time.Now().Before(deadline) {
		pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		lastErr = pool.Ping(pingCtx)
		cancel()
		if lastErr == nil {
			return pool, nil
		}
		select {
		case <-ctx.Done():
			pool.Close()
			return nil, ctx.Err()
		case <-time.After(1 * time.Second):
		}
	}
	pool.Close()
	return nil, fmt.Errorf("database: could not reach Postgres within %s: %w", cfg.DBConnectTimeout, lastErr)
}

// migration is a parsed up-migration file.
type migration struct {
	version int64
	name    string
	sql     string
}

// Migrate applies all pending up-migrations found in the embedded FS. It is
// idempotent: already-applied versions are skipped. Each migration runs inside
// its own transaction so a failure leaves the schema at the last good version.
func Migrate(ctx context.Context, pool *pgxpool.Pool) (applied []int64, err error) {
	if _, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    bigint PRIMARY KEY,
			name       text NOT NULL,
			applied_at timestamptz NOT NULL DEFAULT now()
		)`); err != nil {
		return nil, fmt.Errorf("database: ensure schema_migrations: %w", err)
	}

	migs, err := loadUpMigrations()
	if err != nil {
		return nil, err
	}

	done := map[int64]bool{}
	rows, err := pool.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("database: read applied migrations: %w", err)
	}
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			rows.Close()
			return nil, err
		}
		done[v] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, m := range migs {
		if done[m.version] {
			continue
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			return applied, fmt.Errorf("database: begin tx for migration %d: %w", m.version, err)
		}
		if _, err := tx.Exec(ctx, m.sql); err != nil {
			_ = tx.Rollback(ctx)
			return applied, fmt.Errorf("database: apply migration %d (%s): %w", m.version, m.name, err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`, m.version, m.name); err != nil {
			_ = tx.Rollback(ctx)
			return applied, fmt.Errorf("database: record migration %d: %w", m.version, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return applied, fmt.Errorf("database: commit migration %d: %w", m.version, err)
		}
		applied = append(applied, m.version)
	}
	return applied, nil
}

func loadUpMigrations() ([]migration, error) {
	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		return nil, fmt.Errorf("database: read embedded migrations: %w", err)
	}
	var out []migration
	seen := map[int64]string{}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		version, err := parseVersion(name)
		if err != nil {
			return nil, err
		}
		if prev, dup := seen[version]; dup {
			return nil, fmt.Errorf("database: duplicate migration version %d (%s and %s)", version, prev, name)
		}
		seen[version] = name
		data, err := migrations.FS.ReadFile(name)
		if err != nil {
			return nil, fmt.Errorf("database: read %s: %w", name, err)
		}
		out = append(out, migration{version: version, name: name, sql: string(data)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].version < out[j].version })
	return out, nil
}

// parseVersion extracts the leading integer of "0001_something.up.sql".
func parseVersion(filename string) (int64, error) {
	base := filename
	idx := strings.IndexAny(base, "_.")
	if idx <= 0 {
		return 0, fmt.Errorf("database: migration filename %q must start with a numeric version", filename)
	}
	v, err := strconv.ParseInt(base[:idx], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("database: migration filename %q has invalid version prefix: %w", filename, err)
	}
	return v, nil
}
