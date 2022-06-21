package storage

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

type fileStorage struct {
	path string
}

func (fs fileStorage) Read(value string) (string, error) {
	file, err := os.OpenFile(fs.path, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	var long string

	for scanner.Scan() {
		data := scanner.Text()
		fmt.Println("-----", data)
		if strings.Split(data, "|")[0] == value {
			long = strings.Split(data, "|")[1]
		}
	}

	if long != "" {
		fmt.Println("*****", long)
		return long, nil
	}

	err = errors.New("error: there is no url in database")
	return "", err
}

func (fs fileStorage) Write(key, value string) error {
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

func NewFileStorage(path string) Storage {
	return &fileStorage{
		path: path,
	}
}
