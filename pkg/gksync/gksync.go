package gksync

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Sync struct {
	serverURL      string
	syncWithServer bool
}

var ErrNetworkUnavailable = errors.New("network unavailable")

func NewSync(serverURL string, syncWithServer bool) *Sync {
	return &Sync{
		serverURL:      serverURL,
		syncWithServer: syncWithServer,
	}
}

func (s *Sync) GetData(user_id int, table string, data map[string]string) error {
	if !s.syncWithServer {
		return nil
	}
	// Преобразовать данные в JSON
	dataJson, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Отправить GET-запрос на сервер
	resp, err := http.Get(fmt.Sprintf("%s/getData/%s/%d?data=%s", s.serverURL, table, user_id, dataJson))
	if err != nil {
		return ErrNetworkUnavailable
	}
	defer resp.Body.Close()

	// Прочитать ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Распарсить JSON-ответ в map[string]string
	var responseData map[string]string
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return err
	}

	// Обновить данные
	for key, value := range responseData {
		data[key] = value
	}

	return nil
}

func (s *Sync) AddData(user_id int, table string, data map[string]string) error {
	if !s.syncWithServer {
		return nil
	}
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = http.Post(fmt.Sprintf("%s/addData/%s/%d", s.serverURL, table, user_id), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return ErrNetworkUnavailable
	}

	return nil
}

func (s *Sync) UpdateData(user_id int, table string, data map[string]string) error {
	if !s.syncWithServer {
		return nil
	}
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/updateData/%s/%d", s.serverURL, table, user_id), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return ErrNetworkUnavailable
	}

	return nil
}

func (s *Sync) DeleteData(user_id int, table string, id string) error {
	if !s.syncWithServer {
		return nil
	}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/deleteData/%s/%d/%s", s.serverURL, table, user_id, id), nil)
	if err != nil {
		return err
	}

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return ErrNetworkUnavailable
	}

	return nil
}

func (s *Sync) GetAllData(user_id int, table string) ([]map[string]string, error) {
	if !s.syncWithServer {
		return nil, nil
	}
	// Отправить GET-запрос на сервер
	resp, err := http.Get(fmt.Sprintf("%s/getAllData/%s/%d", s.serverURL, table, user_id))
	if err != nil {
		return nil, ErrNetworkUnavailable
	}
	defer resp.Body.Close()

	// Прочитать ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Распарсить JSON-ответ в слайс map[string]string
	var data []map[string]string
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *Sync) ClearData(user_id int, table string) error {
	if !s.syncWithServer {
		return nil
	}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/clearData/%s/%d", s.serverURL, table, user_id), nil)
	if err != nil {
		return err
	}

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return ErrNetworkUnavailable
	}

	return nil
}
func (s *Sync) GetPassword(username string) (string, error) {
	if !s.syncWithServer {
		return "", nil
	}
	// Отправить GET-запрос на сервер
	resp, err := http.Get(fmt.Sprintf("%s/getPassword/%s", s.serverURL, username))
	if err != nil {
		return "", ErrNetworkUnavailable
	}
	defer resp.Body.Close()

	// Прочитать ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Распарсить JSON-ответ в строку
	var password string
	err = json.Unmarshal(body, &password)
	if err != nil {
		return "", err
	}

	// Возвращаем хешированный пароль
	return password, nil
}

func (s *Sync) GetUserID(username string) (int, error) {
	if !s.syncWithServer {
		return 0, nil
	}
	// Отправить GET-запрос на сервер
	resp, err := http.Get(fmt.Sprintf("%s/getUserID/%s", s.serverURL, username))
	if err != nil {
		return 0, ErrNetworkUnavailable
	}
	defer resp.Body.Close()

	// Прочитать ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	// Распарсить JSON-ответ в целое число
	var userID int
	err = json.Unmarshal(body, &userID)
	if err != nil {
		return 0, err
	}

	// Возвращаем идентификатор пользователя
	return userID, nil
}

func (s *Sync) SendFile(userID int, data []byte) error {
	if !s.syncWithServer {
		return nil
	}
	resp, err := http.Post(fmt.Sprintf("%s/sendFile/%d", s.serverURL, userID), "application/octet-stream", bytes.NewBuffer(data))
	if err != nil {
		return ErrNetworkUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %s", resp.Status)
	}

	return nil
}
