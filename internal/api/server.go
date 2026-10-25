package api

import (
	"fmt"
	"github.com/arseniy9606/cruisecontrol---back/internal/app/car_handler"
	"github.com/arseniy9606/cruisecontrol---back/internal/app/cars"
	"github.com/gin-gonic/gin"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"
)

func NewCarRouter(projectRoot, mediaBase string) *gin.Engine {
	return newCarRouter(projectRoot, mediaBase, "", true)
}

func newCarRouter(projectRoot, mediaBase, likesFile string, likesEnabled bool) *gin.Engine {
	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.SetFuncMap(template.FuncMap{"carsFuel": func(n float64) string { return strings.ReplaceAll(fmt.Sprintf("%.1f", n), ".", ",") }})
	router.LoadHTMLGlob(projectRoot + "/templates/*.html")
	router.Static("/cars-assets", projectRoot+"/resources")
	repository := cars.NewCarRepository(mediaBase)
	if err := repository.UseLikesFile(likesFile); err != nil {
		panic(fmt.Errorf("load likes: %w", err))
	}
	handler := car_handler.NewCarHandler(repository)
	handler.LikesEnabled = likesEnabled
	// Ровно три GET-маршрута страниц. Статика обслуживается отдельно.
	router.GET("/cars", handler.GetCarTiles)
	router.GET("/cars/draft", handler.GetCarDraft)
	// Wildcard позволяет открывать ленту без ID из нижней панели вкладок.
	router.GET("/cars/feed/*carsID", handler.GetCarFeed)
	if likesEnabled {
		router.POST("/cars/:id/like", handler.PostCarLike)
	}
	return router
}
func StartCarServer() error {
	mediaBase := os.Getenv("CARS_MEDIA_URL")
	if mediaBase == "" {
		mediaBase = "http://localhost:9000/cars"
	}
	address := os.Getenv("CARS_LISTEN")
	if address == "" {
		address = "127.0.0.1:8080"
	}
	server := &http.Server{Addr: address, Handler: newCarRouter(".", mediaBase, ".runtime/cars_likes.json", os.Getenv("CARS_STRICT_LAB1") != "true"), ReadHeaderTimeout: 5 * time.Second}
	return server.ListenAndServe()
}
