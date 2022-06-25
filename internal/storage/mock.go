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
	storage := mockStorage{
		db: make(map[string]string),
	}

	storage.db["kUxCqw"] = "https://www.delftstack.com/ru/howto/go/how-to-read-a-file-line-by-line-in-go/"
	storage.db["D-rwfg"] = "https://www.jetbrains.com/ru-ru/"

	return &storage
}
