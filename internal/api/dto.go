package api

import (
	"github.com/arseniy9606/cruisecontrol---back/internal/model"
	"time"
)

// No omitempty: nullable fields remain present before formation.
type CarDTO struct {
	ID              uint    `json:"id"`
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	ConsumptionRate float64 `json:"consumption_rate"`
	FuelTankVolume  float64 `json:"fuel_tank_volume"`
	ImageURL        string  `json:"image_url"`
	VideoURL        string  `json:"video_url"`
}
type ItemDTO struct {
	Car          CarDTO   `json:"car"`
	DistanceKM   *float64 `json:"distance_km"`
	BaseLiters   *float64 `json:"base_liters"`
	CruiseLiters *float64 `json:"cruise_liters"`
	SavedLiters  *float64 `json:"saved_liters"`
	RangeKM      *float64 `json:"range_km"`
}
type RequestDTO struct {
	ID               uint       `json:"id"`
	Status           string     `json:"status"`
	Creator          string     `json:"creator"`
	Moderator        *string    `json:"moderator"`
	CreatedAt        time.Time  `json:"created_at"`
	FormedAt         *time.Time `json:"formed_at"`
	CompletedAt      *time.Time `json:"completed_at"`
	Route            string     `json:"route"`
	SavingPercent    *float64   `json:"saving_percent"`
	TotalSavedLiters *float64   `json:"total_saved_liters"`
	CalculatedCount  int        `json:"calculated_count"`
	Items            []ItemDTO  `json:"items"`
}

func (s *Server) carDTO(c model.Car) CarDTO {
	return CarDTO{c.ID, c.Name, c.Description, c.ConsumptionRate, c.FuelTankVolume, s.Media.URL(c.ImageName), s.Media.URL(c.VideoName)}
}
func (s *Server) requestDTO(r model.FuelConsumption) RequestDTO {
	d := RequestDTO{ID: r.ID, Status: r.Status, Creator: r.Creator.Login, CreatedAt: r.CreatedAt, FormedAt: r.FormedAt, CompletedAt: r.CompletedAt, Route: r.Route, SavingPercent: r.SavingPercent, TotalSavedLiters: r.TotalSavedLiters, Items: []ItemDTO{}}
	if r.Moderator != nil {
		d.Moderator = &r.Moderator.Login
	}
	for _, i := range r.Items {
		d.Items = append(d.Items, ItemDTO{s.carDTO(i.Car), i.DistanceKM, i.BaseLiters, i.CruiseLiters, i.SavedLiters, i.RangeKM})
		if i.SavedLiters != nil {
			d.CalculatedCount++
		}
	}
	return d
}
