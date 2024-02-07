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
)

func main() {

	logger := logger.NewLogger()
	option := config.NewConfig()
	keeper := bdkeeper.NewKeeper()

	sync, err := gksync.NewClientWithResponses(option.ServerURL)
	if err != nil {
		panic("Не удалось создать сервер синхронизации") //!!!
	}

	enc := encription.NewEnc("password")
	service := services.NewServices(keeper, sync, enc, option, option.SyncWithServer, logger)

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
			case <-ctx.Done():
				return
			}
		}
	}()
	gk := client.NewClient(ctx, service, enc, option)
	defer gk.Close()

	gk.Start()

}
