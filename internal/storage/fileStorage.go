package storage

import (
	"bufio"
	"cmd/shortener/main.go/internal/model"
	"encoding/json"
	"errors"
	"os"
)

type fileStorage struct {
	path string
}

func (fs fileStorage) Read(value string) (model.Note, error) {
	file, err := os.OpenFile(fs.path, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return model.Note{}, err
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		noteJSON := scanner.Text()

		noteTemp := model.Note{}
		err = json.Unmarshal([]byte(noteJSON), &noteTemp)
		if err != nil {
			return model.Note{}, err
		}

		if noteTemp.ID == value {
			return noteTemp, nil
		}
	}

	err = errors.New("error: there is no url in database")
	return model.Note{}, err
}

func (fs fileStorage) Write(note model.Note) error {
	file, err := os.OpenFile(fs.path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		noteJSON := scanner.Text()
		noteTemp := model.Note{}
		err = json.Unmarshal([]byte(noteJSON), &noteTemp)
		if err != nil {
			return err
		}

		if note.Short == noteTemp.Short {
			return NewError(CONFLICT, "conflict")
		}
	}

	noteJSON, err := json.Marshal(note)
	_, err = file.Write(noteJSON)
	return err
}

func (fs fileStorage) ReadAll(id string) (map[string]string, error) {
	file, err := os.OpenFile(fs.path, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	result := map[string]string{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		noteJSON := scanner.Text()
		noteTemp := model.Note{}
		err = json.Unmarshal([]byte(noteJSON), &noteTemp)

		result[noteTemp.Short] = noteTemp.Long
	}

	return result, nil
}

func (fs fileStorage) WriteBatch(data []model.DataBatchItem) error {
	var e error
	for _, item := range data {
		note := model.Note{
			ID:    item.ID,
			Short: item.Short,
			Long:  item.Long,
		}
		err := fs.Write(note)
		if err != nil {
			e = err
		}
	}
	return e
}

func (fs fileStorage) Delete(note model.Note) error {
	return nil
}

func (fs fileStorage) Start() error {
	return nil
}

func (fs fileStorage) Ping() error {
	return nil
}

func NewFileStorage(path string) Storage {
	return &fileStorage{
		path: path,
	}
}
