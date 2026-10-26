package api

import (
	"github.com/arseniy9606/cruisecontrol---back/internal/model"
	"github.com/arseniy9606/cruisecontrol---back/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *Server) AddCar(c *gin.Context) {
	carID, ok := id(c, "carID")
	if !ok || !emptyBody(c) {
		return
	}
	var requestID uint
	e := s.DB.Transaction(func(tx *gorm.DB) error {
		// Serialize draft creation for the fixed creator, including two simultaneous requests.
		var u model.User
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&u, service.CurrentUser().CreatorID).Error; e != nil {
			return e
		}
		var car model.Car
		if e := tx.Where("status = ?", "published").First(&car, carID).Error; e != nil {
			return e
		}
		var r model.FuelConsumption
		e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("creator_id = ? AND status = ?", u.ID, "draft").First(&r).Error
		if e == gorm.ErrRecordNotFound {
			r = model.FuelConsumption{CreatorID: u.ID, Status: "draft"}
			e = tx.Create(&r).Error
		}
		if e != nil {
			return e
		}
		requestID = r.ID
		return tx.Create(&model.TripCar{RequestID: r.ID, CarID: carID}).Error
	})
	if e != nil {
		fail(c, e)
		return
	}
	s.respond(c, requestID, 201)
}
func (s *Server) UpdateItem(c *gin.Context) {
	i, ok := id(c, "id")
	if !ok {
		return
	}
	car, ok := id(c, "carID")
	if !ok {
		return
	}
	var b struct {
		DistanceKM *float64 `json:"distance_km"`
	}
	if !readJSON(c, &b) {
		return
	}
	if b.DistanceKM == nil || *b.DistanceKM <= 0 || *b.DistanceKM > 1000000 {
		c.Status(404)
		return
	}
	e := s.DB.Transaction(func(tx *gorm.DB) error {
		if _, e := draft(tx, i); e != nil {
			return e
		}
		q := tx.Model(&model.TripCar{}).Where("request_id = ? AND car_id = ?", i, car).Update("distance_km", b.DistanceKM)
		if q.Error != nil {
			return q.Error
		}
		if q.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
	if e != nil {
		fail(c, e)
		return
	}
	s.respond(c, i, 200)
}
func (s *Server) DeleteItem(c *gin.Context) {
	i, ok := id(c, "id")
	if !ok {
		return
	}
	car, ok := id(c, "carID")
	if !ok || !emptyBody(c) {
		return
	}
	e := s.DB.Transaction(func(tx *gorm.DB) error {
		if _, e := draft(tx, i); e != nil {
			return e
		}
		q := tx.Where("request_id = ? AND car_id = ?", i, car).Delete(&model.TripCar{})
		if q.Error != nil {
			return q.Error
		}
		if q.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
	if e != nil {
		fail(c, e)
		return
	}
	s.respond(c, i, 200)
}
