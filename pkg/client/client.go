// Package client provides functionalities to interact with GophKeeper, a secure password manager.
package client

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
	"github.com/wurt83ow/gophkeeper-client/pkg/appcontext"
	"github.com/wurt83ow/gophkeeper-client/pkg/config"
	"github.com/wurt83ow/gophkeeper-client/pkg/encription"
	"github.com/wurt83ow/gophkeeper-client/pkg/services"
)

// ActionFunc represents a function that performs an action.
type ActionFunc func()

// Client represents a GophKeeper client.
type Client struct {
	rl           *readline.Instance // Readline instance for user input
	service      *services.Service  // Service for backend operations
	enc          *encription.Enc    // Encryption utility
	opt          *config.Options    // Options for client configuration
	ctx          context.Context    // Context for client operations
	userID       int                // User ID of the current user
	token        string             // Authentication token for the current session
	sessionStart time.Time          // Start time of the current session
}

// NewClient initializes a new GophKeeper client.
func NewClient(ctx context.Context, service *services.Service, enc *encription.Enc,
	opt *config.Options, userID int, token string, sessionStart time.Time) *Client {
	rl, err := readline.New("> ")
	if err != nil {
		log.Fatal(err)
	}
	return &Client{rl: rl, ctx: ctx, service: service, enc: enc,
		opt: opt, userID: userID, token: token, sessionStart: sessionStart}
}

// Start starts the GophKeeper client.
func (c *Client) Start(version, buildTime string) {
	// Root command for Cobra CLI
	rootCmd := &cobra.Command{
		Use:           "gophkeeper",
		Short:         "GophKeeper is a secure password manager",
		SilenceErrors: true, // Prevent Cobra from printing errors
	}

	// Map of command names to their corresponding functions
	commands := map[string]func(){
		"register": c.register,
		"login":    c.login,
		"logout":   c.Logout,

		"add":  c.addData,
		"edit": c.editData,
		"ls":   c.list,
		"rm":   c.DeleteData,
		"get":  c.getData,
	}

	// Set JWT token in the context
	c.ctx = appcontext.WithJWTToken(c.ctx, c.token)

	// If session has expired, prompt the user to log in again
	if time.Since(c.sessionStart) > c.opt.SessionDuration {
		fmt.Println("Your session has expired. Please log in again.")
		// c.ClearSession()
	}

	// Add commands to the root command
	for use, runFunc := range commands {
		localRunFunc := runFunc // Create a local variable
		command := &cobra.Command{
			Use:   use,
			Short: use,
			Run: func(cmd *cobra.Command, args []string) {
				localRunFunc() // Use the local variable
			},
		}
		// Add aliases for commands 'ls' and 'rm'
		if use == "ls" {
			command.Aliases = []string{"list"}
		} else if use == "rm" {
			command.Aliases = []string{"remove"}
		}
		rootCmd.AddCommand(command)
	}

	// Execute the root command
	err := rootCmd.Execute()
	if err != nil {
		if strings.Contains(err.Error(), "unknown command") {
			// Вывод информации о версии и времени сборки
			fmt.Printf("Version: %s\nBuild Time: %s\n", version, buildTime)
			fmt.Println("Command not found. Here is the list of available commands:")
			for cmd := range commands {
				fmt.Println("-", cmd)
			}
		} else {
			fmt.Println("Error:", err)
		}
	}
}

// Close closes the GophKeeper client.
func (c *Client) Close() {
	// Получаем записи с статусом "Progress"
	entries, err := c.service.GetSyncEntriesByStatus(context.Background(), "Progress")
	if err != nil {
		// Обработка ошибки
		fmt.Println("Error retrieving sync entries:", err)
		return
	}

	// Если есть записи со статусом "Progress", запрашиваем у пользователя продолжение или прерывание синхронизации
	if len(entries) > 0 {
		fmt.Println("There are pending sync entries. Do you want to continue syncing data before closing? (yes/no)")

		// Читаем ответ пользователя
		choice, _ := c.rl.Readline()
		if strings.ToLower(choice) == "yes" || strings.ToLower(choice) == "y" {
			// Продолжаем синхронизацию
			return
		}
	}

	// Закрываем ридлайнер и завершаем выполнение программы
	c.rl.Close()
}

