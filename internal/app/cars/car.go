package cars

import (
	"strings"
	"sync"
)

type CarStatus string

const (
	CarDraft     CarStatus = "черновик"
	CarPublished CarStatus = "опубликован"
	CarDeleted   CarStatus = "удален"
)

// Car — запись расхода топлива автомобиля и объёма его бака.
// Два предметных поля: FuelTankVolume и CarRate.
// Единственная вложенная коллекция — ID пользователей, поставивших лайк.
type Car struct {
	CarID          int
	CarName        string
	CarDescription string
	CarStatus      CarStatus
	CarImageKey    string
	CarVideoKey    string
	FuelTankVolume float64
	CarRate        float64
	CarLikedBy     []int
}

type CarRepository struct {
	mu            sync.RWMutex
	likesFile     string
	cars          []Car
	carsMediaBase string
}

func NewCarRepository(mediaBase string) *CarRepository {
	return &CarRepository{carsMediaBase: strings.TrimRight(mediaBase, "/"), cars: []Car{
		{1, "Седан", "Комфортный седан для городской и загородной дороги. Учебный пример расхода и объёма бака; фактические значения зависят от модели, комплектации и условий движения.", CarPublished, "cars_sedan.jpg", "cars_sedan.mp4", 70, 6.7, []int{11, 12, 13, 14, 15, 16}},
		{2, "Внедорожник", "Внедорожник для дороги и пересечённой местности. Учебный пример Toyota Land Cruiser 2.8 D-4D: 10,7 л/100 км в смешанном цикле WLTP и бак 80 л. Видео иллюстрирует тип кузова, а не конкретную комплектацию.", CarPublished, "cars_suv.jpg", "cars_suv.mp4", 80, 10.7, []int{11, 17, 18}},
		{3, "Микроавтобус", "Пассажирский микроавтобус для групповых поездок. Вместительный салон и увеличенный бак. Расход и объём бака — учебные значения, а не паспортные данные машины на видео.", CarPublished, "cars_minibus.jpg", "cars_minibus.mp4", 80, 6.6, []int{12, 13, 17, 19}},
		{4, "Фура", "Седельный тягач с полуприцепом для перевозки грузов. Учебные значения расхода и объёма бака. Расход существенно зависит от массы груза, маршрута и скорости.", CarPublished, "cars_truck.jpg", "cars_truck.mp4", 480, 28.8, []int{11, 20}},
		{5, "Внедорожник · черновик", "Черновик карточки автомобиля. Заполните название, описание, расход топлива и объём бака. В первой лабораторной введённые данные не сохраняются.", CarDraft, "cars_suv.jpg", "cars_suv.mp4", 80, 10.7, []int{}},
		{6, "Архивный седан", "Удалённая карточка не доступна в ленте и плитке.", CarDeleted, "cars_sedan.jpg", "cars_sedan.mp4", 70, 6.7, []int{11}},
	}}
}

func (r *CarRepository) PublishedCars() []Car {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Car, 0)
	for _, car := range r.cars {
		if car.CarStatus == CarPublished {
			car.CarLikedBy = append([]int(nil), car.CarLikedBy...)
			result = append(result, car)
		}
	}
	return result
}
func (r *CarRepository) DraftCar() (Car, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, car := range r.cars {
		if car.CarStatus == CarDraft {
			car.CarLikedBy = append([]int(nil), car.CarLikedBy...)
			return car, true
		}
	}
	return Car{}, false
}
func (r *CarRepository) CarMediaURL(key string) string {
	return r.carsMediaBase + "/" + key
}
