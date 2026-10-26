package api

import (
	"github.com/arseniy9606/cruisecontrol---back/internal/model"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"regexp"
	"strings"
)

func (s *Server) Register(c *gin.Context) {
	var b struct {
		Login    string `json:"login"`
		FullName string `json:"full_name"`
		Password string `json:"password"`
	}
	if !readJSON(c, &b) {
		return
	}
	b.FullName = strings.TrimSpace(b.FullName)
	if !regexp.MustCompile(`^[a-zA-Z0-9_]{3,60}$`).MatchString(b.Login) || len(b.Password) < 8 || len(b.Password) > 72 || b.FullName == "" || len([]rune(b.FullName)) > 120 {
		c.Status(404)
		return
	}
	hash, e := bcrypt.GenerateFromPassword([]byte(b.Password), bcrypt.DefaultCost)
	if e != nil {
		fail(c, e)
		return
	}
	u := model.User{Login: b.Login, FullName: b.FullName, PasswordHash: string(hash)}
	if e = s.DB.Create(&u).Error; e != nil {
		fail(c, e)
		return
	}
	c.JSON(201, gin.H{"id": u.ID, "login": u.Login, "full_name": u.FullName})
}

// Required placeholders: these do not authenticate or issue tokens.
func (s *Server) AuthStub(c *gin.Context) { c.Status(200) }
