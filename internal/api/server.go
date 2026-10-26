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
	a.GET("/cars", s.ListCars)
	a.GET("/cars/:id", s.GetCar)
	a.POST("/cars", s.CreateCar)
	a.POST("/fuel_consumptions/draft/cars/:carID", s.AddCar)
	a.PUT("/fuel_consumptions/:id/cars/:carID", s.UpdateItem)
	a.DELETE("/fuel_consumptions/:id/cars/:carID", s.DeleteItem)
	a.GET("/fuel_consumptions/cart", s.Cart)
	a.GET("/fuel_consumptions", s.ListRequests)
	a.GET("/fuel_consumptions/:id", s.GetRequest)
	a.PUT("/fuel_consumptions/:id", s.UpdateRequest)
	a.PUT("/fuel_consumptions/:id/formation", s.FormRequest)
	a.PUT("/fuel_consumptions/:id/completion", s.CompleteRequest)
	a.DELETE("/fuel_consumptions/:id", s.DeleteRequest)
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
