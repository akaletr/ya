package storage

import (
	"bufio"
	"cmd/shortener/main.go/internal/model"
	"errors"
	"fmt"
	"os"
	"strings"
)

type fileStorage struct {
	path string
}

func (fs fileStorage) Read(id, value string) (string, error) {
	file, err := os.OpenFile(fs.path, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		data := scanner.Text()
		if strings.Split(data, "|")[0] == value {
			return strings.Split(data, "|")[1], nil
		}
	}

	err = errors.New("error: there is no url in database")
	return "", err
}

func (fs fileStorage) Write(id, key, value string) error {
	file, err := os.OpenFile(fs.path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	data := fmt.Sprintf("%s|%s\n", key, value)
	_, err = file.Write([]byte(data))
	return err
}

func (s fileStorage) ReadAll(id, baseURL string) (model.AllShortenerRequest, error) {
	return nil, nil
}

func NewFileStorage(path string) Storage {
	return &fileStorage{
		path: path,
	}
}
