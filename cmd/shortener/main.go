package main

import (
	"cmd/shortener/main.go/internal/app"
	"log"
)

func main() {
	myApp := app.New()
	log.Fatal(myApp.Start())
}
