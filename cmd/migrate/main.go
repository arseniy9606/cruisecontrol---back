package main

import (
	"github.com/arseniy9606/cruisecontrol---back/internal/storage"
	"log"
)

func main() {
	db, e := storage.Open()
	if e != nil {
		log.Fatal(e)
	}
	if e = storage.Migrate(db); e != nil {
		log.Fatal(e)
	}
	if e = storage.Seed(db); e != nil {
		log.Fatal(e)
	}
}
