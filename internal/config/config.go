package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
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

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

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