// chooseAction prompts the user to choose an action and executes the corresponding function.
func (c *Client) chooseAction(func1, func2, func3, func4 ActionFunc) {
	printMenu()
	line, _ := c.rl.Readline()
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

// getData prompts the user to choose the type of data to retrieve.
func (c *Client) getData() {
	if c.userID == 0 {
		fmt.Println("Please log in or register.")
		return
	}
	c.chooseAction(c.getLoginPassword, c.getTextData, c.getBinaryData, c.getBankCardData)
}

// printMenu prints the menu for data type selection.
func printMenu() {
	fmt.Println("Choose data type:")
	fmt.Println("1. Login/Password")
	fmt.Println("2. Text data")
	fmt.Println("3. Binary data")
	fmt.Println("4. Bank card data")
}

// getTableNameByChoice returns the table name based on the user's choice.
func getTableNameByChoice(choice string) (string, bool) {
	switch choice {
	case "1":
		return "UserCredentials", true
	case "2":
		return "TextData", true
	case "3":
		return "FilesData", true
	case "4":
		return "CreditCardData", true
	default:
		return "", false
	}
}

// addData prompts the user to choose the type of data to add.
func (c *Client) addData() {
	if c.userID == 0 {
		fmt.Println("Please log in or register.")
		return
	}
	c.chooseAction(c.addLoginPassword, c.addTextData, c.addBinaryData, c.addBankCardData)
}

// editData prompts the user to choose the type of data to edit.
func (c *Client) editData() {
	if c.userID == 0 {
		fmt.Println("Please log in or register.")
		return
	}
	c.chooseAction(c.editLoginPassword, c.editTextData, c.editBinaryData, c.editBankCardData)
}

// register implements the registration function.
func (c *Client) register() {
	c.rl.SetPrompt("Enter username: ")
	username, _ := c.rl.Readline()

	c.rl.SetPrompt("Enter password: ")
	c.rl.Config.EnableMask = true
	password, _ := c.rl.Readline()
	c.rl.Config.EnableMask = false

	// Call the registration function in the service
	err := c.service.Register(c.ctx, username, password)
	if err != nil {
		fmt.Printf("Registration failed: %s\n", err)
	} else {
		fmt.Println("Registration successful!")
	}
}

// login prompts the user to enter their username and password, attempts to log them into the system, and initializes a new session if successful.
func (c *Client) login() {
	c.rl.SetPrompt("Введите имя пользователя: ")
	username, _ := c.rl.Readline()

	c.rl.SetPrompt("Введите пароль: ")
	c.rl.Config.EnableMask = true
	password, _ := c.rl.Readline()
	c.rl.Config.EnableMask = false

	// Call the login function in your service
	userID, token, err := c.service.Login(c.ctx, username, password)
	if err != nil {
		fmt.Printf("Ошибка при входе в систему: %s\n", err)
	} else {
		c.userID = userID
		c.ctx = appcontext.WithJWTToken(c.ctx, token)
		fmt.Println("Вход в систему прошел успешно!")

		// A new session starts here
		if c.opt.SyncWithServer {
			// Extract session data from the file
			sessionUserID, _, _, err := c.opt.LoadSessionData()

			// If unable to extract userID from the session file or the extracted
			// userID differs from the current one, then load all data from the server
			if err != nil || c.userID != sessionUserID {
				err = c.service.SyncAllData(c.ctx, userID, false) // update=false
				if err != nil {
					fmt.Printf("Ошибка при синхронизации данных: %s\n", err)
				}
			}
		}

		encryptedUserID, err := c.enc.Encrypt(strconv.Itoa(c.userID))
		if err != nil {
			fmt.Printf("Ошибка при шифровании userID: %s\n", err)
			return
		}

		// Get the current time and convert it to a string
		c.sessionStart = time.Now()

		// Write userID, token, and session start time to the file
		err = c.saveSessionData(encryptedUserID, token, c.sessionStart.Format(time.RFC3339))
		if err != nil {
			fmt.Printf("Ошибка при записи userID в файл: %s\n", err)
			return
		}
	}
}

// saveSessionData saves session data (userID, token, session start time) to a file.
func (c *Client) saveSessionData(userID, token, sessionStart string) error {
	err := os.WriteFile("session.dat", []byte(userID+"\n"+token+"\n"+sessionStart), 0600)
	return err
}

// Logout terminates the current session by clearing session data.
func (c *Client) Logout() {
	c.ClearSession()
}

// ClearSession clears the current session by resetting user ID and context, and removing session data file.
func (c *Client) ClearSession() {
	c.userID = 0
	c.ctx = context.Background() // Create a new context
	os.Remove("session.dat")
}

// list displays a list of entries of a specified data type for the current user.
func (c *Client) list() {
	if c.userID == 0 {
		fmt.Println("Please log in or register.")
		return
	}
	printMenu()
	line, _ := c.rl.Readline()
	tableName, valid := getTableNameByChoice(strings.TrimSpace(line))
	if !valid {
		fmt.Println("Invalid choice")
		return
	}
	data, _ := c.service.GetAllData(c.ctx, tableName, c.userID, "id", "meta_info")
	if len(data) == 0 {
		fmt.Println("No entries found in the table:", tableName)
		return
	}
	c.printAllData(data)
}

// getDataAndPrint retrieves and prints data entries of a specified data type for the current user.
func (c *Client) getDataAndPrint(tableName string, printFunc func(data map[string]string)) {
	for {
		data, _ := c.service.GetAllData(c.ctx, tableName, c.userID, "id", "meta_info")

		if len(data) == 0 {
			fmt.Println("No entries found in the table:", tableName)
			return
		}
		c.printAllData(data)

		if tableName == "" {
			return
		}
		row, err := getRowFromUserInput(data, c.rl)
		if err != nil {
			fmt.Println(err)
			return
		}

		newdata, err := c.service.GetData(c.ctx, tableName, c.userID, row["id"])
		if err != nil {
			fmt.Printf("Failed to get data: %s\n", err)
		} else {
			printFunc(newdata)
			fmt.Println("Data retrieved successfully!")
		}

		c.rl.SetPrompt("Do you want to continue getting data? (yes/no): ")
		choice, _ := c.rl.Readline()
		if strings.ToLower(choice) != "yes" && strings.ToLower(choice) != "y" {
			break
		}
	}
}

func getStringFromSlice(data []map[string]string, index int) (map[string]string, error) {
	if index < 0 || index >= len(data) {
		return nil, errors.New("index out of range")
	}
	return data[index], nil
}

// getBinaryDataAndSave retrieves binary data entries of a specified data type for the current user and saves them as files.
func (c *Client) getBinaryDataAndSave(tableName string) {
	data, _ := c.service.GetAllData(c.ctx, tableName, c.userID, "id", "meta_info")
	if len(data) == 0 {
		fmt.Println("No entries found in the table:", tableName)
		return
	}

	c.printAllData(data)
	row, err := getRowFromUserInput(data, c.rl)
	if err != nil {
		fmt.Println(err)
		return
	}

	newdata, err := c.service.GetData(c.ctx, tableName, c.userID, row["id"])
	if err != nil {
		fmt.Printf("Failed to get data: %s\n", err)
	} else {
		c.printData(newdata)
		fmt.Println("Data retrieved successfully!")

		// Prompt to save the file
		c.rl.SetPrompt("Do you want to save the file? (yes/no): ")
		choice, _ := c.rl.Readline()

		if strings.ToLower(choice) != "yes" && strings.ToLower(choice) != "y" {
			return
		}

		// Prompt for the path to save the file
		c.rl.SetPrompt("Enter the file name to save the file: ")
		outputFileName, _ := c.rl.Readline()

		// If the file name has no extension, add it from the database record
		if filepath.Ext(outputFileName) == "" && newdata["extension"] != "" {
			outputFileName += "." + newdata["extension"]
		}

		// Get the file path from the data
		fileName := newdata["path"]
		inputPath := filepath.Join(c.opt.FileStoragePath, fileName)

		// Check if the file exists, if not, retrieve it from the server
		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			c.service.RetrieveFile(c.ctx, c.userID, fileName, inputPath)
		}
		// Decrypt the file
		err = c.enc.DecryptFile(inputPath, outputFileName)
		if err != nil {
			fmt.Println("Failed to decrypt file:", err)
			return
		}

		fmt.Println("File decrypted and saved successfully!")
	}
	for {
		c.rl.SetPrompt("Do you want to continue getting files? (yes/no): ")
		choice, _ := c.rl.Readline()
		if strings.ToLower(choice) != "yes" && strings.ToLower(choice) != "y" {
			break
		}

		c.getBinaryDataAndSave("FilesData")
	}
}

