package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wurt83ow/gophkeeper-client/client"
)

func main() {
	gk := client.NewGophKeeper()
	defer gk.Close()

	rootCmd := &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper is a secure password manager",
	}
	var cmdAdd = &cobra.Command{
		Use:   "add",
		Short: "Add a new entry",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Choose data type:")
			fmt.Println("1. Login/Password")
			fmt.Println("2. Text data")
			fmt.Println("3. Binary data")
			fmt.Println("4. Bank card data")

			line, _ := gk.rl.Readline()
			switch strings.TrimSpace(line) {
			case "1":
				gk.addLoginPassword()
			case "2":
				gk.addTextData()
			case "3":
				gk.addBinaryData()
			case "4":
				gk.addBankCardData()
			default:
				fmt.Println("Invalid choice")
			}
		},
	}

	rootCmd.AddCommand(cmdAdd)
	rootCmd.Execute()
}
