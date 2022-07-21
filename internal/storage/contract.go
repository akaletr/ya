package storage

import "cmd/shortener/main.go/internal/model"

type Storage interface {
	Start() error
	Ping() error

	Read(value string) (model.Note, error)
	Write(note model.Note) error
	WriteBatch(data []model.DataBatchItem) error
	ReadAll(id string) (map[string]string, error)
	Delete(note model.Note) error
}
