package main

import (
	"cmd/shortener/main.go/internal/app"
	"cmd/shortener/main.go/internal/config"
	"fmt"
	"log"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Println(err)
	}

	myApp := app.New(cfg)
	log.Fatal(myApp.Start())
}
