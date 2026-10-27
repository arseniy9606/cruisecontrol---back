package storage

import (
	"github.com/arseniy9606/cruisecontrol---back/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

func Open() (*gorm.DB, error) {
	return gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &gorm.Config{TranslateError: true})
}
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&model.User{}, &model.Car{}, &model.CarLike{})
}
func Seed(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, u := range []model.User{{ID: 1, Login: "creator", FullName: "Водитель", PasswordHash: "!disabled"}, {ID: 2, Login: "moderator", FullName: "Другой водитель", PasswordHash: "!disabled"}} {
			if err := tx.Where("id = ?", u.ID).FirstOrCreate(&u).Error; err != nil {
				return err
			}
		}
		names := []string{"Седан", "Внедорожник", "Микроавтобус", "Грузовик"}
		keys := []string{"sedan", "suv", "minibus", "truck"}
		rates := []float64{7.2, 10.5, 12.4, 28.0}
		tanks := []float64{50, 70, 80, 300}
		for i, n := range names {
			c := model.Car{ID: uint(i + 1), Name: n, Description: "Автомобиль: расход топлива и объём бака", Status: "published", ConsumptionRate: &rates[i], FuelTankVolume: &tanks[i], CreatorID: 1, ImageName: "cars_" + keys[i] + ".jpg", VideoName: "cars_" + keys[i] + ".mp4"}
			if err := tx.Where("id = ?", c.ID).FirstOrCreate(&c).Error; err != nil {
				return err
			}
		}
		other := model.Car{ID: 5, Name: "Автомобиль другого водителя", Status: "published", CreatorID: 2, ConsumptionRate: &rates[0], FuelTankVolume: &tanks[0], ImageName: "cars_sedan.jpg", VideoName: "cars_sedan.mp4"}
		if e := tx.Where("creator_id = ?", 2).FirstOrCreate(&other).Error; e != nil {
			return e
		}
		for _, table := range []string{"users", "cars"} {
			if err := tx.Exec("SELECT setval(pg_get_serial_sequence('" + table + "','id'), GREATEST((SELECT MAX(id) FROM " + table + "),1))").Error; err != nil {
				return err
			}
		}
		return nil
	})
}
