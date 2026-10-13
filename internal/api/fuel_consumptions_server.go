package api

import (
	"context"
	"fmt"
	fc "github.com/arseniy9606/cruisecontrol---back/internal/app/fuel_consumptions"
	"github.com/arseniy9606/cruisecontrol---back/internal/app/fuel_consumptions_handler"
	"github.com/gin-gonic/gin"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func NewFuelConsumptionRouter(projectRoot string, repository *fc.FuelConsumptionRepository, media *fc.FuelConsumptionMedia) *gin.Engine {
	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.SetFuncMap(template.FuncMap{"fuel_consumptionsFuel": func(n float64) string { return strings.ReplaceAll(fmt.Sprintf("%.1f", n), ".", ",") }})
	router.LoadHTMLGlob(projectRoot + "/templates/*.html")
	router.Static("/fuel_consumptions-assets", projectRoot+"/resources")
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
	h := &fuel_consumptions_handler.FuelConsumptionHandler{FuelConsumptionRepository: repository, Media: media}
	// Exactly six controllers; static assets are separate routes.
	router.GET("/fuel_consumptions", h.GetFuelConsumptionTiles)
	router.GET("/fuel_consumptions/draft", h.GetFuelConsumptionDraft)
	router.GET("/fuel_consumptions/feed/*fuel_consumptionsID", h.GetFuelConsumptionFeed)
	router.POST("/fuel_consumptions/create", h.PostFuelConsumptionCreate)
	router.POST("/fuel_consumptions/publish", h.PostFuelConsumptionPublish)
	router.POST("/fuel_consumptions/:fuel_consumptionsDeleteID/delete", h.PostFuelConsumptionDelete)
	return router
}
func StartFuelConsumptionServer() error {
	r, err := fc.OpenFuelConsumptionRepository(fc.FuelConsumptionDSN())
	if err != nil {
		return err
	}
	defer r.Close()
	if err = r.ValidateDriver(context.Background()); err != nil {
		return err
	}
	public := fc.FuelConsumptionEnv("FUEL_CONSUMPTIONS_MEDIA_URL", "http://localhost:9100/fuel-consumptions")
	media := fc.NewFuelConsumptionMedia(public, fc.FuelConsumptionEnv("FUEL_CONSUMPTIONS_MEDIA_PROBE_URL", public))
	server := &http.Server{Addr: fc.FuelConsumptionEnv("FUEL_CONSUMPTIONS_LISTEN", "127.0.0.1:8082"), Handler: NewFuelConsumptionRouter(".", r, media), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	return server.ListenAndServe()
}
