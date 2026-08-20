package platform

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func Migrate(ctx context.Context, db *pgxpool.Pool) error {
	if _, e := db.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())`); e != nil {
		return e
	}
	base := os.Getenv("MIGRATIONS_PATH")
	if base == "" {
		base = "migrations"
	}
	if st, err := os.Stat(base); err != nil || !st.IsDir() {
		if executable, err := os.Executable(); err == nil {
			base = filepath.Join(filepath.Dir(executable), "migrations")
		}
	}
	files, err := filepath.Glob(filepath.Join(base, "[0-9][0-9][0-9]_*.sql"))
	if err != nil {
		return err
	}
	sort.Strings(files)
	for _, path := range files {
		name := filepath.Base(path)
		version := strings.TrimSuffix(name, ".sql")
		var applied bool
		if e := db.QueryRow(ctx, `select exists(select 1 from schema_migrations where version=$1)`, version).Scan(&applied); e != nil {
			return e
		}
		if applied {
			continue
		}
		b, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		for _, s := range strings.Split(string(b), ";") {
			if strings.TrimSpace(s) != "" {
				if _, e = db.Exec(ctx, s); e != nil {
					return e
				}
			}
		}
		if _, e = db.Exec(ctx, `insert into schema_migrations(version) values($1)`, version); e != nil {
			return e
		}
	}
	return nil
}
