package fuel_consumptions

import (
	"github.com/joho/godotenv"
	"net/url"
	"os"
)

func FuelConsumptionEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
func FuelConsumptionDSN() string {
	_ = godotenv.Load()
	if dsn := os.Getenv("FUEL_CONSUMPTIONS_DATABASE_URL"); dsn != "" {
		return dsn
	}
	u := url.URL{Scheme: "postgres", Host: FuelConsumptionEnv("FUEL_CONSUMPTIONS_DB_HOST", "localhost") + ":" + FuelConsumptionEnv("FUEL_CONSUMPTIONS_DB_PORT", "55432"), Path: "/" + FuelConsumptionEnv("FUEL_CONSUMPTIONS_DB_NAME", "fuel_consumptions_lab2"), User: url.UserPassword(FuelConsumptionEnv("FUEL_CONSUMPTIONS_DB_USER", "fuel_consumptions_student"), FuelConsumptionEnv("FUEL_CONSUMPTIONS_DB_PASSWORD", "fuel_consumptions_lab2_password"))}
	u.RawQuery = "sslmode=disable"
	return u.String()
}
