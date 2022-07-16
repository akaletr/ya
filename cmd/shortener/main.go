package main

import (
	"cmd/shortener/main.go/internal/app"
	"cmd/shortener/main.go/internal/config"
	"log"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	myApp, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(myApp.Start())
}