// printAllData prints all entries in a given data set.
func (c *Client) printAllData(data []map[string]string) {
	fmt.Println(data)
	for i, entry := range data {
		fmt.Printf("#%d: %s\n", i+1, entry["meta_info"])
	}
}

// printData prints the given data to the standard output.
func (c *Client) printData(data map[string]string) {
	for key, value := range data {
		fmt.Printf("%s: %s\n", key, value)
	}
}

// getLoginPassword fetches and prints user login credentials.
func (c *Client) getLoginPassword() {
	c.getDataAndPrint("UserCredentials", c.printData)
}

// getTextData fetches and prints text data.
func (c *Client) getTextData() {
	c.getDataAndPrint("TextData", c.printData)
}

// getBinaryData fetches binary data.
func (c *Client) getBinaryData() {
	c.getBinaryDataAndSave("FilesData")
}

// getBankCardData fetches and prints bank card data.
func (c *Client) getBankCardData() {
	c.getDataAndPrint("CreditCardData", c.printData)
}

// addDataRepeatedly adds data repeatedly until the user decides to stop.
// It prompts the user to add data of a specified type, then adds it to the service,
// and prints the added data. It continues prompting the user until they choose to stop.
func (c *Client) addDataRepeatedly(dataType string, dataFunc func() map[string]string, printFunc func(data map[string]string)) {
	for {
		data := dataFunc()
		err := c.service.AddData(c.ctx, dataType, c.userID, data)
		if err != nil {
			fmt.Printf("Failed to add data: %s\n", err)
			return
		}
		printFunc(data)
		fmt.Println("Data added successfully!")
		c.rl.SetPrompt("Do you want to continue adding data? (yes/no): ")
		choice, _ := c.rl.Readline()
		if strings.ToLower(choice) != "yes" && strings.ToLower(choice) != "y" {
			break
		}
	}
}

