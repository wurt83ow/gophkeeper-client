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

type ActionFunc func()
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

var user_id int = 1 //!!! Заменить на правильный!!!

func (gk *GophKeeper) Start() {
	rootCmd := &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper is a secure password manager",
	}

	commands := map[string]func(){
		"add":  gk.addData,
		"edit": gk.editData,
		"ls":   gk.list,
		"rm":   gk.DeleteData,
		"get":  gk.getData,
	}

	for use, runFunc := range commands {
		runFunc := runFunc //fix error
		command := &cobra.Command{
			Use:   use,
			Short: use,
			Run:   func(cmd *cobra.Command, args []string) { runFunc() },
		}
		rootCmd.AddCommand(command)
	}

	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}

func (gk *GophKeeper) Close() {
	gk.rl.Close()
}

func (gk *GophKeeper) chooseAction(func1, func2, func3, func4 ActionFunc) {
	printMenu()
	line, _ := gk.rl.Readline()
	switch strings.TrimSpace(line) {
	case "1":
		func1()
	case "2":
		func2()
	case "3":
		func3()
	case "4":
		func4()
	default:
		fmt.Println("Invalid choice")
	}
}

func (gk *GophKeeper) getData() {
	gk.chooseAction(gk.getLoginPassword, gk.getTextData, gk.getBinaryData, gk.getBankCardData)
}

func (gk *GophKeeper) addData() {
	gk.chooseAction(gk.addLoginPassword, gk.addTextData, gk.addBinaryData, gk.addBankCardData)
}

func (gk *GophKeeper) editData() {
	gk.chooseAction(gk.editLoginPassword, gk.editTextData, gk.editBinaryData, gk.editBankCardData)
}
func (gk *GophKeeper) selectData() (string, string) {
	printMenu()
	line, _ := gk.rl.Readline()
	tableName, valid := getTableNameByChoice(line)
	if !valid {
		fmt.Println("Invalid choice")
		return "", ""
	}

	data, _ := gk.storage.GetAllData(tableName, "id", "meta_info")
	if len(data) == 0 {
		fmt.Println("No entries found in the table:", tableName)
		return "", ""
	}
	for _, entry := range data {
		fmt.Printf("#%s: %s\n", entry["id"], entry["meta_info"])
	}

	gk.rl.SetPrompt("Enter the ID of the binary data you want to get: ")
	id, _ := gk.rl.Readline()

	return tableName, id
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

func (gk *GophKeeper) getLoginPassword() {
	tableName, id := gk.selectData()
	if tableName == "" || id == "" {
		return
	}
	loginPasswordData, err := gk.storage.GetData(user_id, tableName, id)
	if err != nil {
		fmt.Printf("Failed to get data: %s\n", err)
	} else {
		fmt.Printf("Login: %s, Password: %s\n", loginPasswordData["login"], loginPasswordData["password"])
		fmt.Println("Data retrieved successfully!")
	}
}

func (gk *GophKeeper) getTextData() {
	tableName, id := gk.selectData()
	if tableName == "" || id == "" {
		return
	}
	textData, err := gk.storage.GetData(user_id, tableName, id)
	if err != nil {
		fmt.Printf("Failed to get data: %s\n", err)
	} else {
		fmt.Printf("Title: %s, Text: %s\n", textData["meta_info"], textData["data"])
		fmt.Println("Data retrieved successfully!")
	}
}

func (gk *GophKeeper) getBinaryData() {
	tableName, id := gk.selectData()
	if tableName == "" || id == "" {
		return
	}

	binaryData, err := gk.storage.GetData(user_id, tableName, id)
	if err != nil {
		fmt.Printf("Failed to get data: %s\n", err)
	} else {
		fmt.Printf("Title: %s, File: %s\n", binaryData["meta_info"], binaryData["path"])
		fmt.Println("Data retrieved successfully!")
	}
}

func (gk *GophKeeper) getBankCardData() {
	tableName, id := gk.selectData()
	if tableName == "" || id == "" {
		return
	}
	bankCardData, err := gk.storage.GetData(user_id, tableName, id)
	if err != nil {
		fmt.Printf("Failed to get data: %s\n", err)
	} else {
		fmt.Printf("Title: %s, Card Number: %s, Expiry Date: %s, CVV: %s\n", bankCardData["meta_info"], bankCardData["card_number"], bankCardData["expiration_date"], bankCardData["cvv"])
		fmt.Println("Data retrieved successfully!")
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
		err := gk.storage.AddData(user_id, "UserCredentials", data)
		if err != nil {
			fmt.Printf("Failed to add data: %s\n", err)
			return
		}

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

	err := gk.storage.AddData(user_id, "TextData", data)
	if err != nil {
		fmt.Printf("Failed to add data: %s\n", err)
		return
	}

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
		block, err := aes.NewCipher([]byte("example key 1234")) // !!!Replace with 16, 24, or 32 byte key
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

		err = gk.storage.AddData(user_id, "FilesData", fileData)
		if err != nil {
			fmt.Printf("Failed to add data: %s\n", err)
			return
		}
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

	err := gk.storage.AddData(user_id, "CreditCardData", cardData)
	if err != nil {
		fmt.Printf("Failed to add data: %s\n", err)
		return
	}
	fmt.Printf("Title: %s, Card Number: %s, Expiry Date: %s, CVV: %s\n", title, cardNumber, expiryDate, cvv)
	fmt.Println("Data added successfully!")
}

func (gk *GophKeeper) editLoginPassword() {
	tableName, id := gk.selectData()
	if tableName == "" || id == "" {
		return
	}
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
	tableName, id := gk.selectData()
	if tableName == "" || id == "" {
		return
	}
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
	tableName, id := gk.selectData()
	if tableName == "" || id == "" {
		return
	}
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
	tableName, id := gk.selectData()
	if tableName == "" || id == "" {
		return
	}
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

					err := gk.storage.DeleteData(user_id, tableName, entry["id"], entry["meta_info"])
					if err != nil {
						fmt.Printf("Failed to delete data: %s\n", err)
						return
					}
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
			err := gk.storage.DeleteData(user_id, tableName, entriesToDelete[0]["id"], entriesToDelete[0]["meta_info"])
			if err != nil {
				fmt.Printf("Failed to delete data: %s\n", err)
				return
			}

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
