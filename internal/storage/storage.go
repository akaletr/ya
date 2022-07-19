package storage

import (
	"cmd/shortener/main.go/internal/model"
	"errors"
	"fmt"
)

type storage struct {
	db   map[string]model.Note
	byID map[string][]model.Note
}

func (s storage) Read(value string) (model.Note, error) {
	if note, ok := s.db[value]; ok {
		return note, nil
	}
	err := errors.New("error: there is no url in database")
	return model.Note{}, err
}

func (s storage) Write(note model.Note) error {
	if _, ok := s.db[note.Short]; ok {
		return NewError(CONFLICT, "conflict")
	}

	s.db[note.Short] = note
	if s.byID[note.ID] == nil {
		s.byID[note.ID] = []model.Note{}
	}
	s.byID[note.ID] = append(s.byID[note.ID], note)
	return nil
}

func (s storage) WriteBatch(data []model.DataBatchItem) error {
	for _, item := range data {
		note := model.Note{
			ID:          item.ID,
			Short:       item.Short,
			Long:        item.Long,
			Correlation: item.Correlation,
			Deleted:     false,
		}
		s.db[item.Short] = note
		if s.byID[item.ID] == nil {
			s.byID[item.ID] = []model.Note{}
		}
		s.byID[item.ID] = append(s.byID[item.ID], note)
	}
	return nil
}
func (s storage) ReadAll(id string) (map[string]string, error) {
	if notes, ok := s.byID[id]; ok {
		result := make(map[string]string)
		for _, note := range notes {
			result[note.Short] = s.db[note.Short].Long
		}
		return result, nil
	}
	return nil, errors.New("no data")
}

func (s storage) Delete(note model.Note) error {
	if s.db[note.Short].ID == note.ID {
		noteTemp := s.db[note.Short]
		noteTemp.Deleted = true
		s.db[note.Short] = noteTemp
		fmt.Println(s.db[note.Short])
	}
	return nil
}

func (s storage) Start() error {
	return nil
}

func (s storage) Ping() error {
	return nil
}

func New() Storage {
	return &storage{
		db:   make(map[string]model.Note),
		byID: map[string][]model.Note{},
	}
}
