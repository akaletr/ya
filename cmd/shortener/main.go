package main

import (
	"cmd/shortener/main.go/internal/app"
	"cmd/shortener/main.go/internal/config"
	"log"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Println(err)
	}

	myApp := app.New(cfg)
	log.Fatal(myApp.Start())
}
