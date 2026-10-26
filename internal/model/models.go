package model

import "time"

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Login        string `gorm:"uniqueIndex;not null;size:60"`
	FullName     string `gorm:"not null;size:120"`
	PasswordHash string `gorm:"not null"`
}
type Car struct {
	ID              uint   `gorm:"primaryKey"`
	Name            string `gorm:"not null;size:120"`
	Description     string
	Status          string  `gorm:"not null;default:published"`
	ConsumptionRate float64 `gorm:"type:numeric(7,2);not null;check:consumption_rate > 0"`
	FuelTankVolume  float64 `gorm:"type:numeric(7,2);not null;check:fuel_tank_volume > 0"`
	ImageName       string
	VideoName       string
}
type FuelConsumption struct {
	ID               uint   `gorm:"primaryKey"`
	Status           string `gorm:"not null;index"`
	CreatorID        uint   `gorm:"not null;uniqueIndex:one_draft,where:status = 'draft'"`
	Creator          User   `gorm:"foreignKey:CreatorID"`
	ModeratorID      *uint
	Moderator        *User `gorm:"foreignKey:ModeratorID"`
	CreatedAt        time.Time
	FormedAt         *time.Time `gorm:"index"`
	CompletedAt      *time.Time
	Route            string
	SavingPercent    *float64  `gorm:"type:numeric(5,2)"`
	TotalSavedLiters *float64  `gorm:"type:numeric(12,2)"`
	Items            []TripCar `gorm:"foreignKey:RequestID"`
}

func (FuelConsumption) TableName() string { return "fuel_consumptions" }

type TripCar struct {
	RequestID    uint     `gorm:"primaryKey;autoIncrement:false"`
	CarID        uint     `gorm:"primaryKey;autoIncrement:false"`
	Car          Car      `gorm:"foreignKey:CarID"`
	DistanceKM   *float64 `gorm:"type:numeric(10,2)"`
	BaseLiters   *float64 `gorm:"type:numeric(12,2)"`
	CruiseLiters *float64 `gorm:"type:numeric(12,2)"`
	SavedLiters  *float64 `gorm:"type:numeric(12,2)"`
	RangeKM      *float64 `gorm:"type:numeric(12,2)"`
}
