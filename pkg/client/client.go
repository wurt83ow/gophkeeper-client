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
			gk.addData()
		},
	}
	var cmdGetData = &cobra.Command{
		Use:   "getData",
		Short: "Retrieve data entries",
		Run: func(cmd *cobra.Command, args []string) {
			gk.getData()
		},
	}

	var cmdEdit = &cobra.Command{
		Use:   "edit",
		Short: "Edit entries",
		Run: func(cmd *cobra.Command, args []string) {
			gk.editData()
		},
	}

	var cmdList = &cobra.Command{
		Use:   "ls",
		Short: "List all entries from a chosen table",
		Run: func(cmd *cobra.Command, args []string) {
			gk.list()
		},
	}

	var cmdDelete = &cobra.Command{
		Use:   "rm",
		Short: "Delete an entry from a chosen table",
		Run: func(cmd *cobra.Command, args []string) {
			gk.DeleteData()
		},
	}

	rootCmd.AddCommand(cmdAdd)
	rootCmd.AddCommand(cmdGetData)
	rootCmd.AddCommand(cmdEdit)
	rootCmd.AddCommand(cmdList)
	rootCmd.AddCommand(cmdDelete)
	rootCmd.Execute()
}

func (gk *GophKeeper) Close() {
	gk.rl.Close()
}

func (gk *GophKeeper) getLoginPassword() {
	printMenu()
	line, _ := gk.rl.Readline()
	tableName, valid := getTableNameByChoice(line)
	if !valid {
		fmt.Println("Invalid choice")
		return
	}

	data, _ := gk.storage.GetAllData(tableName, "id", "meta_info")
	if len(data) == 0 {
		fmt.Println("No entries found in the table:", tableName)
		return
	}
	for _, entry := range data {
		fmt.Printf("#%s: %s\n", entry["id"], entry["meta_info"])
	}

	gk.rl.SetPrompt("Enter the ID of the login/password data you want to get: ")
	id, _ := gk.rl.Readline()
	loginPasswordData, err := gk.storage.GetData(user_id, tableName, id)
	if err != nil {
		fmt.Printf("Failed to get data: %s\n", err)
	} else {
		fmt.Printf("Login: %s, Password: %s\n", loginPasswordData["login"], loginPasswordData["password"])
		fmt.Println("Data retrieved successfully!")
	}
}