// addLoginPassword prompts the user to add login credentials repeatedly.
func (c *Client) addLoginPassword() {
	c.addDataRepeatedly("UserCredentials", func() map[string]string {
		c.rl.SetPrompt("Choose a title (meta-information): ")
		title, _ := c.rl.Readline()
		c.rl.SetPrompt("Enter login: ")
		login, _ := c.rl.Readline()
		c.rl.SetPrompt("Enter password: ")
		c.rl.Config.EnableMask = true
		password, _ := c.rl.Readline()
		c.rl.Config.EnableMask = false
		return map[string]string{
			"login":     login,
			"password":  password,
			"meta_info": title,
		}
	}, func(data map[string]string) {
		fmt.Printf("Login: %s", data["login"])
	})
}

// addTextData prompts the user to add text data repeatedly.
func (c *Client) addTextData() {
	c.addDataRepeatedly("TextData", func() map[string]string {
		c.rl.SetPrompt("Choose a title (meta-information): ")
		title, _ := c.rl.Readline()
		c.rl.SetPrompt("Enter text data: ")
		text, _ := c.rl.Readline()
		return map[string]string{
			"data":      text,
			"meta_info": title,
		}
	}, func(data map[string]string) {
		fmt.Printf("Title: %s, Text: %s\n", data["meta_info"], data["data"])
	})
}

