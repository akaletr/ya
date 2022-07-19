package storage

import (
	"cmd/shortener/main.go/internal/model"
	"errors"
)

type mockStorage struct {
	db map[string]model.Note
}

func (s mockStorage) Read(value string) (model.Note, error) {
	if note, ok := s.db[value]; ok {
		return note, nil
	}
	err := errors.New("error: there is no url in database")
	return model.Note{}, err
}

func (s mockStorage) Write(note model.Note) error {
	return nil
}

func (s mockStorage) WriteBatch(data []model.DataBatchItem) error {
	return nil
}

func (s mockStorage) ReadAll(id string) (map[string]string, error) {
	return nil, nil
}

func (s mockStorage) Delete(note model.Note) error {
	return nil
}

func (s mockStorage) Start() error {
	return nil
}

func (s mockStorage) Ping() error {
	return nil
}

func NewMock() Storage {
	storage := mockStorage{
		db: make(map[string]model.Note),
	}

	storage.db["kUxCqw"] = model.Note{
		Long: "https://www.delftstack.com/ru/howto/go/how-to-read-a-file-line-by-line-in-go/",
	}
	storage.db["D-rwfg"] = model.Note{
		Long: "https://www.jetbrains.com/ru-ru/",
	}
	return &storage
}
