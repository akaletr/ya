package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func GetConfig() (Config, error) {
	cfg := Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
	}

	// берем конфиг из окружения
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	// читаем флаги, если есть - перезаписываем конфиг
	var baseURL, serverAddress, fileStoragePath string

	flag.StringVar(&baseURL, "b", "", "base url")
	flag.StringVar(&serverAddress, "a", "", "host to listen on")
	flag.StringVar(&fileStoragePath, "f", "", "file path")
	flag.Parse()

	if baseURL != "" {
		cfg.BaseURL = baseURL
	}
	if serverAddress != "" {
		cfg.ServerAddress = serverAddress
	}
	if fileStoragePath != "" {
		cfg.FileStoragePath = fileStoragePath
	}

	return cfg, nil
}
