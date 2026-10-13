package fuel_consumptions_handler

import (
	"errors"
	fc "github.com/arseniy9606/cruisecontrol---back/internal/app/fuel_consumptions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

type FuelConsumptionHandler struct {
	FuelConsumptionRepository *fc.FuelConsumptionRepository
	Media                     *fc.FuelConsumptionMedia
}
type FuelConsumptionCard struct {
	fc.FuelConsumption
	FuelConsumptionLikeCount                         int64
	FuelConsumptionImageURL, FuelConsumptionVideoURL string
}

func (h *FuelConsumptionHandler) card(ctx *gin.Context, car fc.FuelConsumption) (FuelConsumptionCard, error) {
	count, err := h.FuelConsumptionRepository.FuelConsumptionLikeCount(ctx.Request.Context(), car.FuelConsumptionID)
	return FuelConsumptionCard{car, count, h.Media.Resolve(ctx.Request.Context(), car.FuelConsumptionImageURL, false), h.Media.Resolve(ctx.Request.Context(), car.FuelConsumptionVideoURL, true)}, err
}
func bound(raw string, min, max float64) (float64, bool) {
	n, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(raw), ",", "."), 64)
	return n, err == nil && !math.IsNaN(n) && !math.IsInf(n, 0) && n >= min && n <= max
}
func (h *FuelConsumptionHandler) failure(ctx *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.HTML(404, "fuel_consumptions_feed.html", gin.H{"FuelConsumptionError": "Автомобиль не найден"})
		return
	}
	log.Printf("fuel consumptions: %v", err)
	ctx.String(500, "Не удалось выполнить операцию с базой данных. Проверьте журналы приложения.")
}

// 1. GET tiles: ORM SELECT, numeric range and status are filtered in SQL.
func (h *FuelConsumptionHandler) GetFuelConsumptionTiles(ctx *gin.Context) {
	minRaw, maxRaw := ctx.Query("fuel_consumptions_fuel_min"), ctx.Query("fuel_consumptions_fuel_max")
	min, max := 0.0, 60.0
	valid := true
	if minRaw != "" {
		min, valid = bound(minRaw, 0, 60)
	}
	if maxRaw != "" {
		var ok bool
		max, ok = bound(maxRaw, 0, 60)
		valid = valid && ok
	} else {
		maxRaw = "60"
	}
	cards := make([]FuelConsumptionCard, 0)
	message := ""
	status := 200
	if !valid || min > max {
		message = "Укажите расход от 0 до 60 л/100 км: нижняя граница не должна превышать верхнюю."
		status = 400
	} else {
		cars, err := h.FuelConsumptionRepository.PublishedFuelConsumptions(ctx.Request.Context(), min, max)
		if err != nil {
			h.failure(ctx, err)
			return
		}
		for _, car := range cars {
			card, err := h.card(ctx, car)
			if err != nil {
				h.failure(ctx, err)
				return
			}
			cards = append(cards, card)
		}
	}
	ctx.HTML(status, "fuel_consumptions_tiles.html", gin.H{"FuelConsumptionCards": cards, "FuelConsumptionFuelMin": minRaw, "FuelConsumptionFuelMax": maxRaw, "FuelConsumptionFilterError": message, "FuelConsumptionFilterActive": ctx.Request.URL.RawQuery != ""})
}

