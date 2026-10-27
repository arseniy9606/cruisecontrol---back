package api

import (
	"encoding/json"
	"errors"
	"github.com/arseniy9606/cruisecontrol---back/internal/service"
	"github.com/arseniy9606/cruisecontrol---back/internal/storage"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"strconv"
)

type Server struct {
	DB    *gorm.DB
	Media *storage.Media
}

func (s *Server) Router() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.HandleMethodNotAllowed = true
	r.NoRoute(func(c *gin.Context) { c.Status(404) })
	r.NoMethod(func(c *gin.Context) { c.Status(404) })
	a := r.Group("/api")
	a.GET("/fuel_consumptions", s.ListCars)
	a.GET("/fuel_consumptions/feed", s.Feed)
	a.GET("/fuel_consumptions/draft", s.GetDraft)
	a.POST("/fuel_consumptions", s.CreateCar)
	a.PUT("/fuel_consumptions/:id/publication", s.PublishCar)
	a.DELETE("/fuel_consumptions/:id", s.DeleteCar)
	a.POST("/fuel_consumptions/:id/likes", s.LikeCar)
	a.POST("/users", s.Register)
	a.POST("/users/login", s.AuthStub)
	a.POST("/users/logout", s.AuthStub)
	return r
}
func fail(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, service.ErrInvalid) || errors.Is(err, gorm.ErrDuplicatedKey) {
		c.Status(404)
		return
	}
	log.Printf("request failed: %v", err)
	c.Status(500)
}
func readJSON(c *gin.Context, v any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	d := json.NewDecoder(c.Request.Body)
	d.DisallowUnknownFields()
	if d.Decode(v) != nil || d.Decode(&struct{}{}) != io.EOF {
		c.Status(404)
		return false
	}
	return true
}
func id(c *gin.Context, key string) (uint, bool) {
	v, e := strconv.ParseUint(c.Param(key), 10, 32)
	if e != nil || v == 0 {
		c.Status(404)
		return 0, false
	}
	return uint(v), true
}
