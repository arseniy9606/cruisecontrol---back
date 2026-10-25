package car_handler

import (
	"errors"
	fc "github.com/arseniy9606/cruisecontrol---back/internal/app/cars"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

type CarHandler struct {
	CarRepository *fc.CarRepository
	Media         *fc.CarMedia
}
type CarCard struct {
	fc.Car
	CarLikeCount             int64
	CarImageURL, CarVideoURL string
}

func (h *CarHandler) card(ctx *gin.Context, car fc.Car) (CarCard, error) {
	count, err := h.CarRepository.CarLikeCount(ctx.Request.Context(), car.CarID)
	return CarCard{car, count, h.Media.Resolve(ctx.Request.Context(), car.CarImageURL, false), h.Media.Resolve(ctx.Request.Context(), car.CarVideoURL, true)}, err
}
func bound(raw string, min, max float64) (float64, bool) {
	n, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(raw), ",", "."), 64)
	return n, err == nil && !math.IsNaN(n) && !math.IsInf(n, 0) && n >= min && n <= max
}
func (h *CarHandler) failure(ctx *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.HTML(404, "cars_feed.html", gin.H{"CarError": "Автомобиль не найден"})
		return
	}
	log.Printf("fuel consumptions: %v", err)
	ctx.String(500, "Не удалось выполнить операцию с базой данных. Проверьте журналы приложения.")
}

// 1. GET tiles: ORM SELECT, numeric range and status are filtered in SQL.
func (h *CarHandler) GetCarTiles(ctx *gin.Context) {
	minRaw, maxRaw := ctx.Query("cars_fuel_min"), ctx.Query("cars_fuel_max")
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
	cards := make([]CarCard, 0)
	message := ""
	status := 200
	if !valid || min > max {
		message = "Укажите расход от 0 до 60 л/100 км: нижняя граница не должна превышать верхнюю."
		status = 400
	} else {
		cars, err := h.CarRepository.PublishedCars(ctx.Request.Context(), min, max)
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
	ctx.HTML(status, "cars_tiles.html", gin.H{"CarCards": cards, "CarFuelMin": minRaw, "CarFuelMax": maxRaw, "CarFilterError": message, "CarFilterActive": ctx.Request.URL.RawQuery != ""})
}

// 2. GET feed: PostgreSQL returns a single card, never a filtered array.
func (h *CarHandler) GetCarFeed(ctx *gin.Context) {
	raw := strings.TrimPrefix(ctx.Param("carsID"), "/")
	var id int64
	if raw != "" {
		var err error
		id, err = strconv.ParseInt(raw, 10, 64)
		if err != nil || id < 1 {
			h.failure(ctx, gorm.ErrRecordNotFound)
			return
		}
	}
	car, err := h.CarRepository.FeedCar(ctx.Request.Context(), id, ctx.Query("next") == "true")
	if err != nil {
		h.failure(ctx, err)
		return
	}
	card, err := h.card(ctx, car)
	if err != nil {
		h.failure(ctx, err)
		return
	}
	ctx.HTML(200, "cars_feed.html", gin.H{"CarCard": card})
}
func (h *CarHandler) draft(ctx *gin.Context, status int, message, name, description, tank, rate string) {
	car, err := h.CarRepository.DraftCar(ctx.Request.Context())
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
		name = car.CarName
	}
	if message == "" {
		description = car.CarDescription
		if car.FuelTankVolume > 0 {
			tank = strconv.FormatFloat(car.FuelTankVolume, 'f', 1, 64)
		}
		if car.CarRate > 0 {
			rate = strconv.FormatFloat(car.CarRate, 'f', 1, 64)
		}
	}
	ctx.HTML(status, "cars_draft.html", gin.H{"CarCard": card, "CarHasDraft": exists, "CarError": message, "CarName": name, "CarDescription": description, "CarTank": tank, "CarRate": rate})
}

// 3. GET draft never creates database rows.
func (h *CarHandler) GetCarDraft(ctx *gin.Context) {
	h.draft(ctx, 200, "", "", "", "", "")
}

// 4. POST create: only a name, no file bytes or filenames are received.
func (h *CarHandler) PostCarCreate(ctx *gin.Context) {
	name := strings.TrimSpace(ctx.PostForm("cars_name"))
	if name == "" || utf8.RuneCountInString(name) > 120 {
		h.draft(ctx, 400, "Введите название автомобиля длиной от 1 до 120 символов.", name, "", "", "")
		return
	}
	err := h.CarRepository.CreateCar(ctx.Request.Context(), name, h.Media.PublicBase)
	if errors.Is(err, fc.ErrCarDraftExists) {
		h.draft(ctx, 409, "Черновик уже существует. Завершите его публикацию.", "", "", "", "")
		return
	}
	if err != nil {
		h.failure(ctx, err)
		return
	}
	ctx.Redirect(303, "/cars/draft")
}

// 5. POST publish: status and timestamps are generated on the server.
func (h *CarHandler) PostCarPublish(ctx *gin.Context) {
	description := strings.TrimSpace(ctx.PostForm("cars_description"))
	tankRaw, rateRaw := ctx.PostForm("cars_tank_volume"), ctx.PostForm("cars_rate")
	tank, okTank := bound(tankRaw, 1, 1500)
	rate, okRate := bound(rateRaw, 0.1, 60)
	if description == "" || utf8.RuneCountInString(description) > 1000 || !okTank || !okRate {
		h.draft(ctx, 400, "Заполните описание (1–1000 символов), объём бака (1–1500 л) и расход (0,1–60 л/100 км).", "", description, tankRaw, rateRaw)
		return
	}
	err := h.CarRepository.PublishCar(ctx.Request.Context(), description, tank, rate)
	if err != nil {
		h.failure(ctx, err)
		return
	}
	ctx.Redirect(303, "/cars")
}

// 6. POST delete: repository executes raw UPDATE through database/sql.
func (h *CarHandler) PostCarDelete(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("carsDeleteID"), 10, 64)
	if err != nil || id < 1 {
		h.failure(ctx, gorm.ErrRecordNotFound)
		return
	}
	if err = h.CarRepository.DeleteCar(ctx.Request.Context(), id); err != nil {
		h.failure(ctx, err)
		return
	}
	ctx.Redirect(303, "/cars")
}