// addBinaryData adds binary data.
// It prompts the user to choose a title and specify the file path.
// Then, it checks if the file exists and if its size is within the allowed limit.
// If the file is valid, it encrypts the file, retrieves its extension, and adds metadata to the file service.
// Finally, it sends the encrypted file to the server in a separate goroutine.
func (c *Client) addBinaryData() {
	c.rl.SetPrompt("Choose a title (meta-information): ")
	title, _ := c.rl.Readline()
	c.rl.SetPrompt("Specify the file path: ")
	inputPath, _ := c.rl.Readline()

	// Check if the file exists
	info, err := os.Stat(inputPath)
	if os.IsNotExist(err) {
		fmt.Println("File not found!")
		return
	}

	// Check the file size
	if info.Size() > int64(c.opt.MaxFileSize) {
		fmt.Println("File is too large!")
		return
	}

	// Read and encrypt the file in parts. Save the encrypted data to the file system.
	_, hash, err := c.enc.EncryptFile(inputPath, c.opt.FileStoragePath)
	if err != nil {
		fmt.Printf("Failed to encrypt or write file: %s\n", err)
		return
	}

	// Get the file extension
	extension := filepath.Ext(inputPath)

	// Add file metadata to the service
	fileData := map[string]string{
		"path":      fmt.Sprintf("%x", hash), // Save the hash instead of the file path
		"meta_info": title,
		"extension": extension, // Save the file extension
	}
	err = c.service.AddData(c.ctx, "FilesData", c.userID, fileData)
	if err != nil {
		fmt.Printf("Failed to add data: %s\n", err)
		return
	}

	// // Send the file to the server in a separate goroutine
	// go c.service.SyncFile(c.ctx, c.userID, encryptedFilePath, fmt.Sprintf("%x", hash))

	fmt.Printf("Title: %s, File: %s\n", title, inputPath)
	fmt.Println("Data added successfully!")

	for {
		c.rl.SetPrompt("Do you want to continue adding files? (yes/no): ")
		choice, _ := c.rl.Readline()
		if strings.ToLower(choice) != "yes" && strings.ToLower(choice) != "y" {
			break
		}

		c.addBinaryData()
	}
}

// addBankCardData prompts the user to add bank card data repeatedly.
func (c *Client) addBankCardData() {
	c.addDataRepeatedly("CreditCardData", func() map[string]string {
		digitsOnly, _ := regexp.Compile(`^\d+$`)
		dateFormat, _ := regexp.Compile(`^\d{2}/\d{2}$`)
		c.rl.SetPrompt("Choose a title (meta-information): ")
		title, _ := c.rl.Readline()
		var cardNumber, expiryDate, cvv string
		for {
			c.rl.SetPrompt("Enter card number: ")
			cardNumber, _ = c.rl.Readline()
			if !digitsOnly.MatchString(cardNumber) {
				fmt.Println("Card number can only contain digits!")
			} else {
				break
			}
		}
		for {
			c.rl.SetPrompt("Enter expiry date (MM/YY): ")
			expiryDate, _ = c.rl.Readline()
			if !dateFormat.MatchString(expiryDate) {
				fmt.Println("Expiry date must be in the format MM/YY!")
			} else {
				break
			}
		}
		for {
			c.rl.SetPrompt("Enter CVV: ")
			c.rl.Config.EnableMask = true
			cvv, _ = c.rl.Readline()
			c.rl.Config.EnableMask = false
			if !digitsOnly.MatchString(cvv) {
				fmt.Println("CVV can only contain digits!")
			} else {
				break
			}
		}
		return map[string]string{
			"card_number":     cardNumber,
			"expiration_date": expiryDate,
			"cvv":             cvv,
			"meta_info":       title,
		}
	}, func(data map[string]string) {
		fmt.Printf("Title: %s, Card Number: %s, Expiry Date: %s\n", data["meta_info"], data["card_number"], data["expiration_date"])
	})
}

