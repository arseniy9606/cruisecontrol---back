package api

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCarHTTPContract(t *testing.T) {
	router := NewCarRouter("../..", "http://localhost:9000/cars")
	cases := []struct {
		path               string
		status             int
		contains, excludes string
	}{
		{"/cars", 200, "Седан", "Архивный"},
		{"/cars", 200, "Внедорожник", "Внедорожник · черновик"},
		{"/cars?cars_fuel_max=11", 200, "Внедорожник", "Фура"},
		{"/cars?cars_fuel_min=6.2&cars_fuel_max=6.7", 200, "Седан", "Внедорожник"},
		{"/cars?cars_fuel_max=0", 200, "Автомобили не найдены", "cars-tile-photo"},
		{"/cars?cars_fuel_min=NaN", 400, "Укажите расход", "cars-tile-photo"},
		{"/cars?cars_fuel_max=+Inf", 400, "Укажите расход", "cars-tile-photo"},
		{"/cars?cars_fuel_min=9&cars_fuel_max=2", 400, "Укажите расход", "cars-tile-photo"},
		{"/cars/feed/", 200, "Седан", "Автомобиль не найден"},
		{"/cars/feed/1?next=true", 200, "Внедорожник", "Седан"},
		{"/cars/feed/4?next=true", 200, "Седан", "Внедорожник"},
		{"/cars/feed/2", 200, "3 лайков", "6 лайков"},
		{"/cars/feed/5", 404, "Автомобиль не найден", "Внедорожник"},
		{"/cars/feed/6", 404, "Автомобиль не найден", "Архивный"},
		{"/cars/feed/999", 404, "Автомобиль не найден", "cars-feed-video"},
		{"/cars/feed/x", 404, "Автомобиль не найден", "cars-feed-video"},
		{"/cars/draft", 200, "Внедорожник · черновик", "Опубликовать"},
	}
	for _, tc := range cases {
		t.Run(tc.path+tc.contains, func(t *testing.T) {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest("GET", tc.path, nil))
			body := w.Body.String()
			if w.Code != tc.status {
				t.Fatalf("status %d, want %d", w.Code, tc.status)
			}
			if !strings.Contains(body, tc.contains) {
				t.Errorf("missing %q", tc.contains)
			}
			if strings.Contains(body, tc.excludes) {
				t.Errorf("unexpected %q", tc.excludes)
			}
			if strings.Contains(body, "<script") {
				t.Error("client-side JavaScript is forbidden")
			}
		})
	}
	for _, path := range []string{"/cars", "/cars/feed/1", "/cars/draft"} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest("POST", path, nil))
		if w.Code != 404 {
			t.Errorf("unexpected writable endpoint %s", path)
		}
	}
}

func TestCarDomainFields(t *testing.T) {
	router := NewCarRouter("../..", "http://localhost:9000/cars")
	for _, path := range []string{"/cars/feed/1", "/cars/draft"} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		body := w.Body.String()
		if strings.Contains(body, "Режим движения") || strings.Contains(body, "driving-mode") {
			t.Errorf("obsolete domain field on %s", path)
		}
		if path == "/cars/feed/1" && !strings.Contains(body, "объём бака, л") {
			t.Errorf("tank volume missing on %s", path)
		}
		if path == "/cars/draft" {
			if strings.Contains(body, "Далее") || strings.Contains(body, "cars_tank_volume") || strings.Contains(body, "<textarea") {
				t.Error("Lab 1 draft must contain only first-stage inputs, without Next")
			}
			if strings.Count(body, `type="file"`) != 2 || !strings.Contains(body, `id="cars-name"`) {
				t.Error("photo, video and car name inputs are required")
			}
		}
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("GET", "/cars?cars_fuel_min=6.2&cars_fuel_max=6.7", nil))
	body := w.Body.String()
	if strings.Count(body, `class="cars-tile"`) != 2 {
		t.Error("inclusive fuel consumption range must return exactly two published cards")
	}
	tiles := strings.SplitN(body, `class="cars-grid"`, 2)[1]
	if strings.Contains(tiles, "объём бака") || strings.Contains(tiles, "FuelTankVolume") {
		t.Error("tiles must show only consumption")
	}
}

func TestLikeEndpointAndStrictMode(t *testing.T) {
	r := NewCarRouter("../..", "http://localhost:9000/cars")
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/cars/1/like", nil)
	r.ServeHTTP(w, req)
	if w.Code != 303 || w.Header().Get("Location") != "/cars/feed/1" {
		t.Fatal(w.Code, w.Header())
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 || !cookies[0].HttpOnly {
		t.Fatal("visitor cookie missing")
	}
	view := httptest.NewRecorder()
	get := httptest.NewRequest("GET", "/cars/feed/1", nil)
	get.AddCookie(cookies[0])
	r.ServeHTTP(view, get)
	if !strings.Contains(view.Body.String(), `aria-pressed="true"`) || !strings.Contains(view.Body.String(), "7 лайков") {
		t.Fatal("like not shown")
	}
	undo := httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/cars/1/like", nil)
	req.AddCookie(cookies[0])
	r.ServeHTTP(undo, req)
	view = httptest.NewRecorder()
	r.ServeHTTP(view, get)
	if !strings.Contains(view.Body.String(), "6 лайков") {
		t.Fatal("undo failed")
	}
	denied := httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/cars/1/like", nil)
	req.Header.Set("Origin", "https://untrusted.example")
	r.ServeHTTP(denied, req)
	if denied.Code != 403 {
		t.Fatal("foreign origin accepted")
	}
	strict := newCarRouter("../..", "media", "", false)
	w = httptest.NewRecorder()
	strict.ServeHTTP(w, httptest.NewRequest("POST", "/cars/1/like", nil))
	if w.Code != 404 {
		t.Fatal("strict mode allows changes")
	}
	w = httptest.NewRecorder()
	strict.ServeHTTP(w, httptest.NewRequest("GET", "/cars/feed/1", nil))
	if strings.Contains(w.Body.String(), `method="post"`) {
		t.Fatal("strict mode renders like form")
	}
}
