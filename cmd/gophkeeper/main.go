package main

import (
	"context"
	"fmt"
	"strings"
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
	serverURL := option.ServerURL

	// Remove "https://" and "http://" prefixes
	serverURL = strings.TrimPrefix(serverURL, "https://")
	serverURL = strings.TrimPrefix(serverURL, "http://")

	// Try HTTPS first
	httpsURL := "https://" + serverURL
	sync, clientErr := gksync.NewClientWithResponses(httpsURL, option.CertFilePath, option.KeyFilePath)
	if clientErr != nil {
		// If HTTPS connection failed, try HTTP
		httpURL := "http://" + serverURL
		sync, clientErr = gksync.NewClientWithResponses(httpURL, "", "")
		if clientErr != nil {
			panic(fmt.Sprintf("Failed to create synchronization server: %s", clientErr))
		}
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
	gk.Start(version, buildTime)
}
