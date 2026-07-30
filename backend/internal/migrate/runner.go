package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	migrationfiles "soulink-backend/db/migrations"
)

// Up 顺序执行所有尚未应用的数据库迁移。
func Up(ctx context.Context, db *sql.DB) error {
	if err := ensureSchemaMigrationsTable(ctx, db); err != nil {
		return err
	}

	entries, err := fs.ReadDir(migrationfiles.Files, ".")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		applied, err := hasMigration(ctx, db, entry.Name())
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		contents, err := migrationfiles.Files.ReadFile(entry.Name())
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration transaction: %w", err)
		}

		if _, err := tx.ExecContext(ctx, normalizeSQLMigration(contents)); err != nil {
			tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", entry.Name(), err)
		}

		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO schema_migrations (version, applied_at) VALUES ($1, $2)`,
			entry.Name(),
			time.Now(),
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", entry.Name(), err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// normalizeSQLMigration 规范化 SQL 文本，避免 UTF-8 BOM 等编码标记影响数据库执行。
func normalizeSQLMigration(contents []byte) string {
	return strings.TrimPrefix(string(contents), "\uFEFF")
}

// ensureSchemaMigrationsTable 确保迁移记录表存在。
func ensureSchemaMigrationsTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}
	return nil
}

// hasMigration 判断某个迁移版本是否已经执行过。
func hasMigration(ctx context.Context, db *sql.DB, version string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`,
		version,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check migration %s: %w", version, err)
	}
	return exists, nil
}
