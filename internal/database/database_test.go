package database

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/kusl/GoTunnels/migrations"
)

func TestParseVersion(t *testing.T) {
	cases := map[string]int64{
		"0001_init.up.sql":             1,
		"0002_auth_credentials.up.sql": 2,
		"10_later.up.sql":              10,
		"0005_csp_reports.down.sql":    5,
	}
	for name, want := range cases {
		got, err := parseVersion(name)
		if err != nil {
			t.Errorf("parseVersion(%q) error: %v", name, err)
			continue
		}
		if got != want {
			t.Errorf("parseVersion(%q) = %d, want %d", name, got, want)
		}
	}
}

func TestParseVersion_Invalid(t *testing.T) {
	for _, bad := range []string{"init.up.sql", "_1.up.sql", "abc_1.up.sql"} {
		if _, err := parseVersion(bad); err == nil {
			t.Errorf("parseVersion(%q) expected error", bad)
		}
	}
}

// TestEmbeddedMigrationsParse ensures every embedded up-migration has a valid,
// unique version and a matching down-migration. This runs without a database.
func TestEmbeddedMigrationsParse(t *testing.T) {
	migs, err := loadUpMigrations()
	if err != nil {
		t.Fatalf("loadUpMigrations: %v", err)
	}
	if len(migs) == 0 {
		t.Fatal("expected at least one embedded migration")
	}
	// Versions must be strictly increasing after sort.
	for i := 1; i < len(migs); i++ {
		if migs[i].version <= migs[i-1].version {
			t.Errorf("migrations not strictly increasing at index %d: %d then %d",
				i, migs[i-1].version, migs[i].version)
		}
	}
	// Each up must have a corresponding down file and non-empty SQL.
	entries, _ := fs.ReadDir(migrations.FS, ".")
	names := map[string]bool{}
	for _, e := range entries {
		names[e.Name()] = true
	}
	for _, m := range migs {
		if strings.TrimSpace(m.sql) == "" {
			t.Errorf("migration %d (%s) has empty SQL", m.version, m.name)
		}
		down := strings.TrimSuffix(m.name, ".up.sql") + ".down.sql"
		if !names[down] {
			t.Errorf("migration %s missing down counterpart %s", m.name, down)
		}
	}
}
