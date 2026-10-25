package cars

import "time"

type CarStatus string

const (
	CarDraft     CarStatus = "черновик"
	CarPublished CarStatus = "опубликован"
	CarDeleted   CarStatus = "удален"
)

// The three models map to subject-specific PostgreSQL tables; no in-memory database.
type Driver struct {
	DriverID     int64  `gorm:"column:driver_id;primaryKey"`
	DriverLogin  string `gorm:"column:login;type:varchar(60);not null;unique"`
	DriverName   string `gorm:"column:full_name;type:varchar(120);not null"`
	PasswordHash string `gorm:"column:password_hash;type:varchar(100);not null;default:!disabled"`
}

func (Driver) TableName() string { return "drivers" }

type Car struct {
	CarID          int64      `gorm:"column:car_id;primaryKey"`
	CarName        string     `gorm:"column:car_name;type:varchar(120);not null"`
	CarDescription string     `gorm:"column:short_description;type:varchar(1000);default:null"`
	CarStatus      CarStatus  `gorm:"column:car_status;type:varchar(20);not null"`
	CarImageURL    string     `gorm:"column:image_url;type:varchar(500);not null"`
	CarVideoURL    string     `gorm:"column:video_url;type:varchar(500);not null"`
	FuelTankVolume float64    `gorm:"column:fuel_tank_volume;type:numeric(7,1);default:null"`
	CarRate        float64    `gorm:"column:consumption_rate;type:numeric(4,1);default:null"`
	CarCreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime"`
	CarCreatorID   int64      `gorm:"column:creator_id;not null"`
	CarFormedAt    *time.Time `gorm:"column:formed_at"`
}

func (Car) TableName() string { return "cars" }

type CarLike struct {
	CarLikeID int64 `gorm:"column:like_id;primaryKey"`
	CarID     int64 `gorm:"column:car_id;not null"`
	DriverID  int64 `gorm:"column:driver_id;not null"`
}

func (CarLike) TableName() string { return "car_likes" }

// Lab 2 has no authentication: all controllers use the demonstration driver.
const CarCurrentDriverID int64 = 1
