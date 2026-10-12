package fuel_consumptions_handler

import (
	"crypto/rand"
	"encoding/binary"
	"github.com/arseniy9606/cruisecontrol---front/internal/app/fuel_consumptions"
	"github.com/gin-gonic/gin"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type FuelConsumptionHandler struct {
	FuelConsumptionRepository *fuel_consumptions.FuelConsumptionRepository
	LikesEnabled              bool
}
type FuelConsumptionCard struct {
	fuel_consumptions.FuelConsumption
	FuelConsumptionLikeCount    int
	FuelConsumptionImageURL     string
	FuelConsumptionVideoURL     string
	FuelConsumptionLiked        bool
	FuelConsumptionLikesEnabled bool
}

func NewFuelConsumptionHandler(r *fuel_consumptions.FuelConsumptionRepository) *FuelConsumptionHandler {
	return &FuelConsumptionHandler{FuelConsumptionRepository: r, LikesEnabled: true}
}
func (h *FuelConsumptionHandler) fuel_consumptionsCard(car fuel_consumptions.FuelConsumption, ctx *gin.Context) FuelConsumptionCard {
	// Количество лайков вычисляет контроллер, оно не хранится отдельным полем модели.
	visitor, _ := ctx.Cookie("fuel_consumptions_visitor")
	id, _ := strconv.Atoi(visitor)
	liked := false
	for _, v := range car.FuelConsumptionLikedBy {
		if v == id {
			liked = true
		}
	}
	return FuelConsumptionCard{car, len(car.FuelConsumptionLikedBy), h.FuelConsumptionRepository.FuelConsumptionMediaURL(car.FuelConsumptionImageKey), h.FuelConsumptionRepository.FuelConsumptionMediaURL(car.FuelConsumptionVideoKey), liked, h.LikesEnabled}
}
func fuel_consumptionsFuelBound(raw string) (float64, bool) {
	n, err := strconv.ParseFloat(raw, 64)
	return n, err == nil && !math.IsNaN(n) && !math.IsInf(n, 0) && n >= 0 && n <= 60
}

func (h *FuelConsumptionHandler) GetFuelConsumptionTiles(ctx *gin.Context) {
	minRaw, maxRaw := ctx.Query("fuel_consumptions_fuel_min"), ctx.Query("fuel_consumptions_fuel_max")
	minFuel, maxFuel := 0.0, 60.0
	valid, message := true, ""
	if minRaw != "" {
		minFuel, valid = fuel_consumptionsFuelBound(minRaw)
	}
	if maxRaw != "" {
		var ok bool
		maxFuel, ok = fuel_consumptionsFuelBound(maxRaw)
		valid = valid && ok
	}
	if !valid || minFuel > maxFuel {
		message = "Укажите расход от 0 до 60 л/100 км: нижняя граница не должна превышать верхнюю."
	}
	cards := make([]FuelConsumptionCard, 0)
	if message == "" {
		for _, car := range h.FuelConsumptionRepository.PublishedFuelConsumptions() {
			if car.FuelConsumptionRate >= minFuel && car.FuelConsumptionRate <= maxFuel {
				cards = append(cards, h.fuel_consumptionsCard(car, ctx))
			}
		}
	}
	if maxRaw == "" {
		maxRaw = "60"
	}
	status := http.StatusOK
	if message != "" {
		status = http.StatusBadRequest
	}
	ctx.HTML(status, "fuel_consumptions_tiles.html", gin.H{"FuelConsumptionCards": cards, "FuelConsumptionFuelMin": minRaw, "FuelConsumptionFuelMax": maxRaw, "FuelConsumptionFilterError": message, "FuelConsumptionFilterActive": ctx.Request.URL.RawQuery != ""})
}
func (h *FuelConsumptionHandler) GetFuelConsumptionFeed(ctx *gin.Context) {
	cars := h.FuelConsumptionRepository.PublishedFuelConsumptions()
	idRaw := strings.TrimPrefix(ctx.Param("fuel_consumptionsID"), "/")
	selected := 0
	if idRaw != "" {
		id, err := strconv.Atoi(idRaw)
		selected = -1
		if err == nil {
			for i, car := range cars {
				if car.FuelConsumptionID == id {
					selected = i
					break
				}
			}
		}
	}
	if selected < 0 || len(cars) == 0 {
		ctx.HTML(http.StatusNotFound, "fuel_consumptions_feed.html", gin.H{"FuelConsumptionError": "Автомобиль не найден"})
		return
	}
	if ctx.Query("next") == "true" {
		selected = (selected + 1) % len(cars)
	}
	ctx.HTML(http.StatusOK, "fuel_consumptions_feed.html", gin.H{"FuelConsumptionCard": h.fuel_consumptionsCard(cars[selected], ctx)})
}
func (h *FuelConsumptionHandler) GetFuelConsumptionDraft(ctx *gin.Context) {
	car, ok := h.FuelConsumptionRepository.DraftFuelConsumption()
	if !ok {
		ctx.HTML(http.StatusNotFound, "fuel_consumptions_draft.html", gin.H{"FuelConsumptionError": "Черновик не найден"})
		return
	}
	ctx.HTML(http.StatusOK, "fuel_consumptions_draft.html", gin.H{"FuelConsumptionCard": h.fuel_consumptionsCard(car, ctx)})
}

// POST changes state; GET remains safe. SameSite and Origin checking prevent cross-site forms.
func (h *FuelConsumptionHandler) PostFuelConsumptionLike(ctx *gin.Context) {
	if origin := ctx.GetHeader("Origin"); origin != "" {
		u, err := url.Parse(origin)
		if err != nil || u.Host != ctx.Request.Host {
			ctx.String(http.StatusForbidden, "Недопустимый источник запроса")
			return
		}
	}
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.String(http.StatusNotFound, "Автомобиль не найден")
		return
	}
	value, _ := ctx.Cookie("fuel_consumptions_visitor")
	visitor, _ := strconv.Atoi(value)
	if visitor < 1000 {
		var raw [8]byte
		if _, err = rand.Read(raw[:]); err != nil {
			ctx.String(http.StatusInternalServerError, "Не удалось создать идентификатор")
			return
		}
		visitor = int(binary.BigEndian.Uint64(raw[:])&((1<<62)-1)) + 1000
		ctx.SetSameSite(http.SameSiteStrictMode)
		ctx.SetCookie("fuel_consumptions_visitor", strconv.Itoa(visitor), 365*24*60*60, "/fuel_consumptions", "", ctx.Request.TLS != nil, true)
	}
	if _, err = h.FuelConsumptionRepository.ToggleLike(id, visitor); err != nil {
		if err == fuel_consumptions.ErrFuelConsumptionNotFound {
			ctx.String(http.StatusNotFound, "Автомобиль не найден")
		} else {
			ctx.String(http.StatusInternalServerError, "Не удалось сохранить лайк")
		}
		return
	}
	target := "/fuel_consumptions/feed/" + strconv.Itoa(id)
	if ctx.PostForm("view") == "tiles" {
		target = "/fuel_consumptions"
	}
	ctx.Redirect(http.StatusSeeOther, target)
}
