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

type ActionFunc func()
type Client struct {
	rl      *readline.Instance
	service *services.Service
	enc     *encription.Enc
	opt     *config.Options
	ctx     context.Context

	userID       int // добавляем поле для хранения идентификатора текущего пользователя
	token        string
	sessionStart time.Time
}

func NewClient(ctx context.Context, service *services.Service, enc *encription.Enc,
	opt *config.Options, userID int, token string, sessionStart time.Time) *Client {
	rl, err := readline.New("> ")
	if err != nil {
		log.Fatal(err)
	}
	return &Client{rl: rl, ctx: ctx, service: service, enc: enc,
		opt: opt, userID: userID, token: token, sessionStart: sessionStart}
}
func (c *Client) Start() {
	rootCmd := &cobra.Command{
		Use:           "gophkeeper",
		Short:         "GophKeeper is a secure password manager",
		SilenceErrors: true, // Предотвращаем вывод ошибок Cobra
	}

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

	c.ctx = appcontext.WithJWTToken(c.ctx, c.token)
	// Если прошло слишком много времени, попросите пользователя войти в систему снова
	if time.Since(c.sessionStart) > c.opt.SessionDuration {
		fmt.Println("Ваш сеанс истек. Пожалуйста, войдите снова.")
		c.ClearSession()
	}

	for use, runFunc := range commands {
		localRunFunc := runFunc // Создаем локальную переменную
		command := &cobra.Command{
			Use:   use,
			Short: use,
			Run: func(cmd *cobra.Command, args []string) {
				localRunFunc() // Используем локальную переменную
			},
		}
		// Добавляем алиасы для команд 'ls' и 'rm'
		if use == "ls" {
			command.Aliases = []string{"list"}
		} else if use == "rm" {
			command.Aliases = []string{"remove"}
		}
		rootCmd.AddCommand(command)
	}

	err := rootCmd.Execute()
	if err != nil {
		if strings.Contains(err.Error(), "unknown command") {
			fmt.Println("Такой команды не существует. Вот список доступных команд:")
			for cmd := range commands {
				fmt.Println("-", cmd)
			}
		} else {
			fmt.Println("Произошла ошибка:", err)
		}
	}
}

func (c *Client) Close() {
	c.rl.Close()
}

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

func (c *Client) getData() {
	if c.userID == 0 {
		fmt.Println("Пожалуйста, войдите в систему или зарегистрируйтесь.")
		return
	}
	c.chooseAction(c.getLoginPassword, c.getTextData, c.getBinaryData, c.getBankCardData)
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
		return "TextData", true
	case "3":
		return "FilesData", true
	case "4":
		return "CreditCardData", true
	default:
		return "", false
	}
}

func (c *Client) addData() {
	if c.userID == 0 {
		fmt.Println("Пожалуйста, войдите в систему или зарегистрируйтесь.")
		return
	}
	c.chooseAction(c.addLoginPassword, c.addTextData, c.addBinaryData, c.addBankCardData)
}
func (c *Client) editData() {
	if c.userID == 0 {
		fmt.Println("Пожалуйста, войдите в систему или зарегистрируйтесь.")
		return
	}
	c.chooseAction(c.editLoginPassword, c.editTextData, c.editBinaryData, c.editBankCardData)
}

// Реализация функции регистрации
func (c *Client) register() {
	c.rl.SetPrompt("Введите имя пользователя: ")
	username, _ := c.rl.Readline()

	c.rl.SetPrompt("Введите пароль: ")
	c.rl.Config.EnableMask = true
	password, _ := c.rl.Readline()
	c.rl.Config.EnableMask = false
	// Вызовите функцию регистрации в вашем сервисе
	err := c.service.Register(c.ctx, username, password)
	if err != nil {
		fmt.Printf("Ошибка при регистрации: %s\n", err)
	} else {
		fmt.Println("Регистрация прошла успешно!")
	}
}

