package fuel_consumptions

import "time"

type FuelConsumptionStatus string

const (
	FuelConsumptionDraft     FuelConsumptionStatus = "черновик"
	FuelConsumptionPublished FuelConsumptionStatus = "опубликован"
	FuelConsumptionDeleted   FuelConsumptionStatus = "удален"
)

// The three models map to subject-specific PostgreSQL tables; no in-memory database.
type FuelConsumptionDriver struct {
	FuelConsumptionDriverID        int64     `gorm:"column:fuel_consumption_driver_id;primaryKey"`
	FuelConsumptionDriverLogin     string    `gorm:"column:fuel_consumption_driver_login;type:varchar(60);not null;unique"`
	FuelConsumptionDriverName      string    `gorm:"column:fuel_consumption_driver_name;type:varchar(120);not null"`
	FuelConsumptionDriverCreatedAt time.Time `gorm:"column:fuel_consumption_driver_created_at;autoCreateTime"`
}

func (FuelConsumptionDriver) TableName() string { return "fuel_consumption_drivers" }

type FuelConsumption struct {
	FuelConsumptionID          int64                 `gorm:"column:fuel_consumption_id;primaryKey"`
	FuelConsumptionName        string                `gorm:"column:fuel_consumption_name;type:varchar(120);not null"`
	FuelConsumptionDescription string                `gorm:"column:fuel_consumption_description;type:varchar(1000);not null"`
	FuelConsumptionStatus      FuelConsumptionStatus `gorm:"column:fuel_consumption_status;type:varchar(20);not null"`
	FuelConsumptionImageURL    string                `gorm:"column:fuel_consumption_image_url;type:varchar(500);not null"`
	FuelConsumptionVideoURL    string                `gorm:"column:fuel_consumption_video_url;type:varchar(500);not null"`
	FuelTankVolume             float64               `gorm:"column:fuel_tank_volume;type:numeric(7,1);not null"`
	FuelConsumptionRate        float64               `gorm:"column:fuel_consumption_rate;type:numeric(4,1);not null"`
	FuelConsumptionCreatedAt   time.Time             `gorm:"column:fuel_consumption_created_at;autoCreateTime"`
	FuelConsumptionCreatorID   int64                 `gorm:"column:fuel_consumption_creator_id;not null"`
	FuelConsumptionFormedAt    *time.Time            `gorm:"column:fuel_consumption_formed_at"`
}

func (FuelConsumption) TableName() string { return "fuel_consumptions" }

type FuelConsumptionLike struct {
	FuelConsumptionLikeID   int64 `gorm:"column:fuel_consumption_like_id;primaryKey"`
	FuelConsumptionID       int64 `gorm:"column:fuel_consumption_id;not null"`
	FuelConsumptionDriverID int64 `gorm:"column:fuel_consumption_driver_id;not null"`
}

func (FuelConsumptionLike) TableName() string { return "fuel_consumption_likes" }

// Lab 2 has no authentication: all controllers use the demonstration driver.
const FuelConsumptionCurrentDriverID int64 = 1
