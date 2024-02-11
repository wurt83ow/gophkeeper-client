package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/wurt83ow/gophkeeper-client/pkg/bdkeeper"
	"github.com/wurt83ow/gophkeeper-client/pkg/config"
	"github.com/wurt83ow/gophkeeper-client/pkg/encription"
	"github.com/wurt83ow/gophkeeper-client/pkg/gksync"
	"github.com/wurt83ow/gophkeeper-client/pkg/models"
	"github.com/wurt83ow/gophkeeper-client/pkg/syncinfo"
)

// Service provides methods for user registration, login, and data synchronization.
type Service struct {
	keeper         *bdkeeper.Keeper
	sync           *gksync.ClientWithResponses
	sm             *syncinfo.SyncManager
	enc            *encription.Enc
	opt            *config.Options
	syncWithServer bool
	logger         Logger
}

// Logger is an interface for logging messages.
type Logger interface {
	Printf(format string, v ...interface{})
}

// NewServices creates a new instance of the Service.
func NewServices(keeper *bdkeeper.Keeper, sync *gksync.ClientWithResponses, sm *syncinfo.SyncManager, enc *encription.Enc,
	opt *config.Options, syncWithServer bool, logger Logger) *Service {
	return &Service{
		keeper:         keeper,
		sync:           sync,
		sm:             sm,
		enc:            enc,
		opt:            opt,
		syncWithServer: syncWithServer,
		logger:         logger,
	}
}

// Register registers a new user with the provided username and password.
func (s *Service) Register(ctx context.Context, username string, password string) error {
	// Check if the user exists in the database
	userExists, err := s.keeper.UserExists(ctx, username)
	if err != nil {
		return err
	}
	if userExists {
		return errors.New("User already exists")
	}

	if resp, err := s.sync.GetGetUserIDUsername(ctx, username); err == nil {
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			var userID int
			err = json.Unmarshal(body, &userID)
			if err == nil && userID != 0 {
				return errors.New("User already exists")
			}
		}
	}

	// Hash the password
	hashedPassword, err := s.enc.HashPassword(password)
	if err != nil {
		return err
	}

	// Save the new user on the server
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

	// Save the new user in the database
	err = s.keeper.AddUser(ctx, username, hashedPassword)
	if err != nil {
		return err
	}

	return nil
}

// Login authenticates the user with the provided username and password and returns the user ID and token.
func (s *Service) Login(ctx context.Context, username string, password string) (int, string, error) {
	var userID int
	var token string
	var err error

	// Attempt to get the hashed password of the user from the local database
	hashedPassword, err := s.keeper.GetPassword(ctx, username)
	if err != nil {
		// If an error occurs, use the original password
		hashedPassword = password
	} else {
		// Compare the hashed password with the hash of the entered password
		if !s.enc.CompareHashAndPassword(hashedPassword, password) {
			return 0, "", errors.New("Invalid password")
		}
	}

	if s.syncWithServer {
		// If syncWithServer=true, get the userID and jwtToken from the server
		body := gksync.PostLoginJSONRequestBody{
			Username: username,
			Password: hashedPassword,
		}
		resp, err := s.sync.PostLoginWithResponse(ctx, body)
		if err != nil {
			return 0, "", err
		}
		// Check if resp is nil
		if resp.JSON200 == nil {
			return 0, "", fmt.Errorf("Unauthorized")
		}

		userID = *resp.JSON200.UserID
		token = string(*resp.JSON200.Token)
	} else {
		// If syncWithServer=false, only get the userID from keeper
		userID, err = s.keeper.GetUserID(ctx, username)
		if err != nil {
			return 0, "", errors.New("Invalid userId")
		}
	}

	// Return the user ID and token
	return userID, token, nil
}

// SyncFile synchronizes a file with the server.
func (s *Service) SyncFile(ctx context.Context, userID int, filePath string, fileName string) {
	if !s.syncWithServer {
		return
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		s.logger.Printf("Error opening file: %v", err)
		return
	}
	defer file.Close()

	// Send the data to the server
	_, err = s.sync.PostSendFileUserIDWithBody(ctx, userID, fileName, "application/octet-stream", file)
	if err != nil {
		s.logger.Printf("Error sending file to server: %v", err)
	}
}

