package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	SecretKey       string `env:"SECRET_KEY"`
}

func GetConfig() (Config, error) {
	cfg := Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
		SecretKey:       "yandex",
	}

	// берем конфиг из окружения
	err := env.Parse(&cfg)
	if err != nil {
		return Config{}, err
	}

	// читаем флаги, если есть - перезаписываем конфиг
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "base url")
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "host to listen on")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "file path")
	flag.StringVar(&cfg.SecretKey, "k", cfg.SecretKey, "secret key")
	flag.Parse()

	return cfg, nil
}
