package config

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

// Мок для интерфейса Encrypt
type MockEncrypt struct{}

func (e *MockEncrypt) Decrypt(encryptedText string) (string, error) {
	return encryptedText, nil
}

// Тест проверяет создание новой конфигурации.
func TestNewConfig(t *testing.T) {
	mockEncrypt := &MockEncrypt{}

	config := NewConfig(mockEncrypt)

	// Проверяем, что значения по умолчанию правильно установлены
	expected := &Options{
		MaxFileSize:     100 * 1024 * 1024,
		FileStoragePath: getDefaultFileStoragePath(),
		ServerURL:       "http://localhost:8080",
		SyncWithServer:  true,
		SessionDuration: time.Minute * 300,
		CertFilePath:    "server.crt",
		KeyFilePath:     "server.key",
		SysInfoPath:     "syncinfo.dat",
		enc:             mockEncrypt,
	}

	if !optionsEqual(config, expected) {
		t.Errorf("Expected %+v, got %+v", expected, config)
	}
}

// Тест проверяет загрузку данных сессии из файла.
func TestOptions_LoadSessionData(t *testing.T) {
	// Создаем временный файл с данными сессии
	file, err := os.CreateTemp("", "session.dat")
	if err != nil {
		t.Fatal("Error creating temporary file:", err)
	}
	defer os.Remove(file.Name())

	// Записываем в файл данные сессии
	userID := 123
	token := "test_token"
	sessionStart := time.Now().Format(time.RFC3339)
	_, err = file.WriteString(strconv.Itoa(userID) + "\n" + token + "\n" + sessionStart)
	if err != nil {
		t.Fatal("Error writing to temporary file:", err)
	}

	// Создаем конфигурацию
	mockEncrypt := &MockEncrypt{}
	config := &Options{
		SysInfoPath: file.Name(),
		enc:         mockEncrypt,
	}

	// Загружаем данные сессии
	loadedUserID, loadedToken, loadedSessionStart, err := config.LoadSessionData()
	if err != nil {
		t.Fatal("Error loading session data:", err)
	}

	// Проверяем, что данные сессии правильно загружены
	if loadedUserID != userID || loadedToken != token || loadedSessionStart.Format(time.RFC3339) != sessionStart {
		t.Errorf("Expected userID=%d, token=%s, sessionStart=%s; got userID=%d, token=%s, sessionStart=%s",
			userID, token, sessionStart, loadedUserID, loadedToken, loadedSessionStart.Format(time.RFC3339))
	}
}

// Проверяет равенство двух конфигураций.
func optionsEqual(opt1, opt2 *Options) bool {
	return opt1.MaxFileSize == opt2.MaxFileSize &&
		opt1.FileStoragePath == opt2.FileStoragePath &&
		opt1.ServerURL == opt2.ServerURL &&
		opt1.SyncWithServer == opt2.SyncWithServer &&
		opt1.SessionDuration == opt2.SessionDuration &&
		opt1.CertFilePath == opt2.CertFilePath &&
		opt1.KeyFilePath == opt2.KeyFilePath &&
		opt1.SysInfoPath == opt2.SysInfoPath
}

// Возвращает путь к хранилищу файлов по умолчанию.
func getDefaultFileStoragePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic("Error getting user home directory: " + err.Error())
	}
	return filepath.Join(home, "gkeeper")
}
