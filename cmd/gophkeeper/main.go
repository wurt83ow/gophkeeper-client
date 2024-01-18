package main

import (
	"github.com/wurt83ow/gophkeeper-client/pkg/bdkeeper"
	"github.com/wurt83ow/gophkeeper-client/pkg/client"
	"github.com/wurt83ow/gophkeeper-client/pkg/storage"
)

func main() {
	keeper := bdkeeper.New()
	storage := storage.New(keeper)
	gk := client.NewGophKeeper(storage)
	defer gk.Close()

	gk.Start()
}
