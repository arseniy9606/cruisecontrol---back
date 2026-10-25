package cars

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

var ErrCarDraftExists = errors.New("у водителя уже есть черновик")

type CarRepository struct {
	db    *gorm.DB
	sqlDB *sql.DB
}

func OpenCarRepository(dsn string) (*CarRepository, error) {
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
	return &CarRepository{db, s}, nil
}
func (r *CarRepository) Close() error  { return r.sqlDB.Close() }
func (r *CarRepository) ORM() *gorm.DB { return r.db }

// WHERE performs filtering before rows leave PostgreSQL.
func (r *CarRepository) PublishedCars(ctx context.Context, min, max float64) ([]Car, error) {
	cars := make([]Car, 0)
	err := r.db.WithContext(ctx).Where("car_status = ? AND consumption_rate BETWEEN ? AND ?", CarPublished, min, max).Order("car_id").Find(&cars).Error
	return cars, err
}

// First generates LIMIT 1. Each feed SELECT returns at most one row.
func (r *CarRepository) FeedCar(ctx context.Context, id int64, next bool) (Car, error) {
	var car Car
	q := r.db.WithContext(ctx).Where("car_status = ?", CarPublished)
	if id > 0 && !next {
		q = q.Where("car_id = ?", id)
	}
	if id > 0 && next {
		var anchor Car
		if err := r.db.WithContext(ctx).Where("car_id = ? AND car_status = ?", id, CarPublished).First(&anchor).Error; err != nil {
			return car, err
		}
		q = q.Where("car_id > ?", id)
	}
	err := q.Order("car_id").First(&car).Error
	if next && errors.Is(err, gorm.ErrRecordNotFound) {
		err = r.db.WithContext(ctx).Where("car_status = ?", CarPublished).Order("car_id").First(&car).Error
	}
	return car, err
}
func (r *CarRepository) DraftCar(ctx context.Context) (Car, error) {
	var car Car
	err := r.db.WithContext(ctx).Where("car_status = ? AND creator_id = ?", CarDraft, CarCurrentDriverID).First(&car).Error
	return car, err
}
func (r *CarRepository) CarLikeCount(ctx context.Context, id int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&CarLike{}).Where("car_id = ?", id).Count(&count).Error
	return count, err
}
func (r *CarRepository) CreateCar(ctx context.Context, name, mediaBase string) error {
	car := Car{CarName: name, CarStatus: CarDraft, CarCreatorID: CarCurrentDriverID, CarCreatedAt: time.Now(), CarImageURL: mediaBase + "/cars_missing.jpg", CarVideoURL: mediaBase + "/cars_missing.mp4"}
	err := r.db.WithContext(ctx).Create(&car).Error
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "cars_one_draft_per_driver" {
		return ErrCarDraftExists
	}
	return err
}
func (r *CarRepository) PublishCar(ctx context.Context, description string, tank, rate float64) error {
	result := r.db.WithContext(ctx).Model(&Car{}).Where("car_status = ? AND creator_id = ?", CarDraft, CarCurrentDriverID).Updates(map[string]interface{}{"short_description": description, "fuel_tank_volume": tank, "consumption_rate": rate, "car_status": CarPublished, "formed_at": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteCar uses an explicit PostgreSQL cursor in a single transaction.
// The row is locked, then soft-deleted at the current cursor position.
func (r *CarRepository) DeleteCar(ctx context.Context, id int64) error {
	tx, err := r.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `DECLARE car_to_delete NO SCROLL CURSOR FOR
 SELECT car_id FROM cars WHERE car_id = $1 AND creator_id = $2
 AND car_status <> $3 FOR UPDATE`, id, CarCurrentDriverID, CarDeleted)
	if err != nil {
		return err
	}
	var selectedID int64
	err = tx.QueryRowContext(ctx, "FETCH NEXT FROM car_to_delete").Scan(&selectedID)
	if errors.Is(err, sql.ErrNoRows) {
		return gorm.ErrRecordNotFound
	}
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE cars SET car_status = $1 WHERE CURRENT OF car_to_delete`, CarDeleted)
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, "CLOSE car_to_delete"); err != nil {
		return err
	}
	return tx.Commit()
}
func (r *CarRepository) ValidateDriver(ctx context.Context) error {
	var driver Driver
	if err := r.db.WithContext(ctx).First(&driver, CarCurrentDriverID).Error; err != nil {
		return fmt.Errorf("водитель 1 отсутствует: выполните migrations/sql/002_seed.sql через Adminer: %w", err)
	}
	return nil
}
