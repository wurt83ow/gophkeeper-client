package client

import (
	"context"
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

	userID int // добавляем поле для хранения идентификатора текущего пользователя
}

func NewClient(ctx context.Context, service *services.Service, enc *encription.Enc, opt *config.Options) *Client {
	rl, err := readline.New("> ")
	if err != nil {
		log.Fatal(err)
	}
	return &Client{rl: rl, ctx: ctx, service: service, enc: enc, opt: opt}
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

	if _, err := os.Stat("session.dat"); err == nil {
		// Прочитайте файл и разделите его на userID и время начала сеанса
		fileContent, err := os.ReadFile("session.dat")
		if err != nil {
			fmt.Println("Ошибка при чтении файла session.dat:", err)
			return
		}
		lines := strings.Split(string(fileContent), "\n")
		if len(lines) < 2 {
			fmt.Println("Файл session.dat имеет неверный формат")
			return
		}

		// Расшифруйте userID
		decryptedUserID, err := c.enc.Decrypt(lines[0])
		if err != nil {
			fmt.Println("Ошибка при расшифровке userID:", err)
			return
		}
		userID, err := strconv.Atoi(decryptedUserID)
		if err != nil {
			fmt.Println("Ошибка при преобразовании userID в целое число:", err)
			return
		}
		c.userID = userID

		// Преобразуйте время начала сеанса обратно в Time
		sessionStart, err := time.Parse(time.RFC3339, lines[1])
		if err != nil {
			fmt.Println("Ошибка при разборе времени начала сеанса:", err)
			return
		}

		// Если прошло слишком много времени, попросите пользователя войти в систему снова
		if time.Since(sessionStart) > c.opt.SessionDuration {
			fmt.Println("Ваш сеанс истек. Пожалуйста, войдите снова.")
			c.userID = 0
			os.Remove("session.dat")
		}
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
	fmt.Println("Get data:")
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
	userID, err := c.service.Login(c.ctx, username, password)
	if err != nil {
		fmt.Printf("Ошибка при входе в систему: %s\n", err)
	} else {
		c.userID = userID
		fmt.Println("Вход в систему прошел успешно!")

		// Здесь начинается новая сессия
		if c.opt.SyncWithServer {
			err = c.service.SyncAllData(c.ctx, userID)
			if err != nil {
				fmt.Printf("Ошибка при синхронизации данных: %s\n", err)
			}
		}

		encryptedUserID, err := c.enc.Encrypt(strconv.Itoa(c.userID))
		if err != nil {
			fmt.Printf("Ошибка при шифровании userID: %s\n", err)
			return
		}

		// Получите текущее время и преобразуйте его в строку
		sessionStart := time.Now().Format(time.RFC3339)

		// Запишите userID и время начала сеанса в файл
		err = os.WriteFile("session.dat", []byte(encryptedUserID+"\n"+sessionStart), 0600)
		if err != nil {
			fmt.Printf("Ошибка при записи userID в файл: %s\n", err)
			return
		}

	}
}

// Метод Logout в модуле client
func (c *Client) Logout() {
	c.userID = 0
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
	data, _ := c.service.GetAllData(c.ctx, c.userID, tableName, "id", "meta_info")
	if len(data) == 0 {
		fmt.Println("No entries found in the table:", tableName)
		return
	}
	for _, entry := range data {
		fmt.Printf("#%s: %s\n", entry["id"], entry["meta_info"])
	}
}

func (c *Client) getDataAndPrint(tableName string, printFunc func(data map[string]string)) {
	for {
		data, _ := c.service.GetAllData(c.ctx, c.userID, tableName, "id", "meta_info")
		if len(data) == 0 {
			fmt.Println("No entries found in the table:", tableName)
			return
		}

		for _, entry := range data {
			fmt.Printf("#%s: %s\n", entry["id"], entry["meta_info"])
		}

		c.rl.SetPrompt("Enter the ID of the entry you want to get: ")
		strid, _ := c.rl.Readline()

		if tableName == "" || strid == "" {
			return
		}
		id, err := strconv.Atoi(strid)
		if err != nil {
			fmt.Println("Ошибка при преобразовании ID в целое число:", err)
			return
		}
		newdata, err := c.service.GetData(c.ctx, c.userID, tableName, id)
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
func (c *Client) getBinaryDataAndSave(tableName string) {

	data, _ := c.service.GetAllData(c.ctx, c.userID, tableName, "id", "meta_info")
	if len(data) == 0 {
		fmt.Println("No entries found in the table:", tableName)
		return
	}

	for _, entry := range data {
		fmt.Printf("#%s: %s\n", entry["id"], entry["meta_info"])
	}

	c.rl.SetPrompt("Enter the ID of the entry you want to get: ")
	strid, _ := c.rl.Readline()

	if strid == "" {
		return
	}
	id, err := strconv.Atoi(strid)
	if err != nil {
		fmt.Println("Ошибка при преобразовании ID в целое число:", err)
		return
	}
	newdata, err := c.service.GetData(c.ctx, c.userID, tableName, id)
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
		c.rl.SetPrompt("Enter the path to save the file: ")
		outputPath, _ := c.rl.Readline()

		// Получение пути к файлу из данных
		fileName := newdata["path"]
		inputPath := filepath.Join(c.opt.FileStoragePath, fileName)

		// Расшифровка файла
		err = c.enc.DecryptFile(inputPath, outputPath)
		if err != nil {
			fmt.Println("Failed to decrypt file:", err)
			return
		}

		fmt.Println("File decrypted and saved successfully!")

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
		err := c.service.AddData(c.ctx, c.userID, dataType, data)
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

	// Добавляем метаданные файла в сервис
	fileData := map[string]string{
		"path":      fmt.Sprintf("%x", hash), // Сохраняем хеш вместо пути к файлу
		"meta_info": title,
	}
	err = c.service.AddData(c.ctx, c.userID, "FilesData", fileData)
	if err != nil {
		fmt.Printf("Failed to add data: %s\n", err)
		return
	}

	// Отправляем файл на сервер
	err = c.service.SyncFile(c.userID, encryptedFilePath)
	if err != nil {
		fmt.Printf("Failed to sync file: %s\n", err)
		return
	}

	fmt.Printf("Title: %s, File: %s\n", title, inputPath)
	fmt.Println("Data added successfully!")
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
		data, _ := c.service.GetAllData(c.ctx, c.userID, tableName, "id", "meta_info")

		if len(data) == 0 {
			fmt.Println("No entries found in the table:", tableName)
			return
		}

		for _, entry := range data {
			fmt.Printf("#%s: %s\n", entry["id"], entry["meta_info"])
		}

		c.rl.SetPrompt("Enter the ID of the entry you want to get: ")
		strid, _ := c.rl.Readline()

		if strid == "" {
			return
		}
		id, err := strconv.Atoi(strid)

		if err != nil {
			fmt.Println("Ошибка при преобразовании ID в целое число:", err)
			return
		}
		// Получение существующих данных
		oldData, err := c.service.GetData(c.ctx, c.userID, tableName, id)

		if err != nil {
			fmt.Printf("Failed to get data: %s\n", err)
			return
		}
		newData := getDataFunc(oldData)
		err = c.service.UpdateData(c.ctx, c.userID, id, tableName, newData)
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

		data, _ := c.service.GetAllData(c.ctx, c.userID, tableName, "id", "meta_info")
		if len(data) == 0 {
			fmt.Println("No entries found in the table:", tableName)
			return
		}
		for _, entry := range data {
			fmt.Printf("#%s: %s\n", entry["id"], entry["meta_info"])
		}

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
						err := c.service.DeleteData(c.ctx, c.userID, tableName, entry["id"])
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
				err := c.service.DeleteData(c.ctx, c.userID, tableName, entriesToDelete[0]["id"])
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
