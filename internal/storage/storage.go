package storage

import (
	"cmd/shortener/main.go/internal/model"
	"errors"
	"fmt"
)

type storage struct {
	db map[string]map[string]string
}

func (s storage) Read(id, value string) (string, error) {
	if user, ok := s.db[id]; ok {
		return user[value], nil
	}
	err := errors.New("error: there is no url in database")
	return "", err
}

func (s storage) Write(id, key, value string) error {
	if len(s.db[id]) == 0 {
		s.db[id] = map[string]string{}
	}
	s.db[id][key] = value
	return nil
}

func (s storage) ReadAll(id, baseURL string) (model.AllShortenerRequest, error) {
	data := model.AllShortenerRequest{}
	if user, ok := s.db[id]; ok {
		for key, value := range user {
			item := model.Item{
				ShortURL:    fmt.Sprintf("%s/%s", baseURL, key),
				OriginalURL: value,
			}
			data = append(data, item)
		}

	}

	fmt.Println(data)
	return data, nil
}

func New() Storage {
	return &storage{
		db: make(map[string]map[string]string),
	}
}
