package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/wurt83ow/gophkeeper-client/pkg/bdkeeper"
	"github.com/wurt83ow/gophkeeper-client/pkg/config"
	"github.com/wurt83ow/gophkeeper-client/pkg/encription"
	"github.com/wurt83ow/gophkeeper-client/pkg/gksync"
	"github.com/wurt83ow/gophkeeper-client/pkg/models"
)

type Service struct {
	keeper *bdkeeper.Keeper

	sync           *gksync.ClientWithResponses
	enc            *encription.Enc
	opt            *config.Options
	syncWithServer bool
	logger         Logger
}

type Logger interface {
	Printf(format string, v ...interface{})
}

func NewServices(keeper *bdkeeper.Keeper, sync *gksync.ClientWithResponses,
	enc *encription.Enc, opt *config.Options, syncWithServer bool, logger Logger) *Service {
	return &Service{
		keeper: keeper,

		sync:           sync,
		enc:            enc,
		opt:            opt,
		syncWithServer: syncWithServer,
		logger:         logger,
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
	fmt.Println("hashedPassword777register      ", hashedPassword)
	// Сохранение нового пользователя на сервере
	if s.syncWithServer {
		body := gksync.PostRegisterJSONRequestBody{
			Username: username,
			Password: hashedPassword,
		}
		_, err = s.sync.PostRegister(ctx, body)
		if err != nil {
			return err
		}
	}

	// Сохранение нового пользователя в базе данных
	err = s.keeper.AddUser(ctx, username, hashedPassword)
	if err != nil {
		return err
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
	fmt.Println("hashedPassword777login      ", hashedPassword)

	if s.syncWithServer {
		// Если syncWithServer=true, получаем userID и jwtToken с сервера
		body := gksync.PostLoginJSONRequestBody{
			Username: username,
			Password: hashedPassword,
		}
		resp, err := s.sync.PostLoginWithResponse(ctx, body)
		if err != nil {
			return 0, "", err
		}
		// Проверяем, равен ли resp nil
		if resp.JSON200 == nil {
			return 0, "", fmt.Errorf("Unauthorized")
		}

		userID = *resp.JSON200.UserID
		token = string(*resp.JSON200.Token)
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

func (s *Service) SyncFile(ctx context.Context, userID int, filePath string) {
	if !s.syncWithServer {
		return
	}

	// Открываем файл
	file, err := os.Open(filePath)
	if err != nil {
		s.logger.Printf("Ошибка при открытии файла: %v", err)
		return
	}
	defer file.Close()

	// Отправляем данные на сервер
	_, err = s.sync.PostSendFileUserIDWithBody(ctx, userID, "application/octet-stream", file)
	if err != nil {
		s.logger.Printf("Ошибка при отправке файла на сервер: %v", err)
	}
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
		resp, err := s.sync.GetGetAllDataTableUserIDWithResponse(ctx, table, user_id)
		if err != nil {
			s.logger.Printf("Ошибка при получении данных из таблицы %s: %v", table, err)
		}

		// Очищаем соответствующую таблицу в локальной базе данных
		err = s.keeper.ClearData(ctx, table, user_id)
		if err != nil {
			s.logger.Printf("Ошибка при очистке таблицы %s: %v", table, err)
		}

		// Добавляем все полученные данные в локальную базу данных
		for _, row := range *resp.JSON200 {
			err = s.keeper.AddData(ctx, table, user_id, row["entry_id"], row)
			if err != nil {
				s.logger.Printf("Ошибка при добавлении данных в таблицу %s: %v", table, err)
			}
		}
	}

	return nil
}

func (s *Service) SyncAllWithServer(ctx context.Context) {
	fmt.Println("888888888888888888888888888888888888888888888888")
	// Получаем все записи из таблицы синхронизации со статусом "В ожидании"
	entries, err := s.keeper.GetPendingSyncEntries(ctx)
	if err != nil {
		fmt.Println("errerrerrerrerrerrerrerrerrerrerrerrerrerrerrerrerr", err)
		return
	}

	for _, entry := range entries {
		err = s.syncData(ctx, entry)

		// Если запрос успешно выполнен, обновляем статус записи на "Done"
		if err == nil {
			err = s.keeper.UpdateSyncEntryStatus(ctx, entry.ID, "Done")
			if err != nil {
				return
			}
		} else {
			// Если произошла ошибка, обрабатываем ее
			s.handleSyncError(ctx, err, entry)
		}
	}
}

// handleSyncError обрабатывает ошибки, возникающие при синхронизации данных с сервером
func (s *Service) handleSyncError(ctx context.Context, err error, entry models.SyncQueue) {
	// Логирование ошибки
	s.logger.Printf("Ошибка при синхронизации данных: %s\n", err)

	// Повторение попытки синхронизации
	retryCount := 0
	for retryCount < 3 {
		err = s.syncData(ctx, entry)

		if err == nil {
			// Если запрос успешно выполнен, обновляем статус записи на "Done"
			err = s.keeper.UpdateSyncEntryStatus(ctx, entry.ID, "Done")
			if err != nil {
				s.logger.Printf("Ошибка при обновлении статуса записи: %s\n", err)
			}
			break
		} else {
			retryCount++
			s.logger.Printf("Ошибка при повторной попытке синхронизации данных: %s\n", err)
		}
	}
}

// syncData синхронизирует данные с сервером
func (s *Service) syncData(ctx context.Context, entry models.SyncQueue) error {
	fmt.Println("9999999999999999999999999999999999999999999999999999")
	bodyReader := bytes.NewReader([]byte(entry.Data))
	switch entry.Operation {
	case "Create":
		_, err := s.sync.PostAddDataTableUserIDEntryIDWithBody(ctx, entry.TableName, entry.UserID, entry.EntryID, "application/json", bodyReader)
		return err
	case "Update":
		_, err := s.sync.PutUpdateDataTableUserIDEntryIDWithBody(ctx, entry.TableName, entry.UserID, entry.EntryID, "application/json", bodyReader)
		return err
	case "Delete":
		_, err := s.sync.DeleteDeleteDataTableUserIDEntryID(ctx, entry.TableName, entry.UserID, entry.EntryID)
		return err
	}
	return nil
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
	fmt.Println("222222222222222222222222222222222222222222222222222222222", table, user_id)
	err = s.keeper.AddData(ctx, table, user_id, entry_id, encryptedData)

	if s.syncWithServer && err == nil {
		err = s.keeper.CreateSyncEntry(ctx, "Create", table, user_id, entry_id, encryptedData)
		if err == nil {
			go s.SyncAllWithServer(ctx)
		}
	}

	return err
}

func (s *Service) GetData(ctx context.Context, table string, user_id int, entry_id string) (map[string]string, error) {

	// Получаем данные из keeper
	data, err := s.keeper.GetData(ctx, table, user_id, entry_id)
	if err != nil {
		return nil, err
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

	if s.syncWithServer && err == nil {
		err = s.keeper.CreateSyncEntry(ctx, "Update", table, user_id, entry_id, encryptedData)
		if err == nil {
			go s.SyncAllWithServer(ctx)
		}
	}

	return err
}

func (s *Service) DeleteData(ctx context.Context, table string, user_id int, entry_id string) error {
	err := s.keeper.DeleteData(ctx, table, user_id, entry_id)

	if s.syncWithServer && err == nil {
		data := map[string]string{"id": entry_id}
		err = s.keeper.CreateSyncEntry(ctx, "Delete", table, user_id, entry_id, data)
		if err == nil {
			go s.SyncAllWithServer(ctx)
		}
	}

	return err
}

func (s *Service) GetAllData(ctx context.Context, table string, user_id int, columns ...string) ([]map[string]string, error) {

	// Попытка получить данные из keeper
	data, err := s.keeper.GetAllData(ctx, table, user_id, columns...)
	if err != nil {
		return nil, err
	}
	// data, err := s.sync.GetGetAllDataTableUserIDWithResponse(ctx, table, user_id)

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

func (s *Service) RetrieveFile(ctx context.Context, user_id int, entry_id string, filePath string) {

	// Получаем данные с сервера
	resp, err := s.sync.GetGetFileUserIDEntryID(ctx, user_id, entry_id)
	if err != nil {
		s.logger.Printf("Ошибка при получении файла с сервера: %v", err)
		return
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		s.logger.Printf("Сервер вернул неожиданный статус: %v", resp.Status)
		return
	}

	// Создаем файл для сохранения данных
	out, err := os.Create(filePath)
	if err != nil {
		s.logger.Printf("Ошибка при создании файла: %v", err)
		return
	}
	defer out.Close()

	// Копируем данные из ответа в файл
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		s.logger.Printf("Ошибка при сохранении файла: %v", err)
		return
	}

	s.logger.Printf("Файл успешно получен и сохранен!")
}

func (s *Service) ClearLocalData(ctx context.Context, table string, user_id int) error {
	return s.keeper.ClearData(ctx, table, user_id)
}

func (s *Service) DeleteAllLocalFiles() error {

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

func (s *Service) GenerateUUID(ctx context.Context) (string, error) {
	// Генерируем UUID
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	// Возвращаем UUID в виде строки
	return uuid.String(), nil
}
