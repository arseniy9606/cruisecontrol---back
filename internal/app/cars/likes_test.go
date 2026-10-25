package cars

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestLikesTogglePersistenceAndStatuses(t *testing.T) {
	path := filepath.Join(t.TempDir(), "likes.json")
	r := NewCarRepository("media")
	if err := r.UseLikesFile(path); err != nil {
		t.Fatal(err)
	}
	baseline := len(r.PublishedCars()[0].CarLikedBy)
	if liked, err := r.ToggleLike(1, 1001); err != nil || !liked {
		t.Fatal(liked, err)
	}
	restored := NewCarRepository("media")
	if err := restored.UseLikesFile(path); err != nil {
		t.Fatal(err)
	}
	if n := len(restored.PublishedCars()[0].CarLikedBy); n != baseline+1 {
		t.Fatal(n)
	}
	if liked, err := restored.ToggleLike(1, 1001); err != nil || liked {
		t.Fatal(liked, err)
	}
	if n := len(restored.PublishedCars()[0].CarLikedBy); n != baseline {
		t.Fatal(n)
	}
	for _, id := range []int{5, 6, 999} {
		if _, err := r.ToggleLike(id, 1001); err != ErrCarNotFound {
			t.Fatal(id, err)
		}
	}
}
func TestConcurrentLikesAndSnapshotIsolation(t *testing.T) {
	r := NewCarRepository("media")
	n := len(r.PublishedCars()[0].CarLikedBy)
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
	copy := r.PublishedCars()
	if len(copy[0].CarLikedBy) != n+30 {
		t.Fatal("lost likes")
	}
	copy[0].CarLikedBy[0] = -1
	if r.PublishedCars()[0].CarLikedBy[0] == -1 {
		t.Fatal("mutable snapshot")
	}
}
func TestSaveFailureRollsBack(t *testing.T) {
	r := NewCarRepository("media")
	p := filepath.Join(t.TempDir(), "parent")
	if err := r.UseLikesFile(filepath.Join(p, "likes.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("block directory"), 0600); err != nil {
		t.Fatal(err)
	}
	n := len(r.PublishedCars()[0].CarLikedBy)
	if _, err := r.ToggleLike(1, 1001); err == nil {
		t.Fatal("expected disk failure")
	}
	if len(r.PublishedCars()[0].CarLikedBy) != n {
		t.Fatal("failed write changed count")
	}
}
