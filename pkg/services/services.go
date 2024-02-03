package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/wurt83ow/gophkeeper-client/pkg/bdkeeper"
	"github.com/wurt83ow/gophkeeper-client/pkg/config"
	"github.com/wurt83ow/gophkeeper-client/pkg/encription"
	"github.com/wurt83ow/gophkeeper-client/pkg/gksync"
)

type Service struct {
	keeper         *bdkeeper.Keeper
	sync           *gksync.Sync
	sync1          *gksync.ClientWithResponses
	enc            *encription.Enc
	opt            *config.Options
	syncWithServer bool
}

func NewServices(keeper *bdkeeper.Keeper, sync *gksync.Sync, sync1 *gksync.ClientWithResponses,
	enc *encription.Enc, opt *config.Options, syncWithServer bool) *Service {
	return &Service{
		keeper:         keeper,
		sync:           sync,
		sync1:          sync1,
		enc:            enc,
		opt:            opt,
		syncWithServer: syncWithServer,
	}
}

func (s *Service) Register(ctx context.Context, username string, password string) error {
	// Проверка наличия пользователя в базе данных
	userExists, err := s.keeper.UserExists(ctx, username)
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
	err = s.keeper.AddUser(ctx, username, hashedPassword)
	if err != nil {
		return err
	}

	// Сохранение нового пользователя на сервере
	if s.syncWithServer {
		body := gksync.PostRegisterJSONRequestBody{
			Username: username,
			Password: hashedPassword,
		}
		_, err = s.sync1.PostRegister(ctx, body)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) Login(ctx context.Context, username string, password string) (int, string, error) {
	var userID int
	var token string
	var err error

	// Попытка получить хешированный пароль пользователя из локальной базы данных
	hashedPassword, err := s.keeper.GetPassword(ctx, username)
	if err != nil {
		return 0, "", err
	}

	// Сравнение хешированного пароля с хешем введенного пароля
	if !s.enc.CompareHashAndPassword(hashedPassword, password) {
		return 0, "", errors.New("Invalid password")
	}

	if s.syncWithServer {
		// Если syncWithServer=true, получаем userID и jwtToken с сервера
		body := gksync.PostLoginJSONRequestBody{
			Username: username,
			Password: hashedPassword,
		}
		resp, err := s.sync1.PostLoginWithResponse(ctx, body)
		if err != nil {
			return 0, "", err
		}

		userID = resp.JSON200.UserID
		token = string(resp.JSON200.Token)
	} else {
		// Если syncWithServer=false, получаем только userID из keeper
		userID, err = s.keeper.GetUserID(ctx, username)
		if err != nil {
			return 0, "", errors.New("Invalid userId")
		}
	}

	// Возвращаем идентификатор пользователя и токен
	return userID, token, nil
}

func (s *Service) SyncFile(userID int, filePath string) error {
	if !s.syncWithServer {
		return nil
	}
	// Отправляем данные на сервер
	err := s.sync.SendFile(userID, filePath)
	if err != nil {
		return fmt.Errorf("Ошибка при отправке файла на сервер: %v", err)
	}

	return nil
}

func (s *Service) SyncAllData(ctx context.Context, user_id int) error {
	if !s.syncWithServer {
		return nil
	}

	// Список всех таблиц данных
	tables := []string{"UserCredentials", "CreditCardData", "TextData", "FilesData"}

	// Проходим по каждой таблице
	for _, table := range tables {
		// Получаем все данные из таблицы на сервере
		data, err := s.sync.GetAllData(table, user_id)
		if err != nil {
			return fmt.Errorf("Ошибка при получении данных из таблицы %s: %v", table, err)
		}

		// Очищаем соответствующую таблицу в локальной базе данных
		err = s.keeper.ClearData(ctx, user_id, table)
		if err != nil {
			return fmt.Errorf("Ошибка при очистке таблицы %s: %v", table, err)
		}

		// Добавляем все полученные данные в локальную базу данных
		for _, row := range data {
			entry_id, err := s.GenerateUUID(ctx)
			if err != nil {
				fmt.Println("Ошибка при генерации UUID:", err)
				continue
			}
			err = s.keeper.AddData(ctx, table, user_id, entry_id, row)
			if err != nil {
				return fmt.Errorf("Ошибка при добавлении данных в таблицу %s: %v", table, err)
			}
		}
	}

	return nil
}
func (s *Service) InitSync(ctx context.Context, table string, user_id int, columns ...string) error {
	if !s.syncWithServer {
		return nil
	}

	data, err := s.sync.GetAllData(table, user_id)
	if err != nil {
		return err
	}
	for _, item := range data {
		entry_id, err := s.GenerateUUID(ctx)
		if err != nil {
			fmt.Println("Ошибка при генерации UUID:", err)
			continue
		}
		err = s.keeper.AddData(ctx, table, user_id, entry_id, item)
		if err != nil {
			return err
		}
	}
	return nil
}

// func (s *Service) GetData(ctx context.Context, user_id int, table string, id string) (map[string]string, error) {
// 	// Получаем данные из keeper
// 	data, err := s.keeper.GetData(ctx, user_id, table, id)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Пытаемся синхронизировать данные
// 	if s.syncWithServer {
// 		resp, err := s.sync1.GetGetDataTableUserID(ctx, table, user_id)
// 		if err == gksync.ErrNetworkUnavailable {
// 			// Если сеть недоступна, помечаем данные для синхронизации
// 			err = s.keeper.MarkForSync(ctx, user_id, table, data)
// 			if err != nil {
// 				return nil, err
// 			}
// 		} else if resp.StatusCode != http.StatusOK {
// 			// Обработка ошибок HTTP
// 			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
// 		}
// 	}

// 	// Расшифровка данных перед возвратом
// 	for key, value := range data {
// 		decryptedValue, err := s.enc.Decrypt(value)
// 		if err != nil {
// 			return nil, err
// 		}
// 		data[key] = decryptedValue
// 	}

//		return data, err
//	}
func (s *Service) GetData(ctx context.Context, table string, user_id int, entry_id string) (map[string]string, error) {
	var data map[string]string
	var err error

	// Пытаемся синхронизировать данные
	if s.syncWithServer {
		resp, err := s.sync1.GetGetDataTableUserIDWithResponse(ctx, table, user_id, entry_id)
		if err != nil && err != gksync.ErrNetworkUnavailable {
			// Обработка ошибок HTTP и других ошибок, кроме недоступности сети
			return nil, fmt.Errorf("unexpected error: %v", err)
		}

		if resp.StatusCode() == http.StatusOK {
			// Если данные успешно получены с сервера, используем их
			serverData := *resp.JSON200 // предполагается, что resp.Body содержит данные

			// Обновляем локальную базу данных новыми данными
			err = s.keeper.UpdateData(ctx, table, user_id, entry_id, serverData)
			if err != nil {
				return nil, err
			}

			data = serverData

		} else if err == gksync.ErrNetworkUnavailable {
			// Если сеть недоступна, помечаем данные для синхронизации
			err = s.keeper.MarkForSync(ctx, table, user_id, entry_id, data)
			if err != nil {
				return nil, err
			}
		}
	}

	if data == nil {
		// Если данные не были получены с сервера, получаем данные из keeper
		data, err = s.keeper.GetData(ctx, table, user_id, entry_id)
		if err != nil {
			return nil, err
		}
	}

	// Расшифровка данных перед возвратом
	for key, value := range data {
		decryptedValue, err := s.enc.Decrypt(value)
		if err != nil {
			return nil, err
		}
		data[key] = decryptedValue
	}

	return data, nil
}

func (s *Service) GenerateUUID(ctx context.Context) (string, error) {
	// Генерируем UUID
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	// Возвращаем UUID в виде строки
	return uuid.String(), nil
}

func (s *Service) AddData(ctx context.Context, table string, user_id int, data map[string]string) error {

	entry_id, err := s.GenerateUUID(ctx)
	if err != nil {
		return err
	}

	// Шифрование каждого значения в данных перед их сохранением
	encryptedData := make(map[string]string)
	for key, value := range data {
		encryptedValue, err := s.enc.Encrypt(value)
		if err != nil {
			return err
		}
		encryptedData[key] = encryptedValue
	}

	err = s.keeper.AddData(ctx, table, user_id, entry_id, encryptedData)
	if err != nil {
		return err
	}
	if s.syncWithServer {
		_, err = s.sync1.PostAddDataTableUserID(ctx, table, user_id, entry_id, encryptedData)
		if err == gksync.ErrNetworkUnavailable {
			err = s.keeper.MarkForSync(ctx, table, user_id, entry_id, encryptedData)
		}
	}

	return err
}

func (s *Service) UpdateData(ctx context.Context, table string, user_id int, entry_id string, data map[string]string) error {
	// Шифрование каждого значения в данных перед их обновлением
	encryptedData := make(map[string]string)
	for key, value := range data {
		encryptedValue, err := s.enc.Encrypt(value)
		if err != nil {
			return err
		}
		encryptedData[key] = encryptedValue
	}

	err := s.keeper.UpdateData(ctx, table, user_id, entry_id, encryptedData)
	if err != nil {
		return err
	}
	if s.syncWithServer {
		err = s.sync.UpdateData(table, user_id, entry_id, encryptedData)
		if err == gksync.ErrNetworkUnavailable {
			err = s.keeper.MarkForSync(ctx, table, user_id, entry_id, encryptedData)
		}
	}

	return err
}

func (s *Service) DeleteData(ctx context.Context, table string, user_id int, entry_id string) error {
	err := s.keeper.DeleteData(ctx, table, user_id, entry_id)
	if err != nil {
		return err
	}
	if s.syncWithServer {
		err = s.sync.DeleteData(table, user_id, entry_id)
		if err == gksync.ErrNetworkUnavailable {
			data := map[string]string{"id": entry_id}
			err = s.keeper.MarkForSync(ctx, table, user_id, entry_id, data)
		}
	}
	return err
}

func (s *Service) GetAllData(ctx context.Context, table string, user_id int, columns ...string) ([]map[string]string, error) {
	var data []map[string]string
	var err error

	// Попытка получить данные из keeper
	data, err = s.keeper.GetAllData(ctx, table, user_id, columns...)
	if err != nil && s.syncWithServer {
		// Если данные не удалось получить из keeper, попытка получить данные из sync
		data, err := s.sync1.GetGetAllDataTableUserIDWithResponse(ctx, table, user_id)

		if err != nil {
			if err == gksync.ErrNetworkUnavailable {
				// Если сеть недоступна, возвращаем ошибку
				return nil, err
			}

			// Пометить все элементы данных для синхронизации
			for _, item := range *data.JSON200 {
				var entry_id string
				if value, ok := item["id"]; ok {
					entry_id = value
				}
				err = s.keeper.MarkForSync(ctx, table, user_id, entry_id, item)
				if err != nil {
					return nil, err // Возвращаем ошибку, если не удалось пометить элемент для синхронизации
				}
			}
		}
	}

	// Расшифровка данных перед возвратом
	for i, item := range data {
		for key, value := range item {
			if key != "id" {
				decryptedValue, err := s.enc.Decrypt(value)
				if err != nil {
					return nil, err
				}
				data[i][key] = decryptedValue
			}
		}
	}

	return data, nil // Возвращаем данные без ошибок
}

func (s *Service) ClearData(ctx context.Context, table string, user_id int) error {
	err := s.keeper.ClearData(ctx, user_id, table)
	if err != nil {
		return err
	}
	if s.syncWithServer {
		err = s.sync.ClearData(table, user_id)
		if err == gksync.ErrNetworkUnavailable {
			// err = s.keeper.MarkForSync(ctx, table, user_id, nil)
		}
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
