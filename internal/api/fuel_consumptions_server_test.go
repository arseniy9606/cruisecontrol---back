package api

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFuelConsumptionHTTPContract(t *testing.T) {
	router := NewFuelConsumptionRouter("../..", "http://localhost:9000/fuel-consumptions")
	cases := []struct {
		path               string
		status             int
		contains, excludes string
	}{
		{"/fuel_consumptions", 200, "Седан", "Архивный"},
		{"/fuel_consumptions", 200, "Внедорожник", "Внедорожник · черновик"},
		{"/fuel_consumptions?fuel_consumptions_fuel_max=11", 200, "Внедорожник", "Фура"},
		{"/fuel_consumptions?fuel_consumptions_fuel_min=6.2&fuel_consumptions_fuel_max=6.7", 200, "Седан", "Внедорожник"},
		{"/fuel_consumptions?fuel_consumptions_fuel_max=0", 200, "Автомобили не найдены", "fuel_consumptions-tile-photo"},
		{"/fuel_consumptions?fuel_consumptions_fuel_min=NaN", 400, "Укажите расход", "fuel_consumptions-tile-photo"},
		{"/fuel_consumptions?fuel_consumptions_fuel_max=+Inf", 400, "Укажите расход", "fuel_consumptions-tile-photo"},
		{"/fuel_consumptions?fuel_consumptions_fuel_min=9&fuel_consumptions_fuel_max=2", 400, "Укажите расход", "fuel_consumptions-tile-photo"},
		{"/fuel_consumptions/feed/", 200, "Седан", "Автомобиль не найден"},
		{"/fuel_consumptions/feed/1?next=true", 200, "Внедорожник", "Седан"},
		{"/fuel_consumptions/feed/4?next=true", 200, "Седан", "Внедорожник"},
		{"/fuel_consumptions/feed/2", 200, "3 лайков", "6 лайков"},
		{"/fuel_consumptions/feed/5", 404, "Автомобиль не найден", "Внедорожник"},
		{"/fuel_consumptions/feed/6", 404, "Автомобиль не найден", "Архивный"},
		{"/fuel_consumptions/feed/999", 404, "Автомобиль не найден", "fuel_consumptions-feed-video"},
		{"/fuel_consumptions/feed/x", 404, "Автомобиль не найден", "fuel_consumptions-feed-video"},
		{"/fuel_consumptions/draft", 200, "Внедорожник · черновик", "Опубликовать"},
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
	for _, path := range []string{"/fuel_consumptions", "/fuel_consumptions/feed/1", "/fuel_consumptions/draft"} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest("POST", path, nil))
		if w.Code != 404 {
			t.Errorf("unexpected writable endpoint %s", path)
		}
	}
}

func TestFuelConsumptionDomainFields(t *testing.T) {
	router := NewFuelConsumptionRouter("../..", "http://localhost:9000/fuel-consumptions")
	for _, path := range []string{"/fuel_consumptions/feed/1", "/fuel_consumptions/draft"} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		body := w.Body.String()
		if strings.Contains(body, "Режим движения") || strings.Contains(body, "driving-mode") {
			t.Errorf("obsolete domain field on %s", path)
		}
		if !strings.Contains(body, "объём бака, л") && !strings.Contains(body, "Объём бака, л") {
			t.Errorf("tank volume missing on %s", path)
		}
		if path == "/fuel_consumptions/draft" && (!strings.Contains(body, `name="fuel_consumptions_tank_volume" type="number"`) || !strings.Contains(body, `name="fuel_consumptions_rate" type="number"`)) {
			t.Error("both editable domain fields must be numeric")
		}
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("GET", "/fuel_consumptions?fuel_consumptions_fuel_min=6.2&fuel_consumptions_fuel_max=6.7", nil))
	body := w.Body.String()
	if strings.Count(body, `class="fuel_consumptions-tile"`) != 2 {
		t.Error("inclusive fuel consumption range must return exactly two published cards")
	}
	tiles := strings.SplitN(body, `class="fuel_consumptions-grid"`, 2)[1]
	if strings.Contains(tiles, "объём бака") || strings.Contains(tiles, "FuelTankVolume") {
		t.Error("tiles must show only consumption")
	}
}

func TestLikeEndpointAndStrictMode(t *testing.T) {
	r := NewFuelConsumptionRouter("../..", "http://localhost:9000/fuel-consumptions")
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/fuel_consumptions/1/like", nil)
	r.ServeHTTP(w, req)
	if w.Code != 303 || w.Header().Get("Location") != "/fuel_consumptions/feed/1" {
		t.Fatal(w.Code, w.Header())
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 || !cookies[0].HttpOnly {
		t.Fatal("visitor cookie missing")
	}
	view := httptest.NewRecorder()
	get := httptest.NewRequest("GET", "/fuel_consumptions/feed/1", nil)
	get.AddCookie(cookies[0])
	r.ServeHTTP(view, get)
	if !strings.Contains(view.Body.String(), `aria-pressed="true"`) || !strings.Contains(view.Body.String(), "7 лайков") {
		t.Fatal("like not shown")
	}
	undo := httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/fuel_consumptions/1/like", nil)
	req.AddCookie(cookies[0])
	r.ServeHTTP(undo, req)
	view = httptest.NewRecorder()
	r.ServeHTTP(view, get)
	if !strings.Contains(view.Body.String(), "6 лайков") {
		t.Fatal("undo failed")
	}
	denied := httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/fuel_consumptions/1/like", nil)
	req.Header.Set("Origin", "https://untrusted.example")
	r.ServeHTTP(denied, req)
	if denied.Code != 403 {
		t.Fatal("foreign origin accepted")
	}
	strict := newFuelConsumptionRouter("../..", "media", "", false)
	w = httptest.NewRecorder()
	strict.ServeHTTP(w, httptest.NewRequest("POST", "/fuel_consumptions/1/like", nil))
	if w.Code != 404 {
		t.Fatal("strict mode allows changes")
	}
	w = httptest.NewRecorder()
	strict.ServeHTTP(w, httptest.NewRequest("GET", "/fuel_consumptions/feed/1", nil))
	if strings.Contains(w.Body.String(), `method="post"`) {
		t.Fatal("strict mode renders like form")
	}
}
