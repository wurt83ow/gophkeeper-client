package main

import (
	"github.com/wurt83ow/gophkeeper-client/pkg/bdkeeper"
	"github.com/wurt83ow/gophkeeper-client/pkg/client"
	"github.com/wurt83ow/gophkeeper-client/pkg/gksync"
	"github.com/wurt83ow/gophkeeper-client/pkg/services"
)

func main() {
	keeper := bdkeeper.NewKeeper()
	sync := gksync.NewSync("http://localhost:8080")
	service := services.NewServices(keeper, sync)
	gk := client.NewClient(service)
	defer gk.Close()

	gk.Start()
}
