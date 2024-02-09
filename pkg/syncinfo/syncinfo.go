// Пакет syncinfo предоставляет функции для работы с информацией о синхронизации.
package syncinfo

import (
	"os"
	"sync"
	"time"
)

// SyncInfo представляет данные о последней синхронизации.
type SyncInfo struct {
	LastSync time.Time
}

// SyncManager управляет доступом и обновлением данных о синхронизации.
type SyncManager struct {
	fileMutex sync.Mutex       // Мьютекс для обеспечения потокобезопасности при работе с файлом
	syncData  *MutexedSyncInfo // Данные о синхронизации
	filename  string           // Имя файла, в котором сохраняются данные о синхронизации
}

// MutexedSyncInfo оборачивает SyncInfo вместе с мьютексом для безопасного доступа из разных потоков.
type MutexedSyncInfo struct {
	sync.RWMutex
	SyncInfo SyncInfo
}

// NewSyncManager создает новый SyncManager и инициализирует файл для хранения данных о синхронизации.
func NewSyncManager() *SyncManager {
	filename := "syncinfo.dat"
	file, err := os.OpenFile(filename, os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	file.Close()

	return &SyncManager{
		syncData: &MutexedSyncInfo{},
		filename: filename,
	}
}

// UpdateSyncInfo обновляет данные о синхронизации.
func (sm *SyncManager) UpdateSyncInfo(info SyncInfo) {
	sm.syncData.Lock()
	defer sm.syncData.Unlock()
	sm.syncData.SyncInfo = info
}

// GetSyncInfo возвращает текущие данные о синхронизации.
func (sm *SyncManager) GetSyncInfo() SyncInfo {
	sm.syncData.RLock()
	defer sm.syncData.RUnlock()
	return sm.syncData.SyncInfo
}

// SaveSyncInfoToFile сохраняет данные о синхронизации в файл.
func (sm *SyncManager) SaveSyncInfoToFile() error {
	sm.fileMutex.Lock()
	defer sm.fileMutex.Unlock()

	syncInfo := sm.GetSyncInfo()
	lastSyncStr := syncInfo.LastSync.Format(time.RFC3339)

	err := os.WriteFile(sm.filename, []byte(lastSyncStr), 0644)
	return err
}

// LoadSyncInfoFromFile загружает данные о синхронизации из файла.
func (sm *SyncManager) LoadSyncInfoFromFile() (time.Time, error) {
	sm.fileMutex.Lock()
	defer sm.fileMutex.Unlock()

	fileContent, err := os.ReadFile(sm.filename)
	if err != nil {
		return time.Time{}, err
	}

	lastSync, err := time.Parse(time.RFC3339, string(fileContent))
	if err != nil {
		return time.Time{}, err
	}

	return lastSync, nil
}

func (sm *SyncManager) UpdateAndSaveSyncInfo(info SyncInfo) error {
	sm.UpdateSyncInfo(info)
	return sm.SaveSyncInfoToFile()
}
