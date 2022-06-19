package main

import (
	"log"

	"cmd/shortener/main.go/internal/app"
)

func main() {
	myApp := app.New()
	log.Fatal(myApp.Start())
}
