package services

import (
	"github.com/wurt83ow/gophkeeper-client/pkg/bdkeeper"
	"github.com/wurt83ow/gophkeeper-client/pkg/gksync"
)

type Service struct {
	keeper *bdkeeper.Keeper
	sync   *gksync.Sync
}

func NewServices(keeper *bdkeeper.Keeper, sync *gksync.Sync) *Service {
	return &Service{
		keeper: keeper,
		sync:   sync,
	}
}

func (s *Service) GetData(user_id int, table string, columns ...string) (map[string]string, error) {
	data, err := s.keeper.GetData(user_id, table, columns...)
	if err != nil {
		return nil, err
	}
	err = s.sync.GetData(user_id, table, data)
	if err == gksync.ErrNetworkUnavailable {
		err = s.keeper.MarkForSync(user_id, table, data)
	}
	return data, err
}

func (s *Service) AddData(user_id int, table string, data map[string]string) error {
	err := s.keeper.AddData(user_id, table, data)
	if err != nil {
		return err
	}
	err = s.sync.AddData(user_id, table, data)
	if err == gksync.ErrNetworkUnavailable {
		err = s.keeper.MarkForSync(user_id, table, data)
	}
	return err
}

func (s *Service) UpdateData(user_id int, table string, data map[string]string) error {
	err := s.keeper.UpdateData(user_id, table, data)
	if err != nil {
		return err
	}
	err = s.sync.UpdateData(user_id, table, data)
	if err == gksync.ErrNetworkUnavailable {
		err = s.keeper.MarkForSync(user_id, table, data)
	}
	return err
}

func (s *Service) DeleteData(user_id int, table string, id string, meta_info string) error {
	err := s.keeper.DeleteData(user_id, table, id, meta_info)
	if err != nil {
		return err
	}
	err = s.sync.DeleteData(user_id, table, id)
	if err == gksync.ErrNetworkUnavailable {
		data := map[string]string{"id": id}
		err = s.keeper.MarkForSync(user_id, table, data)
	}
	return err
}

func (s *Service) GetAllData(user_id int, table string, columns ...string) ([]map[string]string, error) {
	var data []map[string]string
	var err error

	// Попытка получить данные из sync, если сеть доступна
	data, err = s.sync.GetAllData(user_id, table)
	if err != nil {
		if err == gksync.ErrNetworkUnavailable {
			// Если сеть недоступна, получить данные из keeper
			data, err = s.keeper.GetAllData(table, columns...)
			if err != nil {
				return nil, err
			}

			// Пометить все элементы данных для синхронизации
			for _, item := range data {
				err = s.keeper.MarkForSync(user_id, table, item)
				if err != nil {
					return nil, err // Возвращаем ошибку, если не удалось пометить элемент для синхронизации
				}
			}
		} else {
			return nil, err // Возвращаем ошибку, если она не связана с недоступностью сети
		}
	}

	return data, nil // Возвращаем данные без ошибок
}

func (s *Service) ClearData(user_id int, table string) error {
	err := s.keeper.ClearData(table)
	if err != nil {
		return err
	}
	err = s.sync.ClearData(user_id, table)
	if err == gksync.ErrNetworkUnavailable {
		err = s.keeper.MarkForSync(user_id, table, nil)
	}
	return err
}
