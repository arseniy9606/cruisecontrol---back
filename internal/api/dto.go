package api

import (
	"github.com/arseniy9606/cruisecontrol---back/internal/model"
	"github.com/arseniy9606/cruisecontrol---back/internal/service"
	"github.com/gin-gonic/gin"
	"time"
)

// Every response has the same keys, including nullable draft fields. Status is database-only.
type CarDTO struct {
	ID              uint       `json:"id"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	ConsumptionRate *float64   `json:"consumption_rate"`
	FuelTankVolume  *float64   `json:"fuel_tank_volume"`
	ImageURL        string     `json:"image_url"`
	VideoURL        string     `json:"video_url"`
	Creator         string     `json:"creator"`
	CreatedAt       time.Time  `json:"created_at"`
	FormedAt        *time.Time `json:"formed_at"`
	IsOwner         int        `json:"is_owner"`
	Liked           int        `json:"liked"`
	LikesCount      int64      `json:"likes_count"`
}

func (s *Server) carDTO(car model.Car) (CarDTO, error) {
	d := CarDTO{ID: car.ID, Name: car.Name, Description: car.Description, ConsumptionRate: car.ConsumptionRate, FuelTankVolume: car.FuelTankVolume, ImageURL: s.Media.URL(car.ImageName), VideoURL: s.Media.URL(car.VideoName), Creator: car.Creator.Login, CreatedAt: car.CreatedAt, FormedAt: car.FormedAt}
	if car.CreatorID == service.CurrentUser().ID {
		d.IsOwner = 1
	}
	if e := s.DB.Model(&model.CarLike{}).Where("car_id = ?", car.ID).Count(&d.LikesCount).Error; e != nil {
		return d, e
	}
	var count int64
	e := s.DB.Model(&model.CarLike{}).Where("car_id = ? AND user_id = ?", car.ID, service.CurrentUser().ID).Count(&count).Error
	if count > 0 {
		d.Liked = 1
	}
	return d, e
}
func (s *Server) respondCar(c *gin.Context, car model.Car, code int) {
	if e := s.DB.Preload("Creator").First(&car, car.ID).Error; e != nil {
		fail(c, e)
		return
	}
	d, e := s.carDTO(car)
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(code, d)
}
