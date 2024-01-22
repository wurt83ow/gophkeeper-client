package main

import (
	"log"

	"github.com/wurt83ow/gophkeeper-client/pkg/bdkeeper"
	"github.com/wurt83ow/gophkeeper-client/pkg/client"
	"github.com/wurt83ow/gophkeeper-client/pkg/gksync"
	"github.com/wurt83ow/gophkeeper-client/pkg/services"
)

func main() {
	keeper := bdkeeper.NewKeeper()
	sync := gksync.NewSync("http://localhost:8080")
	service := services.NewServices(keeper, sync)
	// Проверка, является ли это первым запуском
	isEmpty, err := keeper.IsEmpty()
	if err != nil {
		log.Fatalf("Ошибка при проверке пустоты хранилища: %v", err)
	}
	if isEmpty {
		err = service.InitSync( /* параметры */ )
		if err != nil {
			log.Fatalf("Ошибка при инициализации синхронизации: %v", err)
		}
	}

	gk := client.NewClient(service)
	defer gk.Close()

	gk.Start()
}
