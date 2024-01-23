package config

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

type Options struct {
	MaxFileSize     int
	FileStoragePath string
	ServerURL       string
	SyncWithServer  bool
}

func NewConfig() *Options {
	maxFileSize := flag.Int("maxFileSize", 100*1024*1024, "maximum file size")
	fileStoragePath := flag.String("fileStoragePath", "", "file storage path")
	serverURL := flag.String("serverURL", "http://localhost:8080", "server URL")
	syncWithServer := flag.Bool("syncWithServer", false, "synchronize with server")

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
	}
}