func (gk *GophKeeper) getData() {
	printMenu()
	line, _ := gk.rl.Readline()
	switch strings.TrimSpace(line) {
	case "1":
		gk.getLoginPassword()
	// case "2":
	// 	gk.getTextData()
	// case "3":
	// 	gk.getBinaryData()
	// case "4":
	// 	gk.getBankCardData()
	default:
		fmt.Println("Invalid choice")
	}
}
func (gk *GophKeeper) addData() {
	printMenu()
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

func (gk *GophKeeper) editData() {
	printMenu()
	line, _ := gk.rl.Readline()
	switch strings.TrimSpace(line) {
	case "1":
		gk.editLoginPassword()
	case "2":
		gk.editTextData()
	case "3":
		gk.editBinaryData()
	case "4":
		gk.editBankCardData()
	default:
		fmt.Println("Invalid choice")
	}
}

func (gk *GophKeeper) editLoginPassword() {
	printMenu()
	line, _ := gk.rl.Readline()
	tableName, valid := getTableNameByChoice(line)
	if !valid {
		fmt.Println("Invalid choice")
		return
	}

	data, _ := gk.storage.GetAllData(tableName, "id", "meta_info")
	if len(data) == 0 {
		fmt.Println("No entries found in the table:", tableName)
		return
	}
	for _, entry := range data {
		fmt.Printf("#%s: %s\n", entry["id"], entry["meta_info"])
	}

	gk.rl.SetPrompt("Enter the ID of the login/password data you want to edit: ")
	id, _ := gk.rl.Readline()
	gk.rl.SetPrompt("Choose a new title (meta-information): ")
	title, _ := gk.rl.Readline()
	gk.rl.SetPrompt("Enter new login: ")
	login, _ := gk.rl.Readline()
	gk.rl.SetPrompt("Enter new password: ")
	gk.rl.Config.EnableMask = true
	password, _ := gk.rl.Readline()
	gk.rl.Config.EnableMask = false
	newData := map[string]string{
		"login":     login,
		"password":  password,
		"meta_info": title,
	}
	err := gk.storage.UpdateData(user_id, id, newData)
	if err != nil {
		fmt.Printf("Failed to edit data: %s\n", err)
	} else {
		fmt.Printf("Login: %s, Password: %s\n", login, password)
		fmt.Println("Data edited successfully!")
	}
}

func (gk *GophKeeper) editTextData() {
	printMenu()
	line, _ := gk.rl.Readline()
	tableName, valid := getTableNameByChoice(line)
	if !valid {
		fmt.Println("Invalid choice")
		return
	}

	data, _ := gk.storage.GetAllData(tableName, "id", "meta_info")
	if len(data) == 0 {
		fmt.Println("No entries found in the table:", tableName)
		return
	}
	for _, entry := range data {
		fmt.Printf("#%s: %s\n", entry["id"], entry["meta_info"])
	}

	gk.rl.SetPrompt("Enter the ID of the text data you want to edit: ")
	id, _ := gk.rl.Readline()
	gk.rl.SetPrompt("Choose a new title (meta-information): ")
	title, _ := gk.rl.Readline()
	gk.rl.SetPrompt("Enter new text data: ")
	text, _ := gk.rl.Readline()
	newData := map[string]string{
		"data":      text,
		"meta_info": title,
	}
	err := gk.storage.UpdateData(user_id, id, newData)
	if err != nil {
		fmt.Printf("Failed to edit data: %s\n", err)
	} else {
		fmt.Printf("Title: %s, Text: %s\n", title, text)
		fmt.Println("Data edited successfully!")
	}
}

func (gk *GophKeeper) editBinaryData() {
	printMenu()
	line, _ := gk.rl.Readline()
	tableName, valid := getTableNameByChoice(line)
	if !valid {
		fmt.Println("Invalid choice")
		return
	}

	data, _ := gk.storage.GetAllData(tableName, "id", "meta_info")
	if len(data) == 0 {
		fmt.Println("No entries found in the table:", tableName)
		return
	}
	for _, entry := range data {
		fmt.Printf("#%s: %s\n", entry["id"], entry["meta_info"])
	}

	gk.rl.SetPrompt("Enter the ID of the binary data you want to edit: ")
	id, _ := gk.rl.Readline()
	gk.rl.SetPrompt("Choose a new title (meta-information): ")
	title, _ := gk.rl.Readline()
	gk.rl.SetPrompt("Specify the new file path: ")
	filePath, _ := gk.rl.Readline()
	newData := map[string]string{
		"path":      filePath,
		"meta_info": title,
	}
	err := gk.storage.UpdateData(user_id, id, newData)
	if err != nil {
		fmt.Printf("Failed to edit data: %s\n", err)
	} else {
		fmt.Printf("Title: %s, File: %s\n", title, filePath)
		fmt.Println("Data edited successfully!")
	}
}

func (gk *GophKeeper) editBankCardData() {
	printMenu()
	line, _ := gk.rl.Readline()
	tableName, valid := getTableNameByChoice(line)
	if !valid {
		fmt.Println("Invalid choice")
		return
	}

	data, _ := gk.storage.GetAllData(tableName, "id", "meta_info")
	if len(data) == 0 {
		fmt.Println("No entries found in the table:", tableName)
		return
	}
	for _, entry := range data {
		fmt.Printf("#%s: %s\n", entry["id"], entry["meta_info"])
	}

	gk.rl.SetPrompt("Enter the ID of the bank card data you want to edit: ")
	id, _ := gk.rl.Readline()
	gk.rl.SetPrompt("Choose a new title (meta-information): ")
	title, _ := gk.rl.Readline()
	gk.rl.SetPrompt("Enter new card number: ")
	cardNumber, _ := gk.rl.Readline()
	gk.rl.SetPrompt("Enter new expiry date (MM/YY): ")
	expiryDate, _ := gk.rl.Readline()
	gk.rl.SetPrompt("Enter new CVV: ")
	cvv, _ := gk.rl.Readline()
	newData := map[string]string{
		"card_number":     cardNumber,
		"expiration_date": expiryDate,
		"cvv":             cvv,
		"meta_info":       title,
	}
	err := gk.storage.UpdateData(user_id, id, newData)
	if err != nil {
		fmt.Printf("Failed to edit data: %s\n", err)
	} else {
		fmt.Printf("Title: %s, Card Number: %s, Expiry Date: %s, CVV: %s\n", title, cardNumber, expiryDate, cvv)
		fmt.Println("Data edited successfully!")
	}
}

func (gk *GophKeeper) list() {
	printMenu()
	line, _ := gk.rl.Readline()
	tableName, valid := getTableNameByChoice(strings.TrimSpace(line))
	if !valid {
		fmt.Println("Invalid choice")
		return
	}
	data, _ := gk.storage.GetAllData(tableName, "id", "meta_info")
	if len(data) == 0 {
		fmt.Println("No entries found in the table:", tableName)
		return
	}
	for _, entry := range data {
		fmt.Printf("#%s: %s\n", entry["id"], entry["meta_info"])
	}
}

func (gk *GophKeeper) DeleteData() {
	printMenu()

	line, _ := gk.rl.Readline()
	tableName, valid := getTableNameByChoice(strings.TrimSpace(line))
	if !valid {
		fmt.Println("Invalid choice")
		return
	}

	data, _ := gk.storage.GetAllData(tableName, "id", "meta_info")
	if len(data) == 0 {
		fmt.Println("No entries found in the table:", tableName)
		return
	}
	for _, entry := range data {
		fmt.Printf("#%s: %s\n", entry["id"], entry["meta_info"])
	}

	fmt.Println("Enter id or meta_info to delete:")
	line, _ = gk.rl.Readline()
	var entriesToDelete []map[string]string
	for _, entry := range data {
		if entry["id"] == line || strings.Contains(entry["meta_info"], line) {
			entriesToDelete = append(entriesToDelete, entry)
		}
	}

	if len(entriesToDelete) > 1 {
		fmt.Println("Multiple entries found. Please enter the id of the entry you want to delete.")
		line, _ = gk.rl.Readline()
		for _, entry := range entriesToDelete {
			if entry["id"] == line {
				fmt.Println("Are you sure you want to delete this entry? (yes/no)")
				line, _ = gk.rl.Readline()
				if strings.ToLower(line) == "yes" {
					gk.storage.DeleteData(user_id, tableName, entry["id"], entry["meta_info"])
					fmt.Println("Entry deleted.")
				}
				return
			}
		}
		fmt.Println("No entry found with the given id.")
	} else if len(entriesToDelete) == 1 {
		fmt.Println("Are you sure you want to delete this entry? (yes/no)")
		line, _ = gk.rl.Readline()
		if strings.ToLower(line) == "yes" {
			gk.storage.DeleteData(user_id, tableName, entriesToDelete[0]["id"], entriesToDelete[0]["meta_info"])
			fmt.Println("Entry deleted.")
		}
	} else {
		fmt.Println("No entry found with the given id or meta_info.")
	}
}

func printMenu() {
	fmt.Println("Choose data type:")
	fmt.Println("1. Login/Password")
	fmt.Println("2. Text data")
	fmt.Println("3. Binary data")
	fmt.Println("4. Bank card data")
}
func getTableNameByChoice(choice string) (string, bool) {
	switch choice {
	case "1":
		return "UserCredentials", true
	case "2":
		return "CreditCardData", true
	case "3":
		return "TextData", true
	case "4":
		return "FilesData", true
	default:
		return "", false
	}
}
