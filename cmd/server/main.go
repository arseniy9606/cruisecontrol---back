package main

import (
	"github.com/arseniy9606/cruisecontrol---back/internal/api"
	"github.com/arseniy9606/cruisecontrol---back/internal/storage"
	"log"
	"net/http"
	"time"
)

func main() {
	db, e := storage.Open()
	if e != nil {
		log.Fatal(e)
	}
	media, e := storage.NewMedia()
	if e != nil {
		log.Fatal(e)
	}
	s := api.Server{DB: db, Media: media}
	server := http.Server{Addr: ":8080", Handler: s.Router(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 60 * time.Second, WriteTimeout: 60 * time.Second}
	log.Fatal(server.ListenAndServe())
}