// editAllData allows the user to edit data in a specified table.
// It retrieves existing data from the table, prompts the user to select a row to edit,
// then prompts the user to enter new data, and updates the existing data in the table.
// It continues editing data until the user decides to stop.
func (c *Client) editAllData(tableName string, getDataFunc func(oldData map[string]string) map[string]string) {
	for {
		data, _ := c.service.GetAllData(c.ctx, tableName, c.userID, "id", "meta_info")

		if len(data) == 0 {
			fmt.Println("No entries found in the table:", tableName)
			return
		}

		c.printAllData(data)
		row, err := getRowFromUserInput(data, c.rl)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Get existing data
		oldData, err := c.service.GetData(c.ctx, tableName, c.userID, row["id"])

		if err != nil {
			fmt.Printf("Failed to get data: %s\n", err)
			return
		}
		newData := getDataFunc(oldData)
		err = c.service.UpdateData(c.ctx, tableName, c.userID, row["id"], newData)
		if err != nil {
			fmt.Printf("Failed to edit data: %s\n", err)
		} else {
			fmt.Println("Data edited successfully!")
		}
		c.rl.SetPrompt("Do you want to continue editing data? (yes/no): ")
		choice, _ := c.rl.Readline()
		if strings.ToLower(choice) != "yes" && strings.ToLower(choice) != "y" {
			break
		}
	}
}

// editLoginPassword allows the user to edit login credentials.
// It retrieves existing login credentials, prompts the user to enter new data,
// and updates the existing login credentials in the "UserCredentials" table.
func (c *Client) editLoginPassword() {
	c.editAllData("UserCredentials", func(oldData map[string]string) map[string]string {
		c.rl.SetPrompt(fmt.Sprintf("Choose a new title (meta-information) [%s]: ", oldData["meta_info"]))
		title, _ := c.rl.Readline()
		if title == "" {
			title = oldData["meta_info"]
		}
		c.rl.SetPrompt(fmt.Sprintf("Enter new login [%s]: ", oldData["login"]))
		login, _ := c.rl.Readline()
		if login == "" {
			login = oldData["login"]
		}
		c.rl.SetPrompt("Enter new password: ")
		c.rl.Config.EnableMask = true
		password, _ := c.rl.Readline()
		c.rl.Config.EnableMask = false
		if password == "" {
			password = oldData["password"]
		}
		return map[string]string{
			"login":     login,
			"password":  password,
			"meta_info": title,
		}
	})
}

// editTextData allows the user to edit text data.
// It retrieves existing text data, prompts the user to enter new data,
// and updates the existing text data in the "TextData" table.
func (c *Client) editTextData() {
	c.editAllData("TextData", func(oldData map[string]string) map[string]string {
		c.rl.SetPrompt(fmt.Sprintf("Choose a new title (meta-information) [%s]: ", oldData["meta_info"]))
		title, _ := c.rl.Readline()
		if title == "" {
			title = oldData["meta_info"]
		}
		c.rl.SetPrompt(fmt.Sprintf("Enter new text data [%s]: ", oldData["data"]))
		text, _ := c.rl.Readline()
		if text == "" {
			text = oldData["data"]
		}
		return map[string]string{
			"data":      text,
			"meta_info": title,
		}
	})
}

// editBinaryData allows the user to edit binary data.
// It retrieves existing binary data, prompts the user to enter new data,
// and updates the existing binary data in the "FilesData" table.
func (c *Client) editBinaryData() {
	c.editAllData("FilesData", func(oldData map[string]string) map[string]string {
		c.rl.SetPrompt(fmt.Sprintf("Choose a new title (meta-information) [%s]: ", oldData["meta_info"]))
		title, _ := c.rl.Readline()
		if title == "" {
			title = oldData["meta_info"]
		}
		return map[string]string{
			"meta_info": title,
		}
	})
}

