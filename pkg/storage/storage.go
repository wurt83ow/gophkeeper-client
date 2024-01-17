package storage

import "github.com/wurt83ow/gophkeeper-client/pkg/bdkeeper"

type Storage struct {
	data   map[string]string
	keeper *bdkeeper.Keeper
}

func New() *Storage {
	return &Storage{
		data:   make(map[string]string),
		keeper: bdkeeper.New(),
	}
}

func (s *Storage) AddData(key string, value string) {
	s.keeper.AddData(key, value)
	s.data[key] = value
}

func (s *Storage) UpdateData(key string, value string) {
	_, exists := s.data[key]
	if exists {
		s.keeper.UpdateData(key, value)
		s.data[key] = value
	}
}

func (s *Storage) DeleteData(key string) {
	s.keeper.DeleteData(key)
	delete(s.data, key)
}

func (s *Storage) GetData(key string) (string, bool) {
	_, exists := s.data[key]
	if exists {
		return s.keeper.GetData(key), exists
	}
	return "", false
}

func (s *Storage) GetAllData() map[string]string {
	return s.keeper.GetAllData()
}

func (s *Storage) ClearData() {
	s.keeper.ClearData()
	s.data = make(map[string]string)
}