// SyncAllData synchronizes all data with the server.
func (s *Service) SyncAllData(ctx context.Context, userID int, update bool) error {
	if !s.syncWithServer {
		return nil
	}

	// List of all data tables
	tables := []string{"UserCredentials", "CreditCardData", "TextData", "FilesData"}

	var lastSync time.Time
	if update {

		// Load lastSync from file and update SyncInfo
		var err error
		lastSync, err = s.sm.LoadAndUpdateLastSyncFromFile()

		if err != nil {
			s.logger.Printf("lastSync", err)
			fmt.Printf("Error loading and updating lastSync: %v\n", err)
		}
	}

	// Iterate over each table
	for _, table := range tables {
		// Get all data from the table on the server
		resp, err := s.sync.GetGetAllDataTableUserIDWithResponse(ctx, table, userID, lastSync)
		if err != nil {
			s.logger.Printf("Error getting data from table %s: %v", table, err)
		}

		if resp == nil || resp.JSON200 == nil {
			continue
		}
		if !update {
			// Clear the corresponding table in the local database
			err = s.keeper.ClearData(ctx, table, userID)
			if err != nil {
				s.logger.Printf("Error clearing table %s: %v", table, err)
			}
		}

		// Add all retrieved data to the local database
		for _, row := range *resp.JSON200 {
			if update {
				// Check if the row was deleted
				deleted, _ := strconv.ParseBool(row["deleted"])
				updatedAt, _ := time.Parse(time.RFC3339, row["updated_at"])

				// Get data from the local database
				localData, err := s.keeper.GetData(ctx, table, userID, row["id"])
				if err != nil {
					// If data does not exist, add a new row
					err = s.keeper.AddData(ctx, table, userID, row["id"], row)
					if err != nil {
						s.logger.Printf("Error adding data to table %s: %v", table, err)
					}
				} else {
					localUpdatedAt, _ := time.Parse(time.RFC3339, localData["updated_at"])
					// If data exists and differs, and updatedAt in row is greater than in localData, update it
					if !mapsEqual(localData, row) && !updatedAt.Before(localUpdatedAt) {
						if deleted {
							// If the row was deleted
							err = s.keeper.DeleteData(ctx, table, userID, row["id"])
							if err != nil {
								s.logger.Printf("Error deleting data from table %s: %v", table, err)
							}
						} else {
							err = s.keeper.UpdateData(ctx, table, userID, row["id"], row)
							if err != nil {
								s.logger.Printf("Error updating data in table %s: %v", table, err)
							}
						}
					} else {
						s.logger.Printf("Data in table %s was not updated because the update time is less than or equal to the last synchronization time", table)
					}
				}
			} else {
				err = s.keeper.AddData(ctx, table, userID, row["id"], row)
				if err != nil {
					s.logger.Printf("Error adding data to table %s: %v", table, err)
				}
			}
		}
	}

	// Create new synchronization information
	info := syncinfo.SyncInfo{
		LastSync: time.Now(), // For example, use the current time
	}

	// Update and save synchronization information
	err := s.sm.UpdateAndSaveSyncInfo(info)
	if err != nil {
		// Handle the error
		fmt.Println("Error updating and saving synchronization information:", err)
	}

	return nil
}

// mapsEqual checks if two cards are equal
func mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

// SyncAllWithServer synchronizes all pending data entries with the server.
func (s *Service) SyncAllWithServer(ctx context.Context) {
	// Send data to the server
	// Get all entries from the sync table with status "Pending"
	entries, err := s.keeper.GetPendingSyncEntries(ctx)
	if err != nil {
		return
	}

	for _, entry := range entries {
		err = s.sendData(ctx, entry)

		// If the request is successful, update the entry status to "Done"
		if err == nil {
			err = s.keeper.UpdateSyncEntryStatus(ctx, entry.ID, "Done")
			if err != nil {
				return
			}
		} else {
			// If an error occurs, handle it
			s.handleSyncError(ctx, err, entry)
		}
	}
}

