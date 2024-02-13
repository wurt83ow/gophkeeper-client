// Package syncinfo provides functions for working with synchronization information.
package syncinfo

import (
	"os"
	"sync"
	"time"
)

// SyncInfo represents data about the last synchronization.
type SyncInfo struct {
	LastSync time.Time // LastSync represents the timestamp of the last synchronization.
}

// SyncManager manages access to and updates of synchronization data.
type SyncManager struct {
	fileMutex sync.RWMutex     // RWMutex to ensure thread safety when working with the file
	syncData  *MutexedSyncInfo // Synchronization data
	filename  string           // File name where synchronization data is stored
}

// MutexedSyncInfo wraps SyncInfo with a mutex for safe access from different threads.
type MutexedSyncInfo struct {
	sync.RWMutex
	SyncInfo SyncInfo // SyncInfo contains synchronization information.
}

// NewSyncManager creates a new SyncManager and initializes a file for storing synchronization data.
func NewSyncManager(fileName string) *SyncManager {

	file, err := os.OpenFile(fileName, os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	file.Close()

	return &SyncManager{
		syncData: &MutexedSyncInfo{},
		filename: fileName,
	}
}

// UpdateSyncInfo updates synchronization data.
func (sm *SyncManager) UpdateSyncInfo(info SyncInfo) {
	sm.syncData.Lock()
	defer sm.syncData.Unlock()
	sm.syncData.SyncInfo = info
}

// GetSyncInfo returns the current synchronization data.
func (sm *SyncManager) GetSyncInfo() SyncInfo {
	sm.syncData.RLock()
	defer sm.syncData.RUnlock()
	return sm.syncData.SyncInfo
}

// SaveSyncInfoToFile saves synchronization data to a file.
func (sm *SyncManager) SaveSyncInfoToFile() error {
	sm.fileMutex.Lock()
	defer sm.fileMutex.Unlock()

	syncInfo := sm.GetSyncInfo()
	lastSyncStr := syncInfo.LastSync.Format(time.RFC3339)

	err := os.WriteFile(sm.filename, []byte(lastSyncStr), 0644)
	return err
}

// LoadSyncInfoFromFile loads synchronization data from a file.
func (sm *SyncManager) LoadSyncInfoFromFile() (time.Time, error) {
	sm.fileMutex.RLock()
	defer sm.fileMutex.RUnlock()

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

// UpdateAndSaveSyncInfo updates and saves synchronization data.
func (sm *SyncManager) UpdateAndSaveSyncInfo(info SyncInfo) error {
	sm.UpdateSyncInfo(info)
	return sm.SaveSyncInfoToFile()
}

// LoadAndUpdateLastSyncFromFile loads the last synchronization time from a file, updates SyncInfo, and returns it.
func (sm *SyncManager) LoadAndUpdateLastSyncFromFile() (time.Time, error) {

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

	// Обновляем SyncInfo с загруженным значением lastSync
	sm.UpdateSyncInfo(SyncInfo{LastSync: lastSync})

	return lastSync, nil
}

// GetTimeWithoutTimeZone возвращает текущее время без информации о временной зоне.
func (sm *SyncManager) GetTimeWithoutTimeZone() time.Time {
	return time.Now().UTC()
}
