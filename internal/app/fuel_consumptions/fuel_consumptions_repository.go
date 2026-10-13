package fuel_consumptions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

var ErrFuelConsumptionDraftExists = errors.New("у водителя уже есть черновик")

type FuelConsumptionRepository struct {
	db    *gorm.DB
	sqlDB *sql.DB
}

func OpenFuelConsumptionRepository(dsn string) (*FuelConsumptionRepository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		return nil, err
	}
	s, err := db.DB()
	if err != nil {
		return nil, err
	}
	s.SetMaxOpenConns(10)
	s.SetMaxIdleConns(5)
	s.SetConnMaxLifetime(30 * time.Minute)
	return &FuelConsumptionRepository{db, s}, nil
}
func (r *FuelConsumptionRepository) Close() error  { return r.sqlDB.Close() }
func (r *FuelConsumptionRepository) ORM() *gorm.DB { return r.db }

// WHERE performs filtering before rows leave PostgreSQL.
func (r *FuelConsumptionRepository) PublishedFuelConsumptions(ctx context.Context, min, max float64) ([]FuelConsumption, error) {
	cars := make([]FuelConsumption, 0)
	err := r.db.WithContext(ctx).Where("fuel_consumption_status = ? AND fuel_consumption_rate BETWEEN ? AND ?", FuelConsumptionPublished, min, max).Order("fuel_consumption_id").Find(&cars).Error
	return cars, err
}

// First generates LIMIT 1. Each feed SELECT returns at most one row.
func (r *FuelConsumptionRepository) FeedFuelConsumption(ctx context.Context, id int64, next bool) (FuelConsumption, error) {
	var car FuelConsumption
	q := r.db.WithContext(ctx).Where("fuel_consumption_status = ?", FuelConsumptionPublished)
	if id > 0 && !next {
		q = q.Where("fuel_consumption_id = ?", id)
	}
	if id > 0 && next {
		var anchor FuelConsumption
		if err := r.db.WithContext(ctx).Where("fuel_consumption_id = ? AND fuel_consumption_status = ?", id, FuelConsumptionPublished).First(&anchor).Error; err != nil {
			return car, err
		}
		q = q.Where("fuel_consumption_id > ?", id)
	}
	err := q.Order("fuel_consumption_id").First(&car).Error
	if next && errors.Is(err, gorm.ErrRecordNotFound) {
		err = r.db.WithContext(ctx).Where("fuel_consumption_status = ?", FuelConsumptionPublished).Order("fuel_consumption_id").First(&car).Error
	}
	return car, err
}
func (r *FuelConsumptionRepository) DraftFuelConsumption(ctx context.Context) (FuelConsumption, error) {
	var car FuelConsumption
	err := r.db.WithContext(ctx).Where("fuel_consumption_status = ? AND fuel_consumption_creator_id = ?", FuelConsumptionDraft, FuelConsumptionCurrentDriverID).First(&car).Error
	return car, err
}
func (r *FuelConsumptionRepository) FuelConsumptionLikeCount(ctx context.Context, id int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&FuelConsumptionLike{}).Where("fuel_consumption_id = ?", id).Count(&count).Error
	return count, err
}
func (r *FuelConsumptionRepository) CreateFuelConsumption(ctx context.Context, name, mediaBase string) error {
	car := FuelConsumption{FuelConsumptionName: name, FuelConsumptionStatus: FuelConsumptionDraft, FuelConsumptionCreatorID: FuelConsumptionCurrentDriverID, FuelConsumptionCreatedAt: time.Now(), FuelConsumptionImageURL: mediaBase + "/fuel_consumptions_missing.jpg", FuelConsumptionVideoURL: mediaBase + "/fuel_consumptions_missing.mp4"}
	err := r.db.WithContext(ctx).Create(&car).Error
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "fuel_consumptions_one_draft_per_driver" {
		return ErrFuelConsumptionDraftExists
	}
	return err
}
func (r *FuelConsumptionRepository) PublishFuelConsumption(ctx context.Context, description string, tank, rate float64) error {
	result := r.db.WithContext(ctx).Model(&FuelConsumption{}).Where("fuel_consumption_status = ? AND fuel_consumption_creator_id = ?", FuelConsumptionDraft, FuelConsumptionCurrentDriverID).Updates(map[string]interface{}{"fuel_consumption_description": description, "fuel_tank_volume": tank, "fuel_consumption_rate": rate, "fuel_consumption_status": FuelConsumptionPublished, "fuel_consumption_formed_at": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Deliberately bypass ORM: database/sql, parameterized UPDATE and RowsAffected.
func (r *FuelConsumptionRepository) DeleteFuelConsumption(ctx context.Context, id int64) error {
	result, err := r.sqlDB.ExecContext(ctx, `UPDATE fuel_consumptions SET fuel_consumption_status = $1 WHERE fuel_consumption_id = $2 AND fuel_consumption_creator_id = $3 AND fuel_consumption_status <> $1`, FuelConsumptionDeleted, id, FuelConsumptionCurrentDriverID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
func (r *FuelConsumptionRepository) ValidateDriver(ctx context.Context) error {
	var driver FuelConsumptionDriver
	if err := r.db.WithContext(ctx).First(&driver, FuelConsumptionCurrentDriverID).Error; err != nil {
		return fmt.Errorf("водитель 1 отсутствует: выполните migrations/sql/002_seed.sql через Adminer: %w", err)
	}
	return nil
}
