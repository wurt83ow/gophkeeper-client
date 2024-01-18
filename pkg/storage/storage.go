package storage

import (
	"github.com/wurt83ow/gophkeeper-client/pkg/bdkeeper"
)

type Storage struct {
	data   map[string]string
	keeper *bdkeeper.Keeper
}

func New(keeper *bdkeeper.Keeper) *Storage {
	return &Storage{
		data:   make(map[string]string),
		keeper: keeper,
	}
}

func (s *Storage) AddData(user_id int, table string, data map[string]string) {
	s.keeper.AddData(user_id, table, data)
	s.data[table] = data["value"]
}

func (s *Storage) UpdateData(user_id int, table string, data map[string]string) {
	_, exists := s.data[table]
	if exists {
		s.keeper.UpdateData(user_id, table, data)
		s.data[table] = data["value"]
	}
}

func (s *Storage) DeleteData(user_id int, table string) {
	s.keeper.DeleteData(user_id, table)
	delete(s.data, table)
}

func (s *Storage) GetData(user_id int, table string, columns ...string) (map[string]string, bool) {
	_, exists := s.data[table]
	if exists {
		return s.keeper.GetData(user_id, table, columns...), exists
	}
	return nil, false
}

func (s *Storage) GetAllData(table string, columns ...string) []map[string]string {
	return s.keeper.GetAllData(table, columns...)
}

func (s *Storage) ClearData(table string) {
	s.keeper.ClearData(table)
	s.data = make(map[string]string)
}
