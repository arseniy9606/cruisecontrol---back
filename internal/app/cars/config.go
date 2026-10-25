package cars

import (
	"github.com/joho/godotenv"
	"net/url"
	"os"
)

func CarEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
func CarDSN() string {
	_ = godotenv.Load()
	if dsn := os.Getenv("CARS_DATABASE_URL"); dsn != "" {
		return dsn
	}
	u := url.URL{Scheme: "postgres", Host: CarEnv("CARS_DB_HOST", "localhost") + ":" + CarEnv("CARS_DB_PORT", "55432"), Path: "/" + CarEnv("CARS_DB_NAME", "cars_lab2"), User: url.UserPassword(CarEnv("CARS_DB_USER", "cars_student"), CarEnv("CARS_DB_PASSWORD", "cars_lab2_password"))}
	u.RawQuery = "sslmode=disable"
	return u.String()
}