// editBankCardData allows the user to edit bank card data.
// It retrieves existing bank card data, prompts the user to enter new data,
// and updates the existing bank card data in the "CreditCardData" table.
func (c *Client) editBankCardData() {
	c.editAllData("CreditCardData", func(oldData map[string]string) map[string]string {
		c.rl.SetPrompt(fmt.Sprintf("Choose a new title (meta-information) [%s]: ", oldData["meta_info"]))
		title, _ := c.rl.Readline()
		if title == "" {
			title = oldData["meta_info"]
		}
		c.rl.SetPrompt(fmt.Sprintf("Enter new card number [%s]: ", oldData["card_number"]))
		cardNumber, _ := c.rl.Readline()
		if cardNumber == "" {
			cardNumber = oldData["card_number"]
		}
		c.rl.SetPrompt(fmt.Sprintf("Enter new expiry date (MM/YY) [%s]: ", oldData["expiration_date"]))
		expiryDate, _ := c.rl.Readline()
		if expiryDate == "" {
			expiryDate = oldData["expiration_date"]
		}
		c.rl.SetPrompt(fmt.Sprintf("Enter new CVV [%s]: ", oldData["cvv"]))
		cvv, _ := c.rl.Readline()
		if cvv == "" {
			cvv = oldData["cvv"]
		}
		return map[string]string{
			"card_number":     cardNumber,
			"expiration_date": expiryDate,
			"cvv":             cvv,
			"meta_info":       title,
		}
	})
}

// DeleteData allows the user to delete data from a specified table.
// It prompts the user to choose a table, then prompts for the entry to delete,
// and deletes the selected entry from the table. It continues this process until the user decides to stop.
func (c *Client) DeleteData() {
	if c.userID == 0 {
		fmt.Println("Please log in or register.")
		return
	}
	printMenu()
	line, _ := c.rl.Readline()
	tableName, valid := getTableNameByChoice(strings.TrimSpace(line))
	if !valid {
		fmt.Println("Invalid choice")
		return
	}

	for {
		data, _ := c.service.GetAllData(c.ctx, tableName, c.userID, "id", "meta_info")
		if len(data) == 0 {
			fmt.Println("No entries found in the table:", tableName)
			return
		}
		c.printAllData(data)

		fmt.Println("Enter the number of the entry to delete:")
		line, _ := c.rl.Readline()
		num, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil || num < 1 || num > len(data) {
			fmt.Println("Invalid input. Please enter a valid number.")
			continue
		}

		entryToDelete := data[num-1]
		fmt.Printf("Are you sure you want to delete the following entry?\n%s\n", formatEntry(num, entryToDelete))
		fmt.Println("Type 'yes' to confirm or 'no' to cancel.")
		line, _ = c.rl.Readline()
		if strings.ToLower(strings.TrimSpace(line)) != "yes" {
			fmt.Println("Deletion canceled.")
			continue
		}

		err = c.service.DeleteData(c.ctx, tableName, c.userID, entryToDelete["id"])
		if err != nil {
			fmt.Printf("Failed to delete data: %s\n", err)
			return
		}
		fmt.Println("Entry deleted.")

		fmt.Println("Do you want to continue deleting data? (yes/no)")
		line, _ = c.rl.Readline()
		if strings.ToLower(strings.TrimSpace(line)) != "yes" {
			break
		}
	}
}

// formatEntry takes a record number and a map of data rows and formats it into a string for output.
func formatEntry(num int, entry map[string]string) string {
	var formattedEntry strings.Builder
	formattedEntry.WriteString("Entry:\n")
	formattedEntry.WriteString(fmt.Sprintf("#: %d\n", num))
	for key, value := range entry {
		if key != "id" { // Skip the id output
			formattedEntry.WriteString(fmt.Sprintf("%s: %s\n", key, value))
		}
	}
	return formattedEntry.String()
}

// getRowFromUserInput prompts the user to enter the number of the entry they want to retrieve and returns the corresponding row of data.
func getRowFromUserInput(data []map[string]string, rl *readline.Instance) (map[string]string, error) {
	rl.SetPrompt("Enter the number of the entry you want to get: ")
	strnum, _ := rl.Readline()

	if strnum == "" {
		return nil, errors.New("no input provided")
	}
	num, err := strconv.Atoi(strnum)
	if err != nil {
		return nil, fmt.Errorf("failed to convert the row number to an integer: %w", err)
	}
	row, err := getStringFromSlice(data, num-1)

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve the row by its number: %w", err)
	}
	return row, nil
}
