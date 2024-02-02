package main

import (
	"context"

	"github.com/wurt83ow/gophkeeper-client/pkg/bdkeeper"
	"github.com/wurt83ow/gophkeeper-client/pkg/client"
	"github.com/wurt83ow/gophkeeper-client/pkg/config"
	"github.com/wurt83ow/gophkeeper-client/pkg/encription"
	"github.com/wurt83ow/gophkeeper-client/pkg/gksync"
	"github.com/wurt83ow/gophkeeper-client/pkg/services"
)

func main() {
	option := config.NewConfig()
	keeper := bdkeeper.NewKeeper()
	sync := gksync.NewSync(option.ServerURL, option.SyncWithServer)
	sync1, err := gksync.NewClientWithResponses(option.ServerURL)
	if err != nil {
		panic("Не удалось создать сервер синхронизации") //!!!
	}

	enc := encription.NewEnc("password")
	service := services.NewServices(keeper, sync, sync1, enc, option, option.SyncWithServer)

	// Create a background context
	ctx := context.Background()
	gk := client.NewClient(ctx, service, enc, option)
	defer gk.Close()

	gk.Start()

}
