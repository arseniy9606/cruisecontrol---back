package main

import (
	"github.com/arseniy9606/cruisecontrol---back/internal/api"
	"github.com/arseniy9606/cruisecontrol---back/internal/auth"
	"github.com/arseniy9606/cruisecontrol---back/internal/storage"
	"log"
	"net/http"
	"os"
	"time"
)

// @title Расход топлива: API ЛР4
// @version 4.0
// @description Автомобили, публикация, лайки. JWT + Redis blacklist.
// @host localhost:8086
// @BasePath /api
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите Bearer и JWT из POST /api/users/login.
func main() {
	db, e := storage.Open()
	if e != nil {
		log.Fatal(e)
	}
	media, e := storage.NewMedia()
	if e != nil {
		log.Fatal(e)
	}
	ttl, e := time.ParseDuration(os.Getenv("JWT_TTL"))
	if e != nil {
		log.Fatal(e)
	}
	manager, e := auth.New(os.Getenv("JWT_SECRET"), os.Getenv("REDIS_ADDR"), ttl)
	if e != nil {
		log.Fatal(e)
	}
	defer manager.Redis.Close()
	s := api.Server{DB: db, Media: media, Auth: manager}
	server := http.Server{Addr: ":8080", Handler: s.Router(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 60 * time.Second, WriteTimeout: 60 * time.Second}
	log.Fatal(server.ListenAndServe())
}