// 2. GET feed: PostgreSQL returns a single card, never a filtered array.
func (h *FuelConsumptionHandler) GetFuelConsumptionFeed(ctx *gin.Context) {
	raw := strings.TrimPrefix(ctx.Param("fuel_consumptionsID"), "/")
	var id int64
	if raw != "" {
		var err error
		id, err = strconv.ParseInt(raw, 10, 64)
		if err != nil || id < 1 {
			h.failure(ctx, gorm.ErrRecordNotFound)
			return
		}
	}
	car, err := h.FuelConsumptionRepository.FeedFuelConsumption(ctx.Request.Context(), id, ctx.Query("next") == "true")
	if err != nil {
		h.failure(ctx, err)
		return
	}
	card, err := h.card(ctx, car)
	if err != nil {
		h.failure(ctx, err)
		return
	}
	ctx.HTML(200, "fuel_consumptions_feed.html", gin.H{"FuelConsumptionCard": card})
}
func (h *FuelConsumptionHandler) draft(ctx *gin.Context, status int, message, name, description, tank, rate string) {
	car, err := h.FuelConsumptionRepository.DraftFuelConsumption(ctx.Request.Context())
	exists := err == nil
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		h.failure(ctx, err)
		return
	}
	card, err := h.card(ctx, car)
	if err != nil {
		h.failure(ctx, err)
		return
	}
	if name == "" {
		name = car.FuelConsumptionName
	}
	if message == "" {
		description = car.FuelConsumptionDescription
		if car.FuelTankVolume > 0 {
			tank = strconv.FormatFloat(car.FuelTankVolume, 'f', 1, 64)
		}
		if car.FuelConsumptionRate > 0 {
			rate = strconv.FormatFloat(car.FuelConsumptionRate, 'f', 1, 64)
		}
	}
	ctx.HTML(status, "fuel_consumptions_draft.html", gin.H{"FuelConsumptionCard": card, "FuelConsumptionHasDraft": exists, "FuelConsumptionError": message, "FuelConsumptionName": name, "FuelConsumptionDescription": description, "FuelConsumptionTank": tank, "FuelConsumptionRate": rate})
}

// 3. GET draft never creates database rows.
func (h *FuelConsumptionHandler) GetFuelConsumptionDraft(ctx *gin.Context) {
	h.draft(ctx, 200, "", "", "", "", "")
}

// 4. POST create: only a name, no file bytes or filenames are received.
func (h *FuelConsumptionHandler) PostFuelConsumptionCreate(ctx *gin.Context) {
	name := strings.TrimSpace(ctx.PostForm("fuel_consumptions_name"))
	if name == "" || utf8.RuneCountInString(name) > 120 {
		h.draft(ctx, 400, "Введите название автомобиля длиной от 1 до 120 символов.", name, "", "", "")
		return
	}
	err := h.FuelConsumptionRepository.CreateFuelConsumption(ctx.Request.Context(), name, h.Media.PublicBase)
	if errors.Is(err, fc.ErrFuelConsumptionDraftExists) {
		h.draft(ctx, 409, "Черновик уже существует. Завершите его публикацию.", "", "", "", "")
		return
	}
	if err != nil {
		h.failure(ctx, err)
		return
	}
	ctx.Redirect(303, "/fuel_consumptions/draft")
}

// 5. POST publish: status and timestamps are generated on the server.
func (h *FuelConsumptionHandler) PostFuelConsumptionPublish(ctx *gin.Context) {
	description := strings.TrimSpace(ctx.PostForm("fuel_consumptions_description"))
	tankRaw, rateRaw := ctx.PostForm("fuel_consumptions_tank_volume"), ctx.PostForm("fuel_consumptions_rate")
	tank, okTank := bound(tankRaw, 1, 1500)
	rate, okRate := bound(rateRaw, 0.1, 60)
	if description == "" || utf8.RuneCountInString(description) > 1000 || !okTank || !okRate {
		h.draft(ctx, 400, "Заполните описание (1–1000 символов), объём бака (1–1500 л) и расход (0,1–60 л/100 км).", "", description, tankRaw, rateRaw)
		return
	}
	err := h.FuelConsumptionRepository.PublishFuelConsumption(ctx.Request.Context(), description, tank, rate)
	if err != nil {
		h.failure(ctx, err)
		return
	}
	ctx.Redirect(303, "/fuel_consumptions")
}

// 6. POST delete: repository executes raw UPDATE through database/sql.
func (h *FuelConsumptionHandler) PostFuelConsumptionDelete(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("fuel_consumptionsDeleteID"), 10, 64)
	if err != nil || id < 1 {
		h.failure(ctx, gorm.ErrRecordNotFound)
		return
	}
	if err = h.FuelConsumptionRepository.DeleteFuelConsumption(ctx.Request.Context(), id); err != nil {
		h.failure(ctx, err)
		return
	}
	ctx.Redirect(303, "/fuel_consumptions")
}
