package main

import (
	"context"
	"flag"
	fc "github.com/arseniy9606/cruisecontrol---back/internal/app/fuel_consumptions"
	"github.com/arseniy9606/cruisecontrol---back/migrations"
	"log"
	"time"
)

func main() {
	seed := flag.Bool("seed", false, "initialize an empty demonstration database")
	flag.Parse()
	r, err := fc.OpenFuelConsumptionRepository(fc.FuelConsumptionDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()
	db, err := r.ORM().DB()
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err = migrations.ApplyFuelConsumptionSchema(ctx, db); err != nil {
		log.Fatal(err)
	}
	if *seed {
		if err = migrations.SeedEmptyFuelConsumptionDatabase(ctx, db); err != nil {
			log.Fatal(err)
		}
	}
	log.Println("ЛР2: схема готова; существующие данные сохранены")
}
