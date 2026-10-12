package fuel_consumptions

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestLikesTogglePersistenceAndStatuses(t *testing.T) {
	path := filepath.Join(t.TempDir(), "likes.json")
	r := NewFuelConsumptionRepository("media")
	if err := r.UseLikesFile(path); err != nil {
		t.Fatal(err)
	}
	baseline := len(r.PublishedFuelConsumptions()[0].FuelConsumptionLikedBy)
	if liked, err := r.ToggleLike(1, 1001); err != nil || !liked {
		t.Fatal(liked, err)
	}
	restored := NewFuelConsumptionRepository("media")
	if err := restored.UseLikesFile(path); err != nil {
		t.Fatal(err)
	}
	if n := len(restored.PublishedFuelConsumptions()[0].FuelConsumptionLikedBy); n != baseline+1 {
		t.Fatal(n)
	}
	if liked, err := restored.ToggleLike(1, 1001); err != nil || liked {
		t.Fatal(liked, err)
	}
	if n := len(restored.PublishedFuelConsumptions()[0].FuelConsumptionLikedBy); n != baseline {
		t.Fatal(n)
	}
	for _, id := range []int{5, 6, 999} {
		if _, err := r.ToggleLike(id, 1001); err != ErrFuelConsumptionNotFound {
			t.Fatal(id, err)
		}
	}
}
func TestConcurrentLikesAndSnapshotIsolation(t *testing.T) {
	r := NewFuelConsumptionRepository("media")
	n := len(r.PublishedFuelConsumptions()[0].FuelConsumptionLikedBy)
	var wg sync.WaitGroup
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if _, err := r.ToggleLike(1, id); err != nil {
				t.Error(err)
			}
		}(1000 + i)
	}
	wg.Wait()
	copy := r.PublishedFuelConsumptions()
	if len(copy[0].FuelConsumptionLikedBy) != n+30 {
		t.Fatal("lost likes")
	}
	copy[0].FuelConsumptionLikedBy[0] = -1
	if r.PublishedFuelConsumptions()[0].FuelConsumptionLikedBy[0] == -1 {
		t.Fatal("mutable snapshot")
	}
}
func TestSaveFailureRollsBack(t *testing.T) {
	r := NewFuelConsumptionRepository("media")
	p := filepath.Join(t.TempDir(), "parent")
	if err := r.UseLikesFile(filepath.Join(p, "likes.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("block directory"), 0600); err != nil {
		t.Fatal(err)
	}
	n := len(r.PublishedFuelConsumptions()[0].FuelConsumptionLikedBy)
	if _, err := r.ToggleLike(1, 1001); err == nil {
		t.Fatal("expected disk failure")
	}
	if len(r.PublishedFuelConsumptions()[0].FuelConsumptionLikedBy) != n {
		t.Fatal("failed write changed count")
	}
}
