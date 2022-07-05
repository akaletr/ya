package storage

import "cmd/shortener/main.go/internal/model"

type Storage interface {
	Read(id, value string) (string, error)
	ReadAll(id, baseURL string) (model.AllShortenerRequest, error)
	Write(id, key, value string) error
}
