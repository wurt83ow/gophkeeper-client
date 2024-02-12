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

// Options represents the configuration options for the application.
type Options struct {
	MaxFileSize     int           // MaxFileSize represents the maximum allowed size for files.
	FileStoragePath string        // FileStoragePath represents the path where files are stored.
	ServerURL       string        // ServerURL represents the URL of the server.
	SyncWithServer  bool          // SyncWithServer determines whether to synchronize data with the server.
	SessionDuration time.Duration // SessionDuration represents the duration of a session.
	CertFilePath    string        // CertFilePath represents the path to the certificate file.
	KeyFilePath     string        // KeyFilePath represents the path to the key file.
	SysInfoPath     string        // SysInfoPath represents the path where synchronization data is stored.
	enc             Encrypt       // enc is an instance implementing the Encrypt interface for encryption operations.
}

// Encrypt is an interface for encryption operations.
type Encrypt interface {
	Decrypt(encryptedText string) (string, error)
}

// NewConfig creates a new instance of Options with the provided encryption implementation.
func NewConfig(enc Encrypt) *Options {
	maxFileSize := flag.Int("maxFileSize", 100*1024*1024, "maximum file size")
	fileStoragePath := flag.String("fileStoragePath", "", "file storage path")
	serverURL := flag.String("serverURL", "http://localhost:8080", "server URL")
	syncWithServer := flag.Bool("syncWithServer", true, "synchronize with server")
	certFilePath := flag.String("certFilePath", "server.crt", "certificate file path")
	keyFilePath := flag.String("keyFilePath", "server.key", "key file path")
	sysInfoPath := flag.String("sysInfoPath", "syncinfo.dat", "synchronization data file path")

	flag.Parse()

	if *fileStoragePath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		*fileStoragePath = filepath.Join(home, "gkeeper")

		// Create the gkeeper directory if it doesn't exist
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
		CertFilePath:    *certFilePath,
		KeyFilePath:     *keyFilePath,
		SysInfoPath:     *sysInfoPath,
		enc:             enc,
	}
}

// LoadSessionData loads session data from the session.dat file.
func (o *Options) LoadSessionData() (int, string, time.Time, error) {
	// Check if the file exists
	if _, err := os.Stat(o.SysInfoPath); os.IsNotExist(err) {
		return 0, "", time.Time{}, fmt.Errorf("%s file does not exist", o.SysInfoPath)
	}

	// Read the file and split it into lines
	fileContent, err := os.ReadFile(o.SysInfoPath)
	if err != nil {
		return 0, "", time.Time{}, fmt.Errorf("error reading %s file: %w", o.SysInfoPath, err)
	}
	lines := strings.Split(string(fileContent), "\n")

	// Check if the file contains at least 3 lines
	if len(lines) < 3 {
		return 0, "", time.Time{}, errors.New(o.SysInfoPath + " file has an invalid format")
	}

	// Decrypt the userID
	decryptedUserID, err := o.enc.Decrypt(lines[0])
	if err != nil {
		return 0, "", time.Time{}, fmt.Errorf("error decrypting userID: %w", err)
	}
	userID, err := strconv.Atoi(decryptedUserID)
	if err != nil {
		return 0, "", time.Time{}, fmt.Errorf("error converting userID to integer: %w", err)
	}

	// Extract the token
	token := lines[1]

	// Parse the session start time
	sessionStart, err := time.Parse(time.RFC3339, lines[2])
	if err != nil {
		return 0, "", time.Time{}, fmt.Errorf("error parsing session start time: %w", err)
	}

	return userID, token, sessionStart, nil
}
