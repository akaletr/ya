package storage

import "cmd/shortener/main.go/internal/model"

type Storage interface {
	Read(value string) (string, error)
	Write(id, key, value string) error
	WriteBatch(data model.DataBatch) error

	ReadAll(id string) (map[string]string, error)
}
