package api

import (
	"github.com/arseniy9606/cruisecontrol---back/internal/model"
	"github.com/arseniy9606/cruisecontrol---back/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
	"time"
)

func loaded(db *gorm.DB) *gorm.DB {
	return db.Preload("Creator").Preload("Moderator").Preload("Items", func(q *gorm.DB) *gorm.DB { return q.Order("car_id") }).Preload("Items.Car")
}
func (s *Server) respond(c *gin.Context, i uint, code int) {
	var r model.FuelConsumption
	if e := loaded(s.DB).Where("status <> ?", "deleted").First(&r, i).Error; e != nil {
		fail(c, e)
		return
	}
	c.JSON(code, s.requestDTO(r))
}
func (s *Server) Cart(c *gin.Context) {
	var r model.FuelConsumption
	e := s.DB.Where("creator_id = ? AND status = ?", service.CurrentUser().CreatorID, "draft").Preload("Items").First(&r).Error
	if e == gorm.ErrRecordNotFound {
		c.JSON(200, gin.H{"id": nil, "count": 0})
		return
	}
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(200, gin.H{"id": r.ID, "count": len(r.Items)})
}
func (s *Server) ListRequests(c *gin.Context) {
	q := loaded(s.DB).Where("status NOT IN ?", []string{"draft", "deleted"})
	if status := c.Query("status"); status != "" {
		if status != "formed" && status != "completed" && status != "rejected" {
			c.Status(404)
			return
		}
		q = q.Where("status = ?", status)
	}
	var from, to *time.Time
	for _, k := range []string{"formed_from", "formed_to"} {
		if raw, ok := c.GetQuery(k); ok {
			t, e := time.Parse("2006-01-02", raw)
			if e != nil {
				c.Status(404)
				return
			}
			if k == "formed_from" {
				from = &t
				q = q.Where("formed_at >= ?", t)
			} else {
				to = &t
				q = q.Where("formed_at < ?", t.AddDate(0, 0, 1))
			}
		}
	}
	if from != nil && to != nil && from.After(*to) {
		c.Status(404)
		return
	}
	rows := []model.FuelConsumption{}
	if e := q.Order("id").Find(&rows).Error; e != nil {
		fail(c, e)
		return
	}
	out := []RequestDTO{}
	for _, r := range rows {
		out = append(out, s.requestDTO(r))
	}
	c.JSON(200, out)
}
func (s *Server) GetRequest(c *gin.Context) {
	i, ok := id(c, "id")
	if ok {
		s.respond(c, i, 200)
	}
}
func draft(tx *gorm.DB, i uint) (*model.FuelConsumption, error) {
	var r model.FuelConsumption
	e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND status = ? AND creator_id = ?", i, "draft", service.CurrentUser().CreatorID).First(&r).Error
	return &r, e
}
func (s *Server) UpdateRequest(c *gin.Context) {
	i, ok := id(c, "id")
	if !ok {
		return
	}
	var body struct {
		Route         string   `json:"route"`
		SavingPercent *float64 `json:"saving_percent"`
	}
	if !readJSON(c, &body) {
		return
	}
	body.Route = strings.TrimSpace(body.Route)
	if len([]rune(body.Route)) > 300 || body.SavingPercent != nil && (*body.SavingPercent < 0 || *body.SavingPercent >= 100) {
		c.Status(404)
		return
	}
	e := s.DB.Transaction(func(tx *gorm.DB) error {
		r, e := draft(tx, i)
		if e != nil {
			return e
		}
		return tx.Model(r).Updates(map[string]any{"route": body.Route, "saving_percent": body.SavingPercent}).Error
	})
	if e != nil {
		fail(c, e)
		return
	}
	s.respond(c, i, 200)
}
func emptyBody(c *gin.Context) bool {
	if c.Request.ContentLength != 0 {
		var b struct{}
		return readJSON(c, &b)
	}
	return true
}
func (s *Server) FormRequest(c *gin.Context) {
	i, ok := id(c, "id")
	if !ok || !emptyBody(c) {
		return
	}
	e := s.DB.Transaction(func(tx *gorm.DB) error {
		r, e := draft(tx, i)
		if e != nil {
			return e
		}
		if e = loaded(tx).First(r, i).Error; e != nil {
			return e
		}
		if e = service.Calculate(r); e != nil {
			return e
		}
		for _, x := range r.Items {
			if e = tx.Model(&model.TripCar{}).Where("request_id = ? AND car_id = ?", i, x.CarID).Updates(map[string]any{"base_liters": x.BaseLiters, "cruise_liters": x.CruiseLiters, "saved_liters": x.SavedLiters, "range_km": x.RangeKM}).Error; e != nil {
				return e
			}
		}
		return tx.Model(&model.FuelConsumption{}).Where("id = ?", i).Updates(map[string]any{"status": "formed", "formed_at": time.Now().UTC(), "total_saved_liters": r.TotalSavedLiters}).Error
	})
	if e != nil {
		fail(c, e)
		return
	}
	s.respond(c, i, 200)
}
func (s *Server) CompleteRequest(c *gin.Context) {
	i, ok := id(c, "id")
	if !ok {
		return
	}
	var b struct {
		Decision string `json:"decision"`
	}
	if !readJSON(c, &b) {
		return
	}
	if b.Decision != "complete" && b.Decision != "reject" {
		c.Status(404)
		return
	}
	e := s.DB.Transaction(func(tx *gorm.DB) error {
		var r model.FuelConsumption
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND status = ?", i, "formed").First(&r).Error; e != nil {
			return e
		}
		state := "completed"
		if b.Decision == "reject" {
			state = "rejected"
		}
		return tx.Model(&r).Updates(map[string]any{"status": state, "moderator_id": service.CurrentUser().ModeratorID, "completed_at": time.Now().UTC()}).Error
	})
	if e != nil {
		fail(c, e)
		return
	}
	s.respond(c, i, 200)
}
func (s *Server) DeleteRequest(c *gin.Context) {
	i, ok := id(c, "id")
	if !ok || !emptyBody(c) {
		return
	}
	e := s.DB.Transaction(func(tx *gorm.DB) error {
		r, e := draft(tx, i)
		if e != nil {
			return e
		}
		return tx.Model(r).Updates(map[string]any{"status": "deleted", "formed_at": time.Now().UTC()}).Error
	})
	if e != nil {
		fail(c, e)
		return
	}
	c.Status(200)
}
