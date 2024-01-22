package services

import (
	"crypto/sha256"
	"errors"

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

func (s *Service) Register(username string, password string) error {
	// Проверка наличия пользователя в базе данных
	userExists, err := s.keeper.UserExists(username)
	if err != nil {
		return err
	}
	if userExists {
		return errors.New("User already exists")
	}

	// Хеширование пароля
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return err
	}

	// Сохранение нового пользователя в базе данных
	err = s.keeper.AddUser(username, hashedPassword)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Login(username string, password string) (int, error) {
	// Попытка получить хешированный пароль пользователя из локальной базы данных
	hashedPassword, err := s.keeper.GetPassword(username)
	if err != nil {
		// Если пользователь не найден в локальной базе данных, попытка получить данные из сервера
		hashedPassword, err = s.sync.GetPassword(username)
		if err != nil {
			return 0, err
		}
	}

	// Сравнение хешированного пароля с хешем введенного пароля
	if !s.keeper.CompareHashAndPassword(hashedPassword, password) {
		return 0, errors.New("Invalid password")
	}

	// Если пароли совпадают, получаем идентификатор пользователя
	userID, err := s.keeper.GetUserID(username)
	if err != nil {
		// Если идентификатор пользователя не найден в локальной базе данных, попытка получить его из сервера
		userID, err = s.sync.GetUserID(username)
		if err != nil {
			return 0, err
		}
	}

	// Возвращаем идентификатор пользователя
	return userID, nil
}

// Перенести в модуль encription
func HashPassword(password string) (string, error) {
	hash := sha256.New()
	_, err := hash.Write([]byte(password))
	if err != nil {
		return "", err
	}
	return string(hash.Sum(nil)), nil
}
func (s *Service) InitSync(user_id int, table string, columns ...string) error {
	data, err := s.sync.GetAllData(user_id, table)
	if err != nil {
		return err
	}
	for _, item := range data {
		err = s.keeper.AddData(user_id, table, item)
		if err != nil {
			return err
		}
	}
	return nil
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