// handleSyncError handles errors that occur during data synchronization with the server.
func (s *Service) handleSyncError(ctx context.Context, err error, entry models.SyncQueue) {
	// Log the error
	s.logger.Printf("Error syncing data: %s\n", err)

	// Retry synchronization
	retryCount := 0
	for retryCount < 3 {
		err = s.sendData(ctx, entry)

		if err == nil {
			// If the request is successful, update the entry status to "Done"
			err = s.keeper.UpdateSyncEntryStatus(ctx, entry.ID, "Done")
			if err != nil {
				s.logger.Printf("Error updating entry status: %s\n", err)
			}
			break
		} else {
			retryCount++
			s.logger.Printf("Error retrying data synchronization: %s\n", err)
		}
	}
}

// sendData synchronizes data with the server.
func (s *Service) sendData(ctx context.Context, entry models.SyncQueue) error {
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

// AddData adds data to the specified table for the user and initiates synchronization if enabled.
func (s *Service) AddData(ctx context.Context, table string, user_id int, data map[string]string) error {
	entry_id, err := s.GenerateUUID(ctx)
	if err != nil {
		return err
	}

	// Encrypt each value in the data before saving it
	encryptedData := make(map[string]string)
	for key, value := range data {
		encryptedValue, err := s.enc.Encrypt(value)
		if err != nil {
			return err
		}
		encryptedData[key] = encryptedValue
	}

	err = s.keeper.AddData(ctx, table, user_id, entry_id, encryptedData)
	if s.syncWithServer && err == nil {
		err = s.keeper.CreateSyncEntry(ctx, "Create", table, user_id, entry_id, encryptedData)
		if err == nil {
			go s.SyncAllWithServer(ctx)
		}
	}
	return err
}

// GetData retrieves data from the specified table for the user.
func (s *Service) GetData(ctx context.Context, table string, user_id int, entry_id string) (map[string]string, error) {
	data, err := s.keeper.GetData(ctx, table, user_id, entry_id)
	if err != nil {
		return nil, err
	}

	// Decrypt the data before returning it
	for key, value := range data {
		decryptedValue, err := s.enc.Decrypt(value)
		if err != nil {
			return nil, err
		}
		data[key] = decryptedValue
	}

	return data, nil
}

// UpdateData updates data in the specified table for the user and initiates synchronization if enabled.
func (s *Service) UpdateData(ctx context.Context, table string, user_id int, entry_id string, data map[string]string) error {
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

// DeleteData deletes data from the specified table for the user and initiates synchronization if enabled.
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

// GetAllData retrieves all data from the specified table for the user and decrypts it before returning.
func (s *Service) GetAllData(ctx context.Context, table string, user_id int, columns ...string) ([]map[string]string, error) {
	data, err := s.keeper.GetAllData(ctx, table, user_id, columns...)
	if err != nil {
		return nil, err
	}

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

	return data, nil
}

// RetrieveFile retrieves a file from the server and saves it locally.
func (s *Service) RetrieveFile(ctx context.Context, user_id int, fileName string, inputPath string) {
	resp, err := s.sync.GetGetFileUserIDEntryID(ctx, user_id, fileName)
	if err != nil {
		s.logger.Printf("Error retrieving file from server: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Printf("Server returned unexpected status: %v", resp.Status)
		return
	}

	out, err := os.Create(inputPath)
	if err != nil {
		s.logger.Printf("Error creating file: %v", err)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		s.logger.Printf("Error saving file: %v", err)
		return
	}

	s.logger.Printf("File successfully retrieved and saved!")
}

// ClearLocalData clears all data from the specified table for the user.
func (s *Service) ClearLocalData(ctx context.Context, table string, user_id int) error {
	return s.keeper.ClearData(ctx, table, user_id)
}

// DeleteAllLocalFiles deletes all files stored locally.
func (s *Service) DeleteAllLocalFiles() error {
	files, err := os.ReadDir(s.opt.FileStoragePath)
	if err != nil {
		return err
	}

	for _, file := range files {
		err = os.Remove(filepath.Join(s.opt.FileStoragePath, file.Name()))
		if err != nil {
			return err
		}
	}

	return nil
}

// GenerateUUID generates a UUID and returns it as a string.
func (s *Service) GenerateUUID(ctx context.Context) (string, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return uuid.String(), nil
}
