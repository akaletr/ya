package storage

import "errors"

type mockStorage struct {
	db map[string]string
}

func (s mockStorage) Read(value string) (string, error) {
	if url, ok := s.db[value]; ok {
		return url, nil
	}
	err := errors.New("error: there is no url in database")
	return "", err
}

func (s mockStorage) Write(key, value string) error {
	return nil
}

func NewMock() Storage {
	storage := storage{
		db: make(map[string]string),
	}

	storage.db["exist"] = "yes"
	storage.db["hello"] = "world"
	return &storage
}
