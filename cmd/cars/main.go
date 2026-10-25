package main

import (
	"github.com/arseniy9606/cruisecontrol---back/internal/api"
	"log"
)

func main() {
	if err := api.StartCarServer(); err != nil {
		log.Fatal(err)
	}
}
