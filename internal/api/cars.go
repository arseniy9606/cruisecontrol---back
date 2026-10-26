package api

import (
	"github.com/arseniy9606/cruisecontrol---back/internal/model"
	"github.com/arseniy9606/cruisecontrol---back/internal/service"
	"github.com/gin-gonic/gin"
	"io"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

func (s *Server) ListCars(c *gin.Context) {
	q := s.DB.Where("status = ?", "published")
	min, max := 0.0, math.MaxFloat64
	for _, k := range []string{"consumption_min", "consumption_max"} {
		if raw, ok := c.GetQuery(k); ok {
			v, e := strconv.ParseFloat(raw, 64)
			if e != nil || math.IsNaN(v) || math.IsInf(v, 0) || v < 0 {
				c.Status(404)
				return
			}
			if k == "consumption_min" {
				min = v
			} else {
				max = v
			}
		}
	}
	if min > max {
		c.Status(404)
		return
	}
	cars := []model.Car{}
	if e := q.Where("consumption_rate BETWEEN ? AND ?", min, max).Order("id").Find(&cars).Error; e != nil {
		fail(c, e)
		return
	}
	out := []CarDTO{}
	for _, x := range cars {
		out = append(out, s.carDTO(x))
	}
	c.JSON(200, out)
}
func (s *Server) GetCar(c *gin.Context) {
	i, ok := id(c, "id")
	if !ok {
		return
	}
	var x model.Car
	if e := s.DB.Where("status = ?", "published").First(&x, i).Error; e != nil {
		fail(c, e)
		return
	}
	c.JSON(200, s.carDTO(x))
}
func (s *Server) CreateCar(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 32<<20)
	if e := c.Request.ParseMultipartForm(2 << 20); e != nil {
		c.Status(404)
		return
	}
	defer c.Request.MultipartForm.RemoveAll()
	allowed := map[string]bool{"name": true, "description": true, "consumption_rate": true, "fuel_tank_volume": true}
	for k := range c.Request.MultipartForm.Value {
		if !allowed[k] {
			c.Status(404)
			return
		}
	}
	for k := range c.Request.MultipartForm.File {
		if k != "image" && k != "video" {
			c.Status(404)
			return
		}
	}
	rate, e := strconv.ParseFloat(c.PostForm("consumption_rate"), 64)
	tank, e2 := strconv.ParseFloat(c.PostForm("fuel_tank_volume"), 64)
	name := strings.TrimSpace(c.PostForm("name"))
	desc := c.PostForm("description")
	if e != nil || e2 != nil || rate <= 0 || rate > 99999 || tank <= 0 || tank > 99999 || math.IsNaN(rate) || math.IsNaN(tank) || name == "" || len([]rune(name)) > 120 || len([]rune(desc)) > 1000 {
		c.Status(404)
		return
	}
	car := model.Car{Name: name, Description: desc, Status: "published", ConsumptionRate: rate, FuelTankVolume: tank}
	success := false
	defer func() {
		if !success {
			s.Media.Remove(c.Request.Context(), car.ImageName)
			s.Media.Remove(c.Request.Context(), car.VideoName)
		}
	}()
	for _, field := range []string{"image", "video"} {
		files := c.Request.MultipartForm.File[field]
		if len(files) != 1 {
			c.Status(404)
			return
		}
		f := files[0]
		stream, e := f.Open()
		if e != nil {
			fail(c, e)
			return
		}
		header := make([]byte, 512)
		n, _ := stream.Read(header)
		_, _ = stream.Seek(0, io.SeekStart)
		ct := http.DetectContentType(header[:n])
		ext := strings.ToLower(filepath.Ext(f.Filename))
		valid := field == "image" && ((ext == ".jpg" || ext == ".jpeg") && ct == "image/jpeg" || ext == ".png" && ct == "image/png") || field == "video" && ext == ".mp4" && ct == "video/mp4"
		if !valid {
			stream.Close()
			fail(c, service.ErrInvalid)
			return
		}
		key, e := s.Media.Upload(c.Request.Context(), stream, f.Size, ext, ct)
		stream.Close()
		if e != nil {
			fail(c, e)
			return
		}
		if field == "image" {
			car.ImageName = key
		} else {
			car.VideoName = key
		}
	}
	if e := s.DB.Create(&car).Error; e != nil {
		fail(c, e)
		return
	}
	success = true
	c.JSON(201, s.carDTO(car))
}
