package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
)

//go:embed sql/*.sql
var carSQL embed.FS

func CarSchema() string {
	b, _ := carSQL.ReadFile("sql/001_schema.sql")
	return string(b)
}
func CarSeed() string {
	b, _ := carSQL.ReadFile("sql/002_seed.sql")
	return string(b)
}
func ApplyCarSchema(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, CarSchema()); err != nil {
		return fmt.Errorf("schema 001: %w", err)
	}
	return tx.Commit()
}

// Seed only a completely empty database. Restarting never resets user changes.
func SeedEmptyCarDatabase(ctx context.Context, db *sql.DB) error {
	var count int
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM cars").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err := db.ExecContext(ctx, CarSeed())
	return err
}