func (c *Client) login() {
	c.rl.SetPrompt("Введите имя пользователя: ")
	username, _ := c.rl.Readline()

	c.rl.SetPrompt("Введите пароль: ")
	c.rl.Config.EnableMask = true
	password, _ := c.rl.Readline()
	c.rl.Config.EnableMask = false

	// Вызовите функцию входа в систему в вашем сервисе
	userID, token, err := c.service.Login(c.ctx, username, password)
	if err != nil {
		fmt.Printf("Ошибка при входе в систему: %s\n", err)
	} else {
		c.userID = userID
		c.ctx = appcontext.WithJWTToken(c.ctx, token)
		fmt.Println("Вход в систему прошел успешно!")

		// Здесь начинается новая сессия
		if c.opt.SyncWithServer {
			// Извлеките данные сессии из файла
			sessionUserID, _, _, err := c.opt.LoadSessionData()

			//Если не удалось извлечь userID из файла сессии или извлеченный
			//userID отличается от текущего, тогда загрузим все данные с сервера
			if err != nil || c.userID != sessionUserID {
				err = c.service.SyncAllData(c.ctx, userID)
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

		// Получите текущее время и преобразуйте его в строку
		c.sessionStart = time.Now()

		// Запишите userID, token и время начала сеанса в файл
		err = c.saveSessionData(encryptedUserID, token, c.sessionStart.Format(time.RFC3339))
		if err != nil {
			fmt.Printf("Ошибка при записи userID в файл: %s\n", err)
			return
		}
	}
}

func (c *Client) saveSessionData(userID, token, sessionStart string) error {
	// Запишите userID, token и время начала сеанса в файл
	err := os.WriteFile("session.dat", []byte(userID+"\n"+token+"\n"+sessionStart), 0600)
	return err
}

// Метод Logout в модуле client
func (c *Client) Logout() {
	c.ClearSession()
}

func (c *Client) ClearSession() {
	c.userID = 0
	c.ctx = context.Background() // Создаем новый контекст
	os.Remove("session.dat")
}

func (c *Client) list() {
	if c.userID == 0 {
		fmt.Println("Пожалуйста, войдите в систему или зарегистрируйтесь.")
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

		// Предложение сохранить файл
		c.rl.SetPrompt("Do you want to save the file? (yes/no): ")
		choice, _ := c.rl.Readline()

		if strings.ToLower(choice) != "yes" && strings.ToLower(choice) != "y" {
			return
		}

		// Предложение ввести путь для сохранения
		c.rl.SetPrompt("Enter the file name to save the file: ")
		outputFileName, _ := c.rl.Readline()

		// Если в имени файла нет расширения, добавляем его из записи БД
		if filepath.Ext(outputFileName) == "" && newdata["extension"] != "" {
			outputFileName += "." + newdata["extension"]
		}

		// Получение пути к файлу из данных
		fileName := newdata["path"]
		inputPath := filepath.Join(c.opt.FileStoragePath, fileName)

		// Проверим существует ли файл, если нет, то получим его с сервера
		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			c.service.RetrieveFile(c.ctx, c.userID, fileName, inputPath)
		}
		// Расшифровка файла
		err = c.enc.DecryptFile(inputPath, outputFileName)
		if err != nil {
			fmt.Println("Failed to decrypt file:", err)
			return
		}

		fmt.Println("File decrypted and saved successfully!")
	}
	for {
		c.rl.SetPrompt("Do you want to continue geting files? (yes/no): ")
		choice, _ := c.rl.Readline()
		if strings.ToLower(choice) != "yes" && strings.ToLower(choice) != "y" {
			break
		}

		c.getBinaryDataAndSave("FilesData")
	}
}

func (c *Client) printAllData(data []map[string]string) {
	fmt.Println(data)
	for i, entry := range data {
		fmt.Printf("#%d: %s\n", i+1, entry["meta_info"])
	}
}

func (c *Client) printData(data map[string]string) {
	for key, value := range data {
		fmt.Printf("%s: %s\n", key, value)
	}
}

func (c *Client) getLoginPassword() {
	c.getDataAndPrint("UserCredentials", c.printData)
}

func (c *Client) getTextData() {
	c.getDataAndPrint("TextData", c.printData)
}

func (c *Client) getBinaryData() {
	c.getBinaryDataAndSave("FilesData")
}

func (c *Client) getBankCardData() {
	c.getDataAndPrint("CreditCardData", c.printData)
}

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

func (c *Client) addBinaryData() {
	c.rl.SetPrompt("Choose a title (meta-information): ")
	title, _ := c.rl.Readline()
	c.rl.SetPrompt("Specify the file path: ")
	inputPath, _ := c.rl.Readline()

	// Проверяем, существует ли файл
	info, err := os.Stat(inputPath)
	if os.IsNotExist(err) {
		fmt.Println("File not found!")
		return
	}

	// Проверяем размер файла
	if info.Size() > int64(c.opt.MaxFileSize) {
		fmt.Println("File is too large!")
		return
	}

	// Читаем и шифруем файл по частям. Сохраняем зашифрованные данные в файловой системе
	encryptedFilePath, hash, err := c.enc.EncryptFile(inputPath, c.opt.FileStoragePath)
	if err != nil {
		fmt.Printf("Failed to encrypt or write file: %s\n", err)
		return
	}

	// Получаем расширение файла
	extension := filepath.Ext(inputPath)

	// Добавляем метаданные файла в сервис
	fileData := map[string]string{
		"path":      fmt.Sprintf("%x", hash), // Сохраняем хеш вместо пути к файлу
		"meta_info": title,
		"extension": extension, // Сохраняем расширение файла
	}
	err = c.service.AddData(c.ctx, "FilesData", c.userID, fileData)
	if err != nil {
		fmt.Printf("Failed to add data: %s\n", err)
		return
	}

	// Отправляем файл на сервер в отдельной горутине
	go c.service.SyncFile(c.ctx, c.userID, encryptedFilePath, fmt.Sprintf("%x", hash))

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

		// Получение существующих данных
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

func (c *Client) DeleteData() {
	if c.userID == 0 {
		fmt.Println("Пожалуйста, войдите в систему или зарегистрируйтесь.")
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

		fmt.Println("Enter id or meta_info to delete:")
		line, _ = c.rl.Readline()
		var entriesToDelete []map[string]string
		for _, entry := range data {
			if entry["id"] == line || strings.Contains(entry["meta_info"], line) {
				entriesToDelete = append(entriesToDelete, entry)
			}
		}

		if len(entriesToDelete) > 1 {
			fmt.Println("Multiple entries found:")
			for _, entry := range entriesToDelete {
				fmt.Printf("ID: %s, Meta Info: %s\n", entry["id"], entry["meta_info"])
			}
			fmt.Println("Please enter the id of the entry you want to delete.")
			line, _ = c.rl.Readline()
			for _, entry := range entriesToDelete {
				if entry["id"] == line {
					fmt.Println("Are you sure you want to delete this entry? (yes/no)")
					line, _ = c.rl.Readline()
					if strings.ToLower(line) == "yes" {
						err := c.service.DeleteData(c.ctx, tableName, c.userID, entry["id"])
						if err != nil {
							fmt.Printf("Failed to delete data: %s\n", err)
							return
						}
						fmt.Println("Entry deleted.")
					}
					// Переходим к метке
					goto Loop
				}
			}
			fmt.Println("No entry found with the given id.")
		} else if len(entriesToDelete) == 1 {
			fmt.Println("Are you sure you want to delete this entry? (yes/no)")
			line, _ = c.rl.Readline()
			if strings.ToLower(line) == "yes" {
				err := c.service.DeleteData(c.ctx, tableName, c.userID, entriesToDelete[0]["id"])
				if err != nil {
					fmt.Printf("Failed to delete data: %s\n", err)
					return
				}

				fmt.Println("Entry deleted.")
			}
		} else {
			fmt.Println("No entry found with the given id or meta_info.")
		}

		// Метка для перехода
	Loop:
		fmt.Println("Do you want to continue deleting data? (yes/no)")
		line, _ = c.rl.Readline()
		if strings.ToLower(line) != "yes" && strings.ToLower(line) != "y" {
			break
		}
	}
}

func getRowFromUserInput(data []map[string]string, rl *readline.Instance) (map[string]string, error) {
	rl.SetPrompt("Enter the number of the entry you want to get: ")
	strnum, _ := rl.Readline()

	if strnum == "" {
		return nil, errors.New("no input provided")
	}
	num, err := strconv.Atoi(strnum)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при преобразовании номера строки в целое число: %w", err)
	}
	row, err := getStringFromSlice(data, num-1)

	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении строки по её номеру: %w", err)
	}
	return row, nil
}
