package client

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
	"github.com/wurt83ow/gophkeeper-client/pkg/storage"
)

type GophKeeper struct {
	rl      *readline.Instance
	storage *storage.Storage
}

func NewGophKeeper(storage *storage.Storage) *GophKeeper {
	rl, err := readline.New("> ")
	if err != nil {
		log.Fatal(err)
	}
	return &GophKeeper{rl: rl, storage: storage}
}

var user_id int = 1 //Заменить на правильный!!!

func (gk *GophKeeper) Start() {
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

	var cmdList = &cobra.Command{
		Use:   "list",
		Short: "List all entries from a chosen table",
		Run: func(cmd *cobra.Command, args []string) {
			gk.list()
		},
	}

	rootCmd.AddCommand(cmdAdd)
	rootCmd.AddCommand(cmdList)
	rootCmd.Execute()
}

func (gk *GophKeeper) Close() {
	gk.rl.Close()
}

func (gk *GophKeeper) addLoginPassword() {
	for {
		gk.rl.SetPrompt("Choose a title (meta-information): ")
		title, _ := gk.rl.Readline()
		gk.rl.SetPrompt("Enter login: ")
		login, _ := gk.rl.Readline()
		gk.rl.SetPrompt("Enter password: ")
		gk.rl.Config.EnableMask = true
		password, _ := gk.rl.Readline()
		gk.rl.Config.EnableMask = false
		data := map[string]string{
			"login":     login,
			"password":  password,
			"meta_info": title,
		}
		gk.storage.AddData(user_id, "UserCredentials", data)
		fmt.Printf("Login: %s, Password: %s\n", login, password)
		fmt.Println("Data added successfully!")

		gk.rl.SetPrompt("Do you want to continue adding data? (yes/no): ")
		choice, _ := gk.rl.Readline()
		if strings.ToLower(choice) != "yes" && strings.ToLower(choice) != "y" {
			fmt.Println("9999999999999999999999999999999999999999", choice)
			break
		}
	}
}

func (gk *GophKeeper) addTextData() {
	gk.rl.SetPrompt("Choose a title (meta-information): ")
	title, _ := gk.rl.Readline()
	gk.rl.SetPrompt("Enter text data: ")
	text, _ := gk.rl.Readline()
	data := map[string]string{
		"data":      text,
		"meta_info": title,
	}
	gk.storage.AddData(user_id, "TextData", data)
	fmt.Printf("Title: %s, Text: %s\n", title, text)
	fmt.Println("Data added successfully!")
}

func (gk *GophKeeper) list() {
	fmt.Println("Choose data type:")
	fmt.Println("1. Login/Password")
	fmt.Println("2. Text data")
	fmt.Println("3. Binary data")
	fmt.Println("4. Bank card data")

	line, _ := gk.rl.Readline()
	var tableName string
	switch strings.TrimSpace(line) {
	case "1":
		tableName = "UserCredentials"
	case "2":
		tableName = "CreditCardData"
	case "3":
		tableName = "TextData"
	case "4":
		tableName = "FilesData"
	default:
		fmt.Println("Invalid choice")
		return
	}

	data, _ := gk.storage.GetAllData(tableName, "meta_info")
	if len(data) == 0 {
		fmt.Println("No entries found in the table:", tableName)
		return
	}
	for i, entry := range data {
		fmt.Printf("# %d: %s\n", i+1, entry["meta_info"])
	}
}

func (gk *GophKeeper) addBinaryData() {
	gk.rl.SetPrompt("Choose a title (meta-information): ")
	title, _ := gk.rl.Readline()
	gk.rl.SetPrompt("Specify the file path: ")
	filePath, _ := gk.rl.Readline()
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Println("File not found!")
	} else {
		// Read the file
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Failed to read file: %s", err)
		}

		// Encrypt the data
		block, err := aes.NewCipher([]byte("example key 1234")) // Replace with 16, 24, or 32 byte key
		if err != nil {
			log.Fatalf("Failed to create cipher: %s", err)
		}
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			log.Fatalf("Failed to create GCM: %s", err)
		}
		nonce := make([]byte, gcm.NonceSize())
		if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
			log.Fatalf("Failed to create nonce: %s", err)
		}
		ciphertext := gcm.Seal(nonce, nonce, data, nil)

		// Write the encrypted data to a new file
		err = os.WriteFile(title+".enc", ciphertext, 0644)
		if err != nil {
			log.Fatalf("Failed to write file: %s", err)
		}

		fileData := map[string]string{
			"path":      filePath,
			"meta_info": title,
		}
		gk.storage.AddData(user_id, "FilesData", fileData)
		fmt.Printf("Title: %s, File: %s\n", title, filePath)
		fmt.Println("Data added successfully!")
	}
}

func (gk *GophKeeper) addBankCardData() {
	digitsOnly, _ := regexp.Compile(`^\d+$`)
	dateFormat, _ := regexp.Compile(`^\d{2}/\d{2}$`)

	gk.rl.SetPrompt("Choose a title (meta-information): ")
	title, _ := gk.rl.Readline()

	var cardNumber, expiryDate, cvv string

	for {
		gk.rl.SetPrompt("Enter card number: ")
		cardNumber, _ = gk.rl.Readline()
		if !digitsOnly.MatchString(cardNumber) {
			fmt.Println("Card number can only contain digits!")
		} else {
			break
		}
	}

	for {
		gk.rl.SetPrompt("Enter expiry date (MM/YY): ")
		expiryDate, _ = gk.rl.Readline()
		if !dateFormat.MatchString(expiryDate) {
			fmt.Println("Expiry date must be in the format MM/YY!")
		} else {
			break
		}
	}

	for {
		gk.rl.SetPrompt("Enter CVV: ")
		cvv, _ = gk.rl.Readline()
		if !digitsOnly.MatchString(cvv) {
			fmt.Println("CVV can only contain digits!")
		} else {
			break
		}
	}

	cardData := map[string]string{
		"card_number":     cardNumber,
		"expiration_date": expiryDate,
		"cvv":             cvv,
		"meta_info":       title,
	}
	gk.storage.AddData(user_id, "CreditCardData", cardData)
	fmt.Printf("Title: %s, Card Number: %s, Expiry Date: %s, CVV: %s\n", title, cardNumber, expiryDate, cvv)
	fmt.Println("Data added successfully!")
}
