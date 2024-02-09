package config

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Options struct {
	MaxFileSize     int
	FileStoragePath string
	ServerURL       string
	SyncWithServer  bool
	SessionDuration time.Duration
	enc             Encrypt
}

type Encrypt interface {
	Decrypt(encryptedText string) (string, error)
}

func NewConfig(enc Encrypt) *Options {
	maxFileSize := flag.Int("maxFileSize", 100*1024*1024, "maximum file size")
	fileStoragePath := flag.String("fileStoragePath", "", "file storage path")
	serverURL := flag.String("serverURL", "http://localhost:8080", "server URL")
	syncWithServer := flag.Bool("syncWithServer", true, "synchronize with server")

	flag.Parse()

	if *fileStoragePath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		*fileStoragePath = filepath.Join(home, "gkeeper")

		// Создание каталога gkeeper, если он не существует
		if _, err := os.Stat(*fileStoragePath); os.IsNotExist(err) {
			err = os.Mkdir(*fileStoragePath, 0755)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// Check if corresponding environment variables are set and override the values if present.
	if envMaxFileSize, exists := os.LookupEnv("MAX_FILE_SIZE"); exists {
		if value, err := strconv.Atoi(envMaxFileSize); err == nil {
			*maxFileSize = value
		}
	}

	if envFileStoragePath, exists := os.LookupEnv("FILE_STORAGE_PATH"); exists {
		*fileStoragePath = envFileStoragePath
	}

	if envServerURL, exists := os.LookupEnv("SERVER_URL"); exists {
		*serverURL = envServerURL
	}

	if envSyncWithServer, exists := os.LookupEnv("SYNC_WITH_SERVER"); exists {
		if value, err := strconv.ParseBool(envSyncWithServer); err == nil {
			*syncWithServer = value
		}
	}

	return &Options{
		MaxFileSize:     *maxFileSize,
		FileStoragePath: *fileStoragePath,
		ServerURL:       *serverURL,
		SyncWithServer:  *syncWithServer,
		SessionDuration: time.Minute * 300,
		enc:             enc,
	}
}

func (o *Options) LoadSessionData() (int, string, time.Time, error) {
	// Проверьте, существует ли файл
	if _, err := os.Stat("session.dat"); os.IsNotExist(err) {
		return 0, "", time.Time{}, fmt.Errorf("Файл session.dat не существует")
	}

	// Прочитайте файл и разделите его на userID, token, время начала сеанса и время последней синхронизации
	fileContent, err := os.ReadFile("session.dat")
	if err != nil {
		return 0, "", time.Time{}, fmt.Errorf("Ошибка при чтении файла session.dat: %w", err)
	}
	lines := strings.Split(string(fileContent), "\n")
	if len(lines) < 4 {
		return 0, "", time.Time{}, errors.New("Файл session.dat имеет неверный формат")
	}

	// Расшифруйте userID
	decryptedUserID, err := o.enc.Decrypt(lines[0])
	if err != nil {
		return 0, "", time.Time{}, fmt.Errorf("Ошибка при расшифровке userID: %w", err)
	}
	userID, err := strconv.Atoi(decryptedUserID)
	if err != nil {
		return 0, "", time.Time{}, fmt.Errorf("Ошибка при преобразовании userID в целое число: %w", err)
	}

	// Извлеките token
	token := lines[1]

	// Преобразуйте время начала сеанса обратно в Time
	sessionStart, err := time.Parse(time.RFC3339, lines[2])
	if err != nil {
		return 0, "", time.Time{}, fmt.Errorf("Ошибка при разборе времени начала сеанса: %w", err)
	}

	return userID, token, sessionStart, nil
}
