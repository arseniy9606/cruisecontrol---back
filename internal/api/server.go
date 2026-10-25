package api

import (
	"context"
	"fmt"
	"github.com/arseniy9606/cruisecontrol---back/internal/app/car_handler"
	fc "github.com/arseniy9606/cruisecontrol---back/internal/app/cars"
	"github.com/gin-gonic/gin"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func NewCarRouter(projectRoot string, repository *fc.CarRepository, media *fc.CarMedia) *gin.Engine {
	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.SetFuncMap(template.FuncMap{"carsFuel": func(n float64) string { return strings.ReplaceAll(fmt.Sprintf("%.1f", n), ".", ",") }})
	router.LoadHTMLGlob(projectRoot + "/templates/*.html")
	router.Static("/cars-assets", projectRoot+"/resources")
	router.Use(func(ctx *gin.Context) {
		if ctx.Request.Method == http.MethodPost {
			if origin := ctx.GetHeader("Origin"); origin != "" {
				u, err := url.Parse(origin)
				if err != nil || u.Host != ctx.Request.Host {
					ctx.AbortWithStatus(403)
					return
				}
			}
			if !strings.HasPrefix(ctx.GetHeader("Content-Type"), "application/x-www-form-urlencoded") {
				ctx.AbortWithStatus(415)
				return
			}
			ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, 16*1024)
			if err := ctx.Request.ParseForm(); err != nil {
				ctx.AbortWithStatus(400)
				return
			}
		}
		ctx.Next()
	})
	h := &car_handler.CarHandler{CarRepository: repository, Media: media}
	// Exactly six controllers; static assets are separate routes.
	router.GET("/cars", h.GetCarTiles)
	router.GET("/cars/draft", h.GetCarDraft)
	router.GET("/cars/feed/*carsID", h.GetCarFeed)
	router.POST("/cars/create", h.PostCarCreate)
	router.POST("/cars/publish", h.PostCarPublish)
	router.POST("/cars/:carsDeleteID/delete", h.PostCarDelete)
	return router
}
func StartCarServer() error {
	r, err := fc.OpenCarRepository(fc.CarDSN())
	if err != nil {
		return err
	}
	defer r.Close()
	if err = r.ValidateDriver(context.Background()); err != nil {
		return err
	}
	public := fc.CarEnv("CARS_MEDIA_URL", "http://localhost:9100/cars")
	media := fc.NewCarMedia(public, fc.CarEnv("CARS_MEDIA_PROBE_URL", public))
	server := &http.Server{Addr: fc.CarEnv("CARS_LISTEN", "127.0.0.1:8082"), Handler: NewCarRouter(".", r, media), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	return server.ListenAndServe()
}
