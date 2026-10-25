package api

import (
	"context"
	"errors"
	"fmt"
	fc "github.com/arseniy9606/cruisecontrol---back/internal/app/cars"
	"github.com/arseniy9606/cruisecontrol---back/migrations"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// Uses a unique temporary schema, never clears the regular demonstration database.
func TestCarLab2PostgreSQL(t *testing.T) {
	dsn := os.Getenv("CARS_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set CARS_TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	admin, err := fc.OpenCarRepository(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	schema := fmt.Sprintf("cars_test_%d", time.Now().UnixNano())
	if err = admin.ORM().Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatal(err)
	}
	defer admin.ORM().Exec("DROP SCHEMA " + schema + " CASCADE") // Only this isolated, test-created schema.
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	r, err := fc.OpenCarRepository(u.String())
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	db, err := r.ORM().DB()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err = migrations.ApplyCarSchema(ctx, db); err != nil {
		t.Fatal(err)
	}
	if err = migrations.ApplyCarSchema(ctx, db); err != nil {
		t.Fatal("migration is not idempotent", err)
	}
	if err = migrations.SeedEmptyCarDatabase(ctx, db); err != nil {
		t.Fatal(err)
	}
	mediaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if strings.Contains(req.URL.Path, "missing") || strings.Contains(req.URL.Path, "not_found") {
			w.WriteHeader(404)
			return
		}
		if strings.HasSuffix(req.URL.Path, ".mp4") {
			w.Header().Set("Content-Type", "video/mp4")
		} else {
			w.Header().Set("Content-Type", "image/jpeg")
		}
		w.WriteHeader(200)
	}))
	defer mediaServer.Close()
	media := fc.NewCarMedia("http://localhost:9100/cars", mediaServer.URL+"/cars")
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	router := NewCarRouter(root, r, media)
	request := func(method, path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if method == "POST" {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}
	check := func(w *httptest.ResponseRecorder, status int) {
		t.Helper()
		if w.Code != status {
			t.Fatalf("expected %d got %d: %s", status, w.Code, w.Body.String())
		}
	}
	t.Run("three_GET_pages_and_readonly_likes", func(t *testing.T) {
		for _, path := range []string{"/cars", "/cars/feed/1", "/cars/draft"} {
			w := request("GET", path, "")
			check(w, 200)
			if strings.Contains(w.Body.String(), "<script") {
				t.Fatal("JS found")
			}
		}
		check(request("POST", "/cars/1/like", ""), 404)
		var pages int
		for _, route := range router.Routes() {
			if !strings.HasPrefix(route.Path, "/cars-assets") {
				pages++
			}
		}
		if pages != 6 {
			t.Fatalf("page controllers=%d", pages)
		}
	})
	t.Run("SQL_filter_and_single_feed_card", func(t *testing.T) {
		w := request("GET", "/cars?cars_fuel_min=10&cars_fuel_max=11", "")
		check(w, 200)
		if !strings.Contains(w.Body.String(), "Внедорожник") || strings.Contains(w.Body.String(), ">Седан<") {
			t.Fatal("filter failed")
		}
		check(request("GET", "/cars?cars_fuel_min=NaN", ""), 400)
		check(request("GET", "/cars/feed/6", ""), 404)
		check(request("GET", "/cars/feed/6?next=true", ""), 404)
		check(request("GET", "/cars/feed/5", ""), 404)
		check(request("GET", "/cars/feed/invalid", ""), 404)
		w = request("GET", "/cars/feed/1?next=true", "")
		check(w, 200)
		if !strings.Contains(w.Body.String(), "Внедорожник") {
			t.Fatal("next failed")
		}
		w = request("GET", "/cars/feed/7?next=true", "")
		check(w, 200)
		if !strings.Contains(w.Body.String(), ">Седан<") {
			t.Fatal("wrap failed")
		}
	})
	t.Run("live_database_changes", func(t *testing.T) {
		if err = r.ORM().Model(&fc.Car{}).Where("car_id = 1").Update("consumption_rate", 8.4).Error; err != nil {
			t.Fatal(err)
		}
		w := request("GET", "/cars/feed/1", "")
		check(w, 200)
		if !strings.Contains(w.Body.String(), "8,4") {
			t.Fatal("database rate not displayed")
		}
		like := fc.CarLike{CarID: 1, DriverID: 1}
		if err = r.ORM().Create(&like).Error; err != nil {
			t.Fatal(err)
		}
		count, err := r.CarLikeCount(ctx, 1)
		if err != nil || count != 7 {
			t.Fatal("like count", count, err)
		}
		if err = r.ORM().Create(&fc.CarLike{CarID: 1, DriverID: 1}).Error; err == nil {
			t.Fatal("duplicate like allowed")
		}
		if err = r.ORM().Delete(&fc.Driver{}, 1).Error; err == nil {
			t.Fatal("cascade/physical deletion unexpectedly allowed")
		}
	})
	t.Run("fallbacks_are_in_HTML", func(t *testing.T) {
		w := request("GET", "/cars/feed/7", "")
		check(w, 200)
		for _, file := range []string{"cars_default.jpg", "cars_default.mp4"} {
			if !strings.Contains(w.Body.String(), file) {
				t.Fatal("fallback missing", file)
			}
		}
		fallback := media.Resolve(ctx, "http://169.254.169.254/metadata", false)
		if !strings.Contains(fallback, "default.jpg") {
			t.Fatal("untrusted URL queried")
		}
	})
	t.Run("create_publish_delete_and_constraints", func(t *testing.T) {
		check(request("POST", "/cars/create", "cars_name=duplicate"), 409)
		check(request("POST", "/cars/publish", "cars_description=&cars_tank_volume=80&cars_rate=10"), 400)
		check(request("POST", "/cars/publish", url.Values{"cars_description": {"Опубликованная карточка"}, "cars_tank_volume": {"80"}, "cars_rate": {"10.7"}}.Encode()), 303)
		var published fc.Car
		if err = r.ORM().First(&published, 5).Error; err != nil || published.CarStatus != fc.CarPublished || published.CarFormedAt == nil {
			t.Fatal("publication", err)
		}
		var before, after int64
		r.ORM().Model(&fc.Car{}).Count(&before)
		check(request("GET", "/cars/draft", ""), 200)
		r.ORM().Model(&fc.Car{}).Count(&after)
		if before != after {
			t.Fatal("GET created a row")
		}
		check(request("POST", "/cars/create", url.Values{"cars_name": {"Новый автомобиль"}}.Encode()), 303)
		draft, err := r.DraftCar(ctx)
		if err != nil {
			t.Fatal(err)
		}
		var optionalFieldsEmpty bool
		if err := r.ORM().Raw("SELECT short_description IS NULL AND fuel_tank_volume IS NULL AND consumption_rate IS NULL AND formed_at IS NULL FROM cars WHERE car_id = ?", draft.CarID).Scan(&optionalFieldsEmpty).Error; err != nil || !optionalFieldsEmpty {
			t.Fatalf("new draft must have four NULL fields: empty=%v err=%v", optionalFieldsEmpty, err)
		}
		if !strings.Contains(draft.CarImageURL, "missing") || draft.CarFormedAt != nil {
			t.Fatal("draft system fields")
		}
		check(request("POST", "/cars/publish", url.Values{"cars_description": {"Новый автомобиль из PostgreSQL"}, "cars_tank_volume": {"70"}, "cars_rate": {"6.7"}, "creator_id": {"11"}, "car_status": {"удален"}}.Encode()), 303)
		check(request("GET", fmt.Sprintf("/cars/feed/%d", draft.CarID), ""), 200)
		check(request("POST", fmt.Sprintf("/cars/%d/delete", draft.CarID), ""), 303)
		check(request("GET", fmt.Sprintf("/cars/feed/%d", draft.CarID), ""), 404)
		var deleted fc.Car
		if err = r.ORM().First(&deleted, draft.CarID).Error; err != nil || deleted.CarStatus != fc.CarDeleted {
			t.Fatal("not soft deleted", err)
		}
		check(request("POST", "/cars/999999/delete", ""), 404)
	})
	t.Run("concurrent_draft_creation_is_unique", func(t *testing.T) {
		var wg sync.WaitGroup
		results := make(chan error, 8)
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() { defer wg.Done(); results <- r.CreateCar(ctx, "Concurrent car", media.PublicBase) }()
		}
		wg.Wait()
		close(results)
		success := 0
		for err := range results {
			if err == nil {
				success++
			} else if !errors.Is(err, fc.ErrCarDraftExists) {
				t.Fatal(err)
			}
		}
		if success != 1 {
			t.Fatalf("successful drafts=%d", success)
		}
		_, err = r.DraftCar(ctx)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatal(err)
		}
	})
	t.Run("reject_files_and_cross_origin", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/cars/create", strings.NewReader("file"))
		req.Header.Set("Content-Type", "multipart/form-data; boundary=lab2")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		check(w, 415)
		req = httptest.NewRequest("POST", "/cars/1/delete", strings.NewReader(""))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Origin", "https://other.example")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		check(w, 403)
	})
}
