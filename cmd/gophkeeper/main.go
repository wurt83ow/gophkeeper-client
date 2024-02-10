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

	// Initialize encryption as a nil pointer
	var enc *encription.Enc

	// Initialize logger
	logger := logger.NewLogger()

	// Initialize encryption service
	enc = encription.NewEnc("password")
	// Initialize configuration options
	option := config.NewConfig(enc)

	// Initialize database keeper
	keeper := bdkeeper.NewKeeper()

	// Initialize synchronization manager
	sm := syncinfo.NewSyncManager()

	// Extract session data from file
	userID, token, sessionStart, err := option.LoadSessionData()

	if err != nil {
		logger.Printf("Failed to retrieve saved values")
	}

	// Initialize synchronization client
	sync, err := gksync.NewClientWithResponses(option.ServerURL)
	if err != nil {
		panic("Failed to create synchronization server")
	}

	// Initialize application services
	service := services.NewServices(keeper, sync, sm, enc, option, option.SyncWithServer, logger)

	// Create a background context
	ctx := context.Background()

	// Perform initial data synchronization
	if err != nil {
		fmt.Printf("Error synchronizing data: %s\n", err)
	}

	// Start periodic data synchronization with the server
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				service.SyncAllWithServer(ctx)
				err = service.SyncAllData(ctx, userID, true)
				if err != nil {
					fmt.Printf("Error synchronizing data: %s\n", err)
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	// Initialize and start the client
	gk := client.NewClient(ctx, service, enc, option, userID, token, sessionStart)
	defer gk.Close()
	gk.Start()
}
