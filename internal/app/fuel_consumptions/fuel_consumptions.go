package fuel_consumptions

import (
	"strings"
	"sync"
)

type FuelConsumptionStatus string

const (
	FuelConsumptionDraft     FuelConsumptionStatus = "черновик"
	FuelConsumptionPublished FuelConsumptionStatus = "опубликован"
	FuelConsumptionDeleted   FuelConsumptionStatus = "удален"
)

// FuelConsumption — запись расхода топлива автомобиля и объёма его бака.
// Два предметных поля: FuelTankVolume и FuelConsumptionRate.
// Единственная вложенная коллекция — ID пользователей, поставивших лайк.
type FuelConsumption struct {
	FuelConsumptionID          int
	FuelConsumptionName        string
	FuelConsumptionDescription string
	FuelConsumptionStatus      FuelConsumptionStatus
	FuelConsumptionImageKey    string
	FuelConsumptionVideoKey    string
	FuelTankVolume             float64
	FuelConsumptionRate        float64
	FuelConsumptionLikedBy     []int
}

type FuelConsumptionRepository struct {
	mu                         sync.RWMutex
	likesFile                  string
	fuel_consumptions          []FuelConsumption
	fuel_consumptionsMediaBase string
}

func NewFuelConsumptionRepository(mediaBase string) *FuelConsumptionRepository {
	return &FuelConsumptionRepository{fuel_consumptionsMediaBase: strings.TrimRight(mediaBase, "/"), fuel_consumptions: []FuelConsumption{
		{1, "Седан", "Комфортный седан для городской и загородной дороги. Учебный пример расхода и объёма бака; фактические значения зависят от модели, комплектации и условий движения.", FuelConsumptionPublished, "fuel_consumptions_sedan.jpg", "fuel_consumptions_sedan.mp4", 70, 6.7, []int{11, 12, 13, 14, 15, 16}},
		{2, "Внедорожник", "Внедорожник для дороги и пересечённой местности. Учебный пример Toyota Land Cruiser 2.8 D-4D: 10,7 л/100 км в смешанном цикле WLTP и бак 80 л. Видео иллюстрирует тип кузова, а не конкретную комплектацию.", FuelConsumptionPublished, "fuel_consumptions_suv.jpg", "fuel_consumptions_suv.mp4", 80, 10.7, []int{11, 17, 18}},
		{3, "Микроавтобус", "Пассажирский микроавтобус для групповых поездок. Вместительный салон и увеличенный бак. Расход и объём бака — учебные значения, а не паспортные данные машины на видео.", FuelConsumptionPublished, "fuel_consumptions_minibus.jpg", "fuel_consumptions_minibus.mp4", 80, 6.6, []int{12, 13, 17, 19}},
		{4, "Фура", "Седельный тягач с полуприцепом для перевозки грузов. Учебные значения расхода и объёма бака. Расход существенно зависит от массы груза, маршрута и скорости.", FuelConsumptionPublished, "fuel_consumptions_truck.jpg", "fuel_consumptions_truck.mp4", 480, 28.8, []int{11, 20}},
		{5, "Внедорожник · черновик", "Черновик карточки автомобиля. Заполните название, описание, расход топлива и объём бака. В первой лабораторной введённые данные не сохраняются.", FuelConsumptionDraft, "fuel_consumptions_suv.jpg", "fuel_consumptions_suv.mp4", 80, 10.7, []int{}},
		{6, "Архивный седан", "Удалённая карточка не доступна в ленте и плитке.", FuelConsumptionDeleted, "fuel_consumptions_sedan.jpg", "fuel_consumptions_sedan.mp4", 70, 6.7, []int{11}},
	}}
}

func (r *FuelConsumptionRepository) PublishedFuelConsumptions() []FuelConsumption {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]FuelConsumption, 0)
	for _, car := range r.fuel_consumptions {
		if car.FuelConsumptionStatus == FuelConsumptionPublished {
			car.FuelConsumptionLikedBy = append([]int(nil), car.FuelConsumptionLikedBy...)
			result = append(result, car)
		}
	}
	return result
}
func (r *FuelConsumptionRepository) DraftFuelConsumption() (FuelConsumption, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, car := range r.fuel_consumptions {
		if car.FuelConsumptionStatus == FuelConsumptionDraft {
			car.FuelConsumptionLikedBy = append([]int(nil), car.FuelConsumptionLikedBy...)
			return car, true
		}
	}
	return FuelConsumption{}, false
}
func (r *FuelConsumptionRepository) FuelConsumptionMediaURL(key string) string {
	return r.fuel_consumptionsMediaBase + "/" + key
}
