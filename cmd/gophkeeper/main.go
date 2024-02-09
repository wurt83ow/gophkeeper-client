package main

import (
	"context"
	"fmt"
	"time"

	"github.com/wurt83ow/gophkeeper-client/pkg/bdkeeper"
	"github.com/wurt83ow/gophkeeper-client/pkg/client"
	"github.com/wurt83ow/gophkeeper-client/pkg/config"
	"github.com/wurt83ow/gophkeeper-client/pkg/encription"
	"github.com/wurt83ow/gophkeeper-client/pkg/gksync"
	"github.com/wurt83ow/gophkeeper-client/pkg/logger"
	"github.com/wurt83ow/gophkeeper-client/pkg/services"
	"github.com/wurt83ow/gophkeeper-client/pkg/syncinfo"
)

func main() {

	// Создайте sync как указатель на nil
	var enc *encription.Enc

	logger := logger.NewLogger()
	option := config.NewConfig(enc)
	keeper := bdkeeper.NewKeeper()
	sm := syncinfo.NewSyncManager()

	// Извлеките данные сессии из файла
	userID, token, sessionStart, err := option.LoadSessionData()
	if err != nil {
		logger.Printf("Не удалось получить сохраненные значения")
	}

	sync, err := gksync.NewClientWithResponses(option.ServerURL)
	if err != nil {
		panic("Не удалось создать сервер синхронизации") //!!!
	}

	enc = encription.NewEnc("password")
	service := services.NewServices(keeper, sync, sm, enc, option, option.SyncWithServer, logger)

	// Create a background context
	ctx := context.Background()

	if err != nil {
		fmt.Printf("Ошибка при синхронизации данных: %s\n", err)
	}
	go func() {
		ticker := time.NewTicker(5 * time.Minute) //!!! Вынести в параметры
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				service.SyncAllWithServer(ctx)
				// Вот здесь получим данные с сервера от других клиентов. Раз в 5 минут будет достаточно!!!

				// Создайте новую информацию о синхронизации
				info := syncinfo.SyncInfo{
					LastSync: time.Now(), // Например, используйте текущее время
				}

				// Обновите и сохраните информацию о синхронизации
				err := sm.UpdateAndSaveSyncInfo(info)
				if err != nil {
					// Обработайте ошибку
					fmt.Println("Ошибка при обновлении и сохранении информации о синхронизации:", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	gk := client.NewClient(ctx, service, enc, option, userID, token, sessionStart)
	defer gk.Close()

	gk.Start()

}
