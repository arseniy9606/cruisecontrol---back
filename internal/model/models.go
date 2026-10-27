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
	Status          string    `gorm:"not null;index"`
	ConsumptionRate *float64  `gorm:"type:numeric(7,2)"`
	FuelTankVolume  *float64  `gorm:"type:numeric(7,2)"`
	ImageName       string    `gorm:"not null"`
	VideoName       string    `gorm:"not null"`
	CreatorID       uint      `gorm:"not null;uniqueIndex:one_car_draft,where:status = 'draft'"`
	Creator         User      `gorm:"foreignKey:CreatorID"`
	CreatedAt       time.Time `gorm:"not null"`
	FormedAt        *time.Time
}
type CarLike struct {
	ID     uint `gorm:"primaryKey"`
	CarID  uint `gorm:"not null;uniqueIndex:one_car_like"`
	UserID uint `gorm:"not null;uniqueIndex:one_car_like"`
	Car    Car  `gorm:"foreignKey:CarID"`
	User   User `gorm:"foreignKey:UserID"`
}
