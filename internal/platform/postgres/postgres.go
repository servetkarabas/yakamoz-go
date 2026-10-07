package postgres

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Pool struct {
	*sql.DB
}

func (p *Pool) Ping(ctx context.Context) error {
	return p.PingContext(ctx)
}

func Open(_ context.Context, dsn string) (*Pool, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return &Pool{DB: db}, nil
}

func Migrate(ctx context.Context, pool *Pool, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)
	for _, name := range files {
		content, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		if _, err := pool.ExecContext(ctx, string(content)); err != nil {
			return err
		}
	}
	return nil
}
