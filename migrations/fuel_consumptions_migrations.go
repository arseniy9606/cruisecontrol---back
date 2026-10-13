package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
)

//go:embed sql/*.sql
var fuelConsumptionSQL embed.FS

func FuelConsumptionSchema() string {
	b, _ := fuelConsumptionSQL.ReadFile("sql/001_schema.sql")
	return string(b)
}
func FuelConsumptionSeed() string {
	b, _ := fuelConsumptionSQL.ReadFile("sql/002_seed.sql")
	return string(b)
}
func ApplyFuelConsumptionSchema(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, FuelConsumptionSchema()); err != nil {
		return fmt.Errorf("schema 001: %w", err)
	}
	return tx.Commit()
}

// Seed only a completely empty database. Restarting never resets user changes.
func SeedEmptyFuelConsumptionDatabase(ctx context.Context, db *sql.DB) error {
	var count int
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM fuel_consumptions").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err := db.ExecContext(ctx, FuelConsumptionSeed())
	return err
}
