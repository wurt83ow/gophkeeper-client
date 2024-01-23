package main

import (
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
	enc := encription.NewEnc("password")
	service := services.NewServices(keeper, sync, enc, option)

	gk := client.NewClient(service, enc, option)
	defer gk.Close()

	gk.Start()
}
