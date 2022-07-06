package storage

import (
	"errors"
)

type storage struct {
	db   map[string]string
	byID map[string][]string
}

func (s storage) Read(value string) (string, error) {
	if url, ok := s.db[value]; ok {
		return url, nil
	}
	err := errors.New("error: there is no url in database")
	return "", err
}

func (s storage) Write(id, key, value string) error {
	s.db[key] = value
	if _, ok := s.byID[id]; !ok {
		s.byID[id] = make([]string, 1)
	}
	s.byID[id] = append(s.byID[id], key)
	return nil
}

func (s storage) ReadAll(id string) (map[string]string, error) {
	if keys, ok := s.byID[id]; ok {
		result := make(map[string]string)
		for _, key := range keys {
			result[key] = s.db[key]
		}
		return result, nil
	}
	return nil, errors.New("no data")
}

func New() Storage {
	return &storage{
		db: make(map[string]string),
	}
}
