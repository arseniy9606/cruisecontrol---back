package fuel_consumptions

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

var ErrFuelConsumptionNotFound = errors.New("published fuel consumption not found")

// UseLikesFile adds local persistence to the same collection, without a database.
func (r *FuelConsumptionRepository) UseLikesFile(path string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.likesFile = path
	if path == "" {
		return nil
	}
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var saved map[int][]int
	if err = json.Unmarshal(raw, &saved); err != nil {
		return err
	}
	for i := range r.fuel_consumptions {
		if ids, ok := saved[r.fuel_consumptions[i].FuelConsumptionID]; ok {
			unique := make(map[int]bool)
			clean := make([]int, 0, len(ids))
			for _, id := range ids {
				if id > 0 && !unique[id] {
					clean = append(clean, id)
					unique[id] = true
				}
			}
			r.fuel_consumptions[i].FuelConsumptionLikedBy = clean
		}
	}
	return nil
}

// ToggleLike is atomic: the same visitor can add or remove only their own like.
func (r *FuelConsumptionRepository) ToggleLike(id, visitor int) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if visitor <= 0 {
		return false, errors.New("invalid visitor")
	}
	for i := range r.fuel_consumptions {
		car := &r.fuel_consumptions[i]
		if car.FuelConsumptionID != id || car.FuelConsumptionStatus != FuelConsumptionPublished {
			continue
		}
		previous := append([]int(nil), car.FuelConsumptionLikedBy...)
		liked := true
		next := make([]int, 0, len(previous)+1)
		for _, v := range previous {
			if v == visitor {
				liked = false
			} else {
				next = append(next, v)
			}
		}
		if liked {
			next = append(next, visitor)
		}
		car.FuelConsumptionLikedBy = next
		if r.likesFile != "" {
			saved := make(map[int][]int)
			for _, c := range r.fuel_consumptions {
				saved[c.FuelConsumptionID] = c.FuelConsumptionLikedBy
			}
			raw, err := json.MarshalIndent(saved, "", "  ")
			if err == nil {
				err = os.MkdirAll(filepath.Dir(r.likesFile), 0700)
			}
			if err == nil {
				var tmp *os.File
				tmp, err = os.CreateTemp(filepath.Dir(r.likesFile), "fuel_consumptions-likes-*.tmp")
				if err == nil {
					name := tmp.Name()
					_, err = tmp.Write(raw)
					if err == nil {
						err = tmp.Sync()
					}
					closeErr := tmp.Close()
					if err == nil {
						err = closeErr
					}
					if err == nil {
						err = os.Rename(name, r.likesFile)
					}
					if err != nil {
						_ = os.Remove(name)
					}
				}
			}
			if err != nil {
				car.FuelConsumptionLikedBy = previous
				return false, err
			}
		}
		return liked, nil
	}
	return false, ErrFuelConsumptionNotFound
}
