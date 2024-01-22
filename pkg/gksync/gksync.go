package gksync

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"github.com/wurt83ow/gophkeeper-client/pkg/models"
)

func Sync() (models.SyncData, error) {
	db, err := sql.Open("sqlite3", "./client_database.db")
	if err != nil {
		return models.SyncData{}, err
	}
	defer db.Close()

	var syncData models.SyncData

	// UserCredentials
	rows, err := db.Query("SELECT * FROM UserCredentials")
	if err != nil {
		return models.SyncData{}, err
	}
	for rows.Next() {
		var uc models.UserCredentials
		err = rows.Scan(&uc.ID, &uc.UserID, &uc.Login, &uc.Password, &uc.MetaInfo, &uc.UpdatedAt)
		if err != nil {
			return models.SyncData{}, err
		}
		syncData.UserCredentials = append(syncData.UserCredentials, uc)
	}
	rows.Close()

	// TextData
	rows, err = db.Query("SELECT * FROM TextData")
	if err != nil {
		return models.SyncData{}, err
	}
	for rows.Next() {
		var td models.TextData
		err = rows.Scan(&td.ID, &td.UserID, &td.Data, &td.MetaInfo, &td.UpdatedAt)
		if err != nil {
			return models.SyncData{}, err
		}
		syncData.TextData = append(syncData.TextData, td)
	}
	rows.Close()

	// CreditCardData
	rows, err = db.Query("SELECT * FROM CreditCardData")
	if err != nil {
		return models.SyncData{}, err
	}
	for rows.Next() {
		var ccd models.CreditCardData
		err = rows.Scan(&ccd.ID, &ccd.UserID, &ccd.CardNumber, &ccd.ExpirationDate, &ccd.CVV, &ccd.MetaInfo, &ccd.UpdatedAt)
		if err != nil {
			return models.SyncData{}, err
		}
		syncData.CreditCardData = append(syncData.CreditCardData, ccd)
	}
	rows.Close()

	// FilesData
	rows, err = db.Query("SELECT * FROM FilesData")
	if err != nil {
		return models.SyncData{}, err
	}
	for rows.Next() {
		var fd models.FilesData
		err = rows.Scan(&fd.ID, &fd.UserID, &fd.Path, &fd.MetaInfo, &fd.UpdatedAt)
		if err != nil {
			return models.SyncData{}, err
		}
		syncData.FilesData = append(syncData.FilesData, fd)
	}
	rows.Close()

	return syncData, nil
}

func sendToServer(syncData models.SyncData) error {
	data, err := json.Marshal(syncData)
	if err != nil {
		return err
	}

	resp, err := http.Post("http://your-server.com/sync", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server responded with status code %d", resp.StatusCode)
	}

	return nil
}
