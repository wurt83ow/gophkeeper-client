package syncinfo

import (
	"os"
	"testing"
	"time"
)

func TestSyncManager(t *testing.T) {
	// Создаем временный файл для тестирования
	tmpfile, err := os.CreateTemp("", "syncinfo-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Создаем новый SyncManager с использованием временного файла
	sm := NewSyncManager(tmpfile.Name())

	// Создаем SyncInfo для тестирования
	testSyncInfo := SyncInfo{
		LastSync: time.Now(),
	}

	// Сохраняем SyncInfo в файл
	err = sm.UpdateAndSaveSyncInfo(testSyncInfo)
	if err != nil {
		t.Fatalf("Failed to update and save sync info: %v", err)
	}

	// Загружаем SyncInfo из файла
	loadedSyncInfo, err := sm.LoadSyncInfoFromFile()
	if err != nil {
		t.Fatalf("Failed to load sync info from file: %v", err)
	}

	// Проверяем, что загруженный SyncInfo соответствует ожидаемому
	if loadedSyncInfo.Format(time.RFC3339) != testSyncInfo.LastSync.Format(time.RFC3339) {
		t.Errorf("Loaded sync info does not match expected value. Expected: %v, Got: %v", testSyncInfo.LastSync, loadedSyncInfo)
	}

	// Проверяем, что данные в файле соответствуют ожидаемому значению с округлением до миллисекунд
	fileContent, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to read file content: %v", err)
	}

	// Преобразуем содержимое файла в формат времени
	loadedTime, err := time.Parse(time.RFC3339Nano, string(fileContent))
	if err != nil {
		t.Fatalf("Failed to parse file content as time: %v", err)
	}

	// Проверяем, что время из файла соответствует времени в SyncInfo (округленному до миллисекунд)
	if loadedTime.Format(time.RFC3339) != testSyncInfo.LastSync.Format(time.RFC3339) {
		t.Errorf("File content does not match expected value. Expected: %v, Got: %v", loadedTime, testSyncInfo.LastSync)
	}
}
