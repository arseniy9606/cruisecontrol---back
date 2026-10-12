package api

import (
	"fmt"
	"github.com/arseniy9606/cruisecontrol---back/internal/app/fuel_consumptions"
	"github.com/arseniy9606/cruisecontrol---back/internal/app/fuel_consumptions_handler"
	"github.com/gin-gonic/gin"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"
)

func NewFuelConsumptionRouter(projectRoot, mediaBase string) *gin.Engine {
	return newFuelConsumptionRouter(projectRoot, mediaBase, "", true)
}

func newFuelConsumptionRouter(projectRoot, mediaBase, likesFile string, likesEnabled bool) *gin.Engine {
	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.SetFuncMap(template.FuncMap{"fuel_consumptionsFuel": func(n float64) string { return strings.ReplaceAll(fmt.Sprintf("%.1f", n), ".", ",") }})
	router.LoadHTMLGlob(projectRoot + "/templates/*.html")
	router.Static("/fuel_consumptions-assets", projectRoot+"/resources")
	repository := fuel_consumptions.NewFuelConsumptionRepository(mediaBase)
	if err := repository.UseLikesFile(likesFile); err != nil {
		panic(fmt.Errorf("load likes: %w", err))
	}
	handler := fuel_consumptions_handler.NewFuelConsumptionHandler(repository)
	handler.LikesEnabled = likesEnabled
	// Ровно три GET-маршрута страниц. Статика обслуживается отдельно.
	router.GET("/fuel_consumptions", handler.GetFuelConsumptionTiles)
	router.GET("/fuel_consumptions/draft", handler.GetFuelConsumptionDraft)
	// Wildcard позволяет открывать ленту без ID из нижней панели вкладок.
	router.GET("/fuel_consumptions/feed/*fuel_consumptionsID", handler.GetFuelConsumptionFeed)
	if likesEnabled {
		router.POST("/fuel_consumptions/:id/like", handler.PostFuelConsumptionLike)
	}
	return router
}
func StartFuelConsumptionServer() error {
	mediaBase := os.Getenv("FUEL_CONSUMPTIONS_MEDIA_URL")
	if mediaBase == "" {
		mediaBase = "http://localhost:9000/fuel-consumptions"
	}
	address := os.Getenv("FUEL_CONSUMPTIONS_LISTEN")
	if address == "" {
		address = "127.0.0.1:8080"
	}
	server := &http.Server{Addr: address, Handler: newFuelConsumptionRouter(".", mediaBase, ".runtime/fuel_consumptions_likes.json", os.Getenv("FUEL_CONSUMPTIONS_STRICT_LAB1") != "true"), ReadHeaderTimeout: 5 * time.Second}
	return server.ListenAndServe()
}
