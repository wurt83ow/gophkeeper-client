package main

import (
	"context"
	"fmt"

	"github.com/wurt83ow/gophkeeper-client/pkg/bdkeeper"
	"github.com/wurt83ow/gophkeeper-client/pkg/client"
	"github.com/wurt83ow/gophkeeper-client/pkg/config"
	"github.com/wurt83ow/gophkeeper-client/pkg/encription"
	"github.com/wurt83ow/gophkeeper-client/pkg/gksync"
	"github.com/wurt83ow/gophkeeper-client/pkg/logger"
	"github.com/wurt83ow/gophkeeper-client/pkg/services"
	"github.com/wurt83ow/gophkeeper-client/pkg/syncinfo"
)

var (
	version   string
	buildTime string
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
	keeper := bdkeeper.NewKeeper(nil)

	// Initialize synchronization manager
	sm := syncinfo.NewSyncManager(option.SysInfoPath)

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

	// Initialize and start the client
	gk := client.NewClient(ctx, service, enc, option, userID, token, sessionStart, sm.GetTimeWithoutTimeZone)
	defer gk.Close()
	gk.Start(version, buildTime)

}
