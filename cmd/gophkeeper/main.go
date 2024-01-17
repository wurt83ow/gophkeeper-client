package main

import (
	"github.com/wurt83ow/gophkeeper-client/pkg/client"
)

func main() {
	gk := client.NewGophKeeper()
	defer gk.Close()

	gk.Start()
}
