package api

import (
	"errors"
	"github.com/arseniy9606/cruisecontrol---back/internal/auth"
	"github.com/arseniy9606/cruisecontrol---back/internal/model"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"strings"
)

type LoginRequest struct {
	Login    string `json:"login" example:"creator"`
	Password string `json:"password" example:"Lab4Demo123"`
}
type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}
type RegisterRequest struct {
	Login    string `json:"login"`
	FullName string `json:"full_name"`
	Password string `json:"password"`
}
type UserResponse struct {
	ID       uint   `json:"id"`
	Login    string `json:"login"`
	FullName string `json:"full_name"`
}

func currentUserID(c *gin.Context) uint {
	if v, ok := c.Get("user_id"); ok {
		return v.(uint)
	}
	return 0
}

// Public methods remain readable even with an absent, expired or malformed token.
// Redis errors never grant creator permissions.
func (s *Server) identify(c *gin.Context) {
	parts := strings.Fields(c.GetHeader("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		c.Next()
		return
	}
	claims, e := s.Auth.Parse(parts[1])
	if e != nil {
		c.Next()
		return
	}
	revoked, e := s.Auth.Revoked(c.Request.Context(), parts[1])
	if e != nil {
		c.Set("auth_unavailable", true)
		c.Next()
		return
	}
	if revoked {
		c.Next()
		return
	}
	var u model.User
	if e = s.DB.First(&u, "id = ?", claims.Subject).Error; e != nil {
		if !errors.Is(e, gorm.ErrRecordNotFound) {
			c.Set("auth_unavailable", true)
		}
		c.Next()
		return
	}
	c.Set("user_id", u.ID)
	c.Set("token", parts[1])
	c.Set("claims", claims)
	c.Next()
}
func requireUser(c *gin.Context) {
	if currentUserID(c) == 0 {
		if _, ok := c.Get("auth_unavailable"); ok {
			c.AbortWithStatus(503)
		} else {
			c.AbortWithStatus(403)
		}
		return
	}
	c.Next()
}

// Login godoc
// @Summary Вход по логину и паролю
// @Tags Водители
// @Produce json
// @Accept json
// @Param body body LoginRequest true "Учётные данные"
// @Success 200 {object} LoginResponse
// @Failure 403
// @Failure 503
// @Failure 404
// @Failure 500
// @Router /users/login [post]
func (s *Server) Login(c *gin.Context) {
	var b LoginRequest
	if !readJSON(c, &b) {
		return
	}
	var u model.User
	e := s.DB.Where("login = ?", b.Login).First(&u).Error
	if e != nil && !errors.Is(e, gorm.ErrRecordNotFound) {
		fail(c, e)
		return
	}
	if e != nil || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(b.Password)) != nil {
		c.Status(403)
		return
	}
	if e = s.Auth.Redis.Ping(c.Request.Context()).Err(); e != nil {
		c.Status(503)
		return
	}
	token, e := s.Auth.Issue(u.ID)
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(200, LoginResponse{token, "Bearer", int(s.Auth.TTL.Seconds())})
}

// Logout godoc
// @Summary Выход: отзыв JWT в Redis
// @Tags Водители
// @Produce json
// @Success 200
// @Failure 404
// @Failure 500
// @Security BearerAuth
// @Failure 403
// @Failure 503
// @Router /users/logout [post]
func (s *Server) Logout(c *gin.Context) {
	raw := c.MustGet("token").(string)
	claims := c.MustGet("claims").(*auth.Claims)
	if e := s.Auth.Revoke(c.Request.Context(), raw, claims); e != nil {
		c.Status(503)
		return
	}
	c.Status(200)
}

type PublicationRequest struct {
	Description     string  `json:"description"`
	ConsumptionRate float64 `json:"consumption_rate"`
	FuelTankVolume  float64 `json:"fuel_tank_volume"`
}
type LikeRequest struct {
	Liked int `json:"liked" enums:"0,1"`
}
