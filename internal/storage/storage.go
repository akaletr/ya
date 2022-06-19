package storage

import (
	"errors"
)

type storage struct {
	db map[string]string
}

func (s storage) Read(value string) (string, error) {
	if url, ok := s.db[value]; ok {
		return url, nil
	}
	err := errors.New("error: there is no url in database")
	return "", err
}

func (s storage) Write(key, value string) error {
	s.db[key] = value
	return nil
}

func New() Storage {
	return &storage{
		db: make(map[string]string),
	}
}
