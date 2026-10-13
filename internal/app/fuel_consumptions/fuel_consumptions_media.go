package fuel_consumptions

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Probe only the configured MinIO origin/bucket. Database URLs cannot trigger arbitrary requests.
type FuelConsumptionMedia struct {
	PublicBase, ProbeBase string
	Client                *http.Client
}

func NewFuelConsumptionMedia(publicBase, probeBase string) *FuelConsumptionMedia {
	return &FuelConsumptionMedia{strings.TrimRight(publicBase, "/"), strings.TrimRight(probeBase, "/"), &http.Client{Timeout: 800 * time.Millisecond, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}
func (m *FuelConsumptionMedia) Resolve(ctx context.Context, stored string, video bool) string {
	fallback := "/fuel_consumptions-assets/defaults/fuel_consumptions_default.jpg"
	if video {
		fallback = "/fuel_consumptions-assets/defaults/fuel_consumptions_default.mp4"
	}
	u, err := url.Parse(stored)
	base, e := url.Parse(m.PublicBase)
	if err != nil || e != nil || u.User != nil || u.Host != base.Host || u.Scheme != base.Scheme || u.RawQuery != "" || u.Fragment != "" || !strings.HasPrefix(u.Path, base.Path+"/") {
		return fallback
	}
	key := strings.TrimPrefix(u.Path, base.Path+"/")
	if key == "" || strings.Contains(key, "/") || key == "." || key == ".." {
		return fallback
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, m.ProbeBase+"/"+url.PathEscape(key), nil)
	if err != nil {
		return fallback
	}
	res, err := m.Client.Do(req)
	if err != nil {
		return fallback
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fallback
	}
	ct := res.Header.Get("Content-Type")
	if video && !strings.HasPrefix(ct, "video/") {
		return fallback
	}
	if !video && !strings.HasPrefix(ct, "image/") {
		return fallback
	}
	return stored
}
