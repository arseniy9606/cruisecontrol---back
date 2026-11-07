package api

import (
	"github.com/arseniy9606/cruisecontrol---back/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"math"
	"strconv"
	"strings"
	"time"
)

func validNumber(v float64) bool { return v > 0 && v <= 99999 && !math.IsNaN(v) && !math.IsInf(v, 0) }

// Feed godoc
// @Summary Лента автомобилей
// @Tags Автомобили
// @Produce json
// @Param id query int false "ID автомобиля"
// @Param next query bool false "Следующий автомобиль"
// @Success 200 {object} CarDTO
// @Failure 404
// @Failure 500
// @Router /fuel_consumptions/feed [get]
func (s *Server) Feed(c *gin.Context) {
	q := s.DB.Where("status = ?", "published")
	raw, hasID := c.GetQuery("id")
	next := c.Query("next")
	if next != "" && next != "true" && next != "false" {
		c.Status(404)
		return
	}
	if hasID {
		i, e := strconv.ParseUint(raw, 10, 32)
		if e != nil || i == 0 {
			c.Status(404)
			return
		}
		if next == "true" {
			q = q.Where("id > ?", i)
		} else {
			q = q.Where("id = ?", i)
		}
	} else if next == "true" {
		c.Status(404)
		return
	}
	var car model.Car
	if e := q.Order("id").First(&car).Error; e != nil {
		fail(c, e)
		return
	}
	s.respondCar(c, car, 200)
}

// GetDraft godoc
// @Summary Черновик текущего водителя
// @Tags Автомобили
// @Produce json
// @Success 200 {object} CarDTO
// @Failure 404
// @Failure 500
// @Security BearerAuth
// @Failure 403
// @Failure 503
// @Router /fuel_consumptions/draft [get]
func (s *Server) GetDraft(c *gin.Context) {
	var car model.Car
	if e := s.DB.Where("status = ? AND creator_id = ?", "draft", currentUserID(c)).First(&car).Error; e != nil {
		fail(c, e)
		return
	}
	s.respondCar(c, car, 200)
}

// PublishCar godoc
// @Summary Опубликовать свой черновик
// @Tags Автомобили
// @Produce json
// @Accept json
// @Param body body PublicationRequest true "Поля публикации"
// @Param id path int true "ID автомобиля"
// @Success 200 {object} CarDTO
// @Failure 404
// @Failure 500
// @Security BearerAuth
// @Failure 403
// @Failure 503
// @Router /fuel_consumptions/{id}/publication [put]
func (s *Server) PublishCar(c *gin.Context) {
	i, ok := id(c, "id")
	if !ok {
		return
	}
	var b struct {
		Description     string   `json:"description"`
		ConsumptionRate *float64 `json:"consumption_rate"`
		FuelTankVolume  *float64 `json:"fuel_tank_volume"`
	}
	if !readJSON(c, &b) {
		return
	}
	if strings.TrimSpace(b.Description) == "" || len([]rune(b.Description)) > 1000 || b.ConsumptionRate == nil || b.FuelTankVolume == nil || !validNumber(*b.ConsumptionRate) || !validNumber(*b.FuelTankVolume) {
		c.Status(404)
		return
	}
	var car model.Car
	e := s.DB.Transaction(func(tx *gorm.DB) error {
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND creator_id = ? AND status = ?", i, currentUserID(c), "draft").First(&car).Error; e != nil {
			return e
		}
		return tx.Model(&car).Updates(map[string]any{"description": b.Description, "consumption_rate": b.ConsumptionRate, "fuel_tank_volume": b.FuelTankVolume, "status": "published", "formed_at": time.Now().UTC()}).Error
	})
	if e != nil {
		fail(c, e)
		return
	}
	s.respondCar(c, car, 200)
}

// DeleteCar godoc
// @Summary Удалить свой автомобиль
// @Tags Автомобили
// @Produce json
// @Param id path int true "ID автомобиля"
// @Success 200
// @Failure 404
// @Failure 500
// @Security BearerAuth
// @Failure 403
// @Failure 503
// @Router /fuel_consumptions/{id} [delete]
func (s *Server) DeleteCar(c *gin.Context) {
	i, ok := id(c, "id")
	if !ok {
		return
	}
	if c.Request.ContentLength != 0 {
		var b struct{}
		if !readJSON(c, &b) {
			return
		}
	}
	e := s.DB.Transaction(func(tx *gorm.DB) error {
		var car model.Car
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND creator_id = ? AND status <> ?", i, currentUserID(c), "deleted").First(&car).Error; e != nil {
			return e
		}
		return tx.Model(&car).Update("status", "deleted").Error
	})
	if e != nil {
		fail(c, e)
		return
	}
	c.Status(200)
}

// LikeCar godoc
// @Summary Поставить или убрать лайк
// @Tags Автомобили
// @Produce json
// @Accept json
// @Param body body LikeRequest true "0 отменяет, 1 ставит лайк"
// @Param id path int true "ID автомобиля"
// @Success 200 {object} CarDTO
// @Failure 404
// @Failure 500
// @Security BearerAuth
// @Failure 403
// @Failure 503
// @Router /fuel_consumptions/{id}/likes [post]
func (s *Server) LikeCar(c *gin.Context) {
	i, ok := id(c, "id")
	if !ok {
		return
	}
	var b struct {
		Liked *int `json:"liked"`
	}
	if !readJSON(c, &b) {
		return
	}
	if b.Liked == nil || (*b.Liked != 0 && *b.Liked != 1) {
		c.Status(404)
		return
	}
	var car model.Car
	e := s.DB.Transaction(func(tx *gorm.DB) error {
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND status = ?", i, "published").First(&car).Error; e != nil {
			return e
		}
		if *b.Liked == 0 {
			return tx.Where("car_id = ? AND user_id = ?", i, currentUserID(c)).Delete(&model.CarLike{}).Error
		}
		like := model.CarLike{CarID: i, UserID: currentUserID(c)}
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&like).Error
	})
	if e != nil {
		fail(c, e)
		return
	}
	s.respondCar(c, car, 200)
}
