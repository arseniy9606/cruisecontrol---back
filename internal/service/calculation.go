package service

import (
	"errors"
	"github.com/arseniy9606/cruisecontrol---back/internal/model"
	"math"
)

var ErrInvalid = errors.New("invalid operation")

func Round(v float64) float64 { return math.Round(v*100) / 100 }
func Calculate(r *model.FuelConsumption) error {
	if r.Route == "" || r.SavingPercent == nil || *r.SavingPercent < 0 || *r.SavingPercent >= 100 || len(r.Items) == 0 {
		return ErrInvalid
	}
	total := 0.0
	for i := range r.Items {
		x := &r.Items[i]
		if x.DistanceKM == nil || *x.DistanceKM <= 0 || x.Car.ConsumptionRate <= 0 || x.Car.FuelTankVolume <= 0 || x.Car.Status != "published" {
			return ErrInvalid
		}
		base := *x.DistanceKM * x.Car.ConsumptionRate / 100
		saved := Round(base * *r.SavingPercent / 100)
		cruise := Round(base * (1 - *r.SavingPercent/100))
		base = Round(base)
		reach := Round(x.Car.FuelTankVolume / (x.Car.ConsumptionRate * (1 - *r.SavingPercent/100)) * 100)
		x.BaseLiters = &base
		x.CruiseLiters = &cruise
		x.SavedLiters = &saved
		x.RangeKM = &reach
		total += saved
	}
	total = Round(total)
	r.TotalSavedLiters = &total
	return nil
}
