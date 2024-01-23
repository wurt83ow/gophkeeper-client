package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wurt83ow/gophkeeper-client/pkg/bdkeeper"
	"github.com/wurt83ow/gophkeeper-client/pkg/config"
	"github.com/wurt83ow/gophkeeper-client/pkg/encription"
	"github.com/wurt83ow/gophkeeper-client/pkg/gksync"
)

type Service struct {
	keeper *bdkeeper.Keeper
	sync   *gksync.Sync
	enc    *encription.Enc
	opt    *config.Options
}

func NewServices(keeper *bdkeeper.Keeper, sync *gksync.Sync, enc *encription.Enc, opt *config.Options) *Service {
	return &Service{
		keeper: keeper,
		sync:   sync,
		enc:    enc,
		opt:    opt,
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
	hashedPassword, err := s.enc.HashPassword(password)
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
	if !s.enc.CompareHashAndPassword(hashedPassword, password) {
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

func (s *Service) SyncFile(userID int, filePath string) error {
	// Читаем файл
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("Ошибка при чтении файла: %v", err)
	}

	// Отправляем данные на сервер
	err = s.sync.SendFile(userID, data)
	if err != nil {
		return fmt.Errorf("Ошибка при отправке файла на сервер: %v", err)
	}

	return nil
}

func (s *Service) SyncAllData(user_id int) error {
	// Список всех таблиц данных
	tables := []string{"UserCredentials", "CreditCardData", "TextData", "FilesData"}

	// Проходим по каждой таблице
	for _, table := range tables {
		// Получаем все данные из таблицы на сервере
		data, err := s.sync.GetAllData(user_id, table)
		if err != nil {
			return fmt.Errorf("Ошибка при получении данных из таблицы %s: %v", table, err)
		}

		// Очищаем соответствующую таблицу в локальной базе данных
		err = s.keeper.ClearData(user_id, table)
		if err != nil {
			return fmt.Errorf("Ошибка при очистке таблицы %s: %v", table, err)
		}

		// Добавляем все полученные данные в локальную базу данных
		for _, row := range data {
			err = s.keeper.AddData(user_id, table, row)
			if err != nil {
				return fmt.Errorf("Ошибка при добавлении данных в таблицу %s: %v", table, err)
			}
		}
	}

	return nil
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
	// Шифрование каждого значения в данных перед их сохранением
	encryptedData := make(map[string]string)
	for key, value := range data {
		encryptedValue, err := s.enc.Encrypt(value)
		if err != nil {
			return err
		}
		encryptedData[key] = encryptedValue
	}

	err := s.keeper.AddData(user_id, table, encryptedData)
	if err != nil {
		return err
	}
	err = s.sync.AddData(user_id, table, encryptedData)
	if err == gksync.ErrNetworkUnavailable {
		err = s.keeper.MarkForSync(user_id, table, encryptedData)
	}
	return err
}

func (s *Service) UpdateData(user_id int, table string, data map[string]string) error {
	// Шифрование каждого значения в данных перед их обновлением
	encryptedData := make(map[string]string)
	for key, value := range data {
		encryptedValue, err := s.enc.Encrypt(value)
		if err != nil {
			return err
		}
		encryptedData[key] = encryptedValue
	}

	err := s.keeper.UpdateData(user_id, table, encryptedData)
	if err != nil {
		return err
	}
	err = s.sync.UpdateData(user_id, table, encryptedData)
	if err == gksync.ErrNetworkUnavailable {
		err = s.keeper.MarkForSync(user_id, table, encryptedData)
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
	err := s.keeper.ClearData(user_id, table)
	if err != nil {
		return err
	}
	err = s.sync.ClearData(user_id, table)
	if err == gksync.ErrNetworkUnavailable {
		err = s.keeper.MarkForSync(user_id, table, nil)
	}
	return err
}
func (s *Service) DeleteAllFiles() error {

	// Получаем список всех файлов в каталоге
	files, err := os.ReadDir(s.opt.FileStoragePath)
	if err != nil {
		return err
	}

	for _, file := range files {
		// Удаляем файл
		err = os.Remove(filepath.Join(s.opt.FileStoragePath, file.Name()))
		if err != nil {
			return err
		}
	}

	return nil
}
