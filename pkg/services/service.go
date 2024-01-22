package services

import (
	"github.com/wurt83ow/gophkeeper-client/pkg/bdkeeper"
	"github.com/wurt83ow/gophkeeper-client/pkg/gksync"
)

type Services struct {
	keeper *bdkeeper.Keeper
	sync   *gksync.Sync
}

func New(keeper *bdkeeper.Keeper, sync *gksync.Sync) *Services {
	return &Services{
		keeper: keeper,
		sync:   sync,
	}
}

func (s *Services) GetData(user_id int, table string, columns ...string) (map[string]string, error) {
	data, err := s.keeper.GetData(user_id, table, columns...)
	if err != nil {
		return nil, err
	}
	err = s.sync.SyncData(user_id, table, data)
	if err == s.sync.ErrNetworkUnavailable {
		err = s.keeper.MarkForSync(user_id, table, data)
	}
	return data, err
}

func (s *Services) AddData(user_id int, table string, data map[string]string) error {
	err := s.keeper.AddData(user_id, table, data)
	if err != nil {
		return err
	}
	err = s.sync.SyncData(user_id, table, data)
	if err == s.sync.ErrNetworkUnavailable {
		err = s.keeper.MarkForSync(user_id, table, data)
	}
	return err
}

func (s *Services) UpdateData(user_id int, table string, data map[string]string) error {
	err := s.keeper.UpdateData(user_id, table, data)
	if err != nil {
		return err
	}
	err = s.sync.SyncData(user_id, table, data)
	if err == s.sync.ErrNetworkUnavailable {
		err = s.keeper.MarkForSync(user_id, table, data)
	}
	return err
}

func (s *Services) DeleteData(user_id int, table string, id string, meta_info string) error {
	err := s.keeper.DeleteData(user_id, table, id, meta_info)
	if err != nil {
		return err
	}
	err = s.sync.DeleteData(user_id, table, id)
	if err == s.sync.ErrNetworkUnavailable {
		err = s.keeper.MarkForSync(user_id, table, id)
	}
	return err
}

func (s *Services) GetAllData(table string, columns ...string) ([]map[string]string, error) {
	data, err := s.keeper.GetAllData(table, columns...)
	if err != nil {
		return nil, err
	}
	err = s.sync.SyncAllData(table, data)
	if err == s.sync.ErrNetworkUnavailable {
		for _, item := range data {
			err = s.keeper.MarkForSync(user_id, table, item)
			if err != nil {
				break
			}
		}
	}
	return data, err
}

func (s *Services) ClearData(table string) error {
	err := s.keeper.ClearData(table)
	if err != nil {
		return err
	}
	err = s.sync.ClearData(table)
	if err == s.sync.ErrNetworkUnavailable {
		err = s.keeper.MarkForSync(user_id, table, nil)
	}
	return err
}
