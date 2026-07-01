// Package migrations embeds the SQL migration files so they travel inside the
// compiled binary. The application applies pending "up" migrations on startup.
package migrations

import "embed"

// FS holds every .sql migration file in this directory.
//
//go:embed *.sql
var FS embed.FS
