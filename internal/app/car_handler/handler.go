package car_handler

import (
	"crypto/rand"
	"encoding/binary"
	"github.com/arseniy9606/cruisecontrol---back/internal/app/cars"
	"github.com/gin-gonic/gin"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type CarHandler struct {
	CarRepository *cars.CarRepository
	LikesEnabled  bool
}
type CarCard struct {
	cars.Car
	CarLikeCount    int
	CarImageURL     string
	CarVideoURL     string
	CarLiked        bool
	CarLikesEnabled bool
}

func NewCarHandler(r *cars.CarRepository) *CarHandler {
	return &CarHandler{CarRepository: r, LikesEnabled: true}
}
func (h *CarHandler) carsCard(car cars.Car, ctx *gin.Context) CarCard {
	// Количество лайков вычисляет контроллер, оно не хранится отдельным полем модели.
	visitor, _ := ctx.Cookie("cars_visitor")
	id, _ := strconv.Atoi(visitor)
	liked := false
	for _, v := range car.CarLikedBy {
		if v == id {
			liked = true
		}
	}
	return CarCard{car, len(car.CarLikedBy), h.CarRepository.CarMediaURL(car.CarImageKey), h.CarRepository.CarMediaURL(car.CarVideoKey), liked, h.LikesEnabled}
}
func carsFuelBound(raw string) (float64, bool) {
	n, err := strconv.ParseFloat(raw, 64)
	return n, err == nil && !math.IsNaN(n) && !math.IsInf(n, 0) && n >= 0 && n <= 60
}

func (h *CarHandler) GetCarTiles(ctx *gin.Context) {
	minRaw, maxRaw := ctx.Query("cars_fuel_min"), ctx.Query("cars_fuel_max")
	minFuel, maxFuel := 0.0, 60.0
	valid, message := true, ""
	if minRaw != "" {
		minFuel, valid = carsFuelBound(minRaw)
	}
	if maxRaw != "" {
		var ok bool
		maxFuel, ok = carsFuelBound(maxRaw)
		valid = valid && ok
	}
	if !valid || minFuel > maxFuel {
		message = "Укажите расход от 0 до 60 л/100 км: нижняя граница не должна превышать верхнюю."
	}
	cards := make([]CarCard, 0)
	if message == "" {
		for _, car := range h.CarRepository.PublishedCars() {
			if car.CarRate >= minFuel && car.CarRate <= maxFuel {
				cards = append(cards, h.carsCard(car, ctx))
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
	ctx.HTML(status, "cars_tiles.html", gin.H{"CarCards": cards, "CarFuelMin": minRaw, "CarFuelMax": maxRaw, "CarFilterError": message, "CarFilterActive": ctx.Request.URL.RawQuery != ""})
}
func (h *CarHandler) GetCarFeed(ctx *gin.Context) {
	cars := h.CarRepository.PublishedCars()
	idRaw := strings.TrimPrefix(ctx.Param("carsID"), "/")
	selected := 0
	if idRaw != "" {
		id, err := strconv.Atoi(idRaw)
		selected = -1
		if err == nil {
			for i, car := range cars {
				if car.CarID == id {
					selected = i
					break
				}
			}
		}
	}
	if selected < 0 || len(cars) == 0 {
		ctx.HTML(http.StatusNotFound, "cars_feed.html", gin.H{"CarError": "Автомобиль не найден"})
		return
	}
	if ctx.Query("next") == "true" {
		selected = (selected + 1) % len(cars)
	}
	ctx.HTML(http.StatusOK, "cars_feed.html", gin.H{"CarCard": h.carsCard(cars[selected], ctx)})
}
func (h *CarHandler) GetCarDraft(ctx *gin.Context) {
	car, ok := h.CarRepository.DraftCar()
	if !ok {
		ctx.HTML(http.StatusNotFound, "cars_draft.html", gin.H{"CarError": "Черновик не найден"})
		return
	}
	ctx.HTML(http.StatusOK, "cars_draft.html", gin.H{"CarCard": h.carsCard(car, ctx)})
}

// POST changes state; GET remains safe. SameSite and Origin checking prevent cross-site forms.
func (h *CarHandler) PostCarLike(ctx *gin.Context) {
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
	value, _ := ctx.Cookie("cars_visitor")
	visitor, _ := strconv.Atoi(value)
	if visitor < 1000 {
		var raw [8]byte
		if _, err = rand.Read(raw[:]); err != nil {
			ctx.String(http.StatusInternalServerError, "Не удалось создать идентификатор")
			return
		}
		visitor = int(binary.BigEndian.Uint64(raw[:])&((1<<62)-1)) + 1000
		ctx.SetSameSite(http.SameSiteStrictMode)
		ctx.SetCookie("cars_visitor", strconv.Itoa(visitor), 365*24*60*60, "/cars", "", ctx.Request.TLS != nil, true)
	}
	if _, err = h.CarRepository.ToggleLike(id, visitor); err != nil {
		if err == cars.ErrCarNotFound {
			ctx.String(http.StatusNotFound, "Автомобиль не найден")
		} else {
			ctx.String(http.StatusInternalServerError, "Не удалось сохранить лайк")
		}
		return
	}
	target := "/cars/feed/" + strconv.Itoa(id)
	if ctx.PostForm("view") == "tiles" {
		target = "/cars"
	}
	ctx.Redirect(http.StatusSeeOther, target)
}
