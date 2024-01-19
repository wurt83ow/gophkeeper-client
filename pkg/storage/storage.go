package storage

import (
	"github.com/wurt83ow/gophkeeper-client/pkg/bdkeeper"
)

type Storage struct {
	keeper *bdkeeper.Keeper
}

func New(keeper *bdkeeper.Keeper) *Storage {
	return &Storage{
		keeper: keeper,
	}
}

func (s *Storage) AddData(user_id int, table string, data map[string]string) error {
	return s.keeper.AddData(user_id, table, data)
}

func (s *Storage) UpdateData(user_id int, table string, data map[string]string) error {
	return s.keeper.UpdateData(user_id, table, data)
}

func (s *Storage) DeleteData(user_id int, table string, id string, meta_info string) error {
	return s.keeper.DeleteData(user_id, table, id, meta_info)
}

func (s *Storage) GetData(user_id int, table string, columns ...string) (map[string]string, error) {
	return s.keeper.GetData(user_id, table, columns...)
}

func (s *Storage) GetAllData(table string, columns ...string) ([]map[string]string, error) {
	return s.keeper.GetAllData(table, columns...)
}

func (s *Storage) ClearData(table string) error {
	return s.keeper.ClearData(table)
}
