package client

import (
	"crypto/sha256"
	"fmt"
	"io"
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

	userID int // добавляем поле для хранения идентификатора текущего пользователя
}

func NewClient(service *services.Service, enc *encription.Enc, opt *config.Options) *Client {
	rl, err := readline.New("> ")
	if err != nil {
		log.Fatal(err)
	}
	return &Client{rl: rl, service: service, enc: enc, opt: opt}
}
func (c *Client) Start() {
	rootCmd := &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper is a secure password manager",
	}

	commands := map[string]func(){
		"add":  c.addData,
		"edit": c.editData,
		"ls":   c.list,
		"rm":   c.DeleteData,

		"register": c.register,
		"login":    c.login,
		"logout":   c.Logout,
		"get":      c.getData,
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
		rootCmd.AddCommand(command)
	}

	err := rootCmd.Execute()
	if err != nil {
		panic(err)
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

func (c *Client) selectData() (string, string) {
	// printMenu()
	line, _ := c.rl.Readline()
	tableName, valid := getTableNameByChoice(line)
	if !valid {
		fmt.Println("Invalid choice")
		return "", ""
	}

	data, _ := c.service.GetAllData(c.userID, tableName, "id", "meta_info")
	if len(data) == 0 {
		fmt.Println("No entries found in the table:", tableName)
		return "", ""
	}
	for _, entry := range data {
		fmt.Printf("#%s: %s\n", entry["id"], entry["meta_info"])
	}

	c.rl.SetPrompt("Enter the ID of the binary data you want to get: ")
	id, _ := c.rl.Readline()

	return tableName, id
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
		fmt.Println("Пожалуйста, войдите в систему или зарегистрируйтесь.")
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
	err := c.service.Register(username, password)
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
	userID, err := c.service.Login(username, password)
	if err != nil {
		fmt.Printf("Ошибка при входе в систему: %s\n", err)
	} else {
		c.userID = userID
		fmt.Println("Вход в систему прошел успешно!")

		// Здесь начинается новая сессия
		if c.opt.SyncWithServer {
			err = c.service.SyncAllData(userID)
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
	data, _ := c.service.GetAllData(c.userID, tableName, "id", "meta_info")
	if len(data) == 0 {
		fmt.Println("No entries found in the table:", tableName)
		return
	}
	for _, entry := range data {
		fmt.Printf("#%s: %s\n", entry["id"], entry["meta_info"])
	}
}
func (c *Client) getDataAndPrint(tableName string, printFunc func(data map[string]string)) {
	data, _ := c.service.GetAllData(c.userID, tableName, "id", "meta_info")
	if len(data) == 0 {
		fmt.Println("No entries found in the table:", tableName)
		return
	}

	for _, entry := range data {
		fmt.Printf("#%s: %s\n", entry["id"], entry["meta_info"])
	}

	c.rl.SetPrompt("Enter the ID of the binary data you want to get: ")
	id, _ := c.rl.Readline()

	if tableName == "" || id == "" {
		return
	}
	strid, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println("Ошибка при преобразовании userID в целое число:", err)
		return
	}
	newdata, err := c.service.GetData(c.userID, tableName, strid)
	if err != nil {
		fmt.Printf("Failed to get data: %s\n", err)
	} else {
		printFunc(newdata)
		fmt.Println("Data retrieved successfully!")
	}
}

func (c *Client) getLoginPassword() {
	c.getDataAndPrint("UserCredentials", func(data map[string]string) {
		fmt.Printf("Login: %s, Password: %s\n", data["login"], data["password"])
	})
}

func (c *Client) getTextData() {
	c.getDataAndPrint("TextData", func(data map[string]string) {
		fmt.Printf("Title: %s, Text: %s\n", data["meta_info"], data["data"])
	})
}

func (c *Client) getBinaryData() {
	c.getDataAndPrint("FilesData", func(data map[string]string) {
		fmt.Printf("Title: %s, File: %s\n", data["meta_info"], data["path"])
	})
}

func (c *Client) getBankCardData() {
	c.getDataAndPrint("CreditCardData", func(data map[string]string) {
		fmt.Printf("Title: %s, Card Number: %s, Expiry Date: %s, CVV: %s\n",
			data["meta_info"], data["card_number"], data["expiration_date"], data["cvv"])
	})
}

func (c *Client) addDataRepeatedly(dataType string, dataFunc func() map[string]string, printFunc func(data map[string]string)) {
	for {
		data := dataFunc()
		err := c.service.AddData(c.userID, dataType, data)
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
	filePath, _ := c.rl.Readline()

	// Проверяем, существует ли файл
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		fmt.Println("File not found!")
		return
	}

	// Проверяем размер файла
	if info.Size() > int64(c.opt.MaxFileSize) {
		fmt.Println("File is too large!")
		return
	}

	// Открываем файл
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Failed to open file: %s\n", err)
		return
	}
	defer file.Close()

	// Читаем и шифруем файл по частям
	var encryptedData []byte
	hasher := sha256.New()
	buf := make([]byte, 1024)
	for {
		n, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("Failed to read file: %s\n", err)
			return
		}

		data := buf[:n]
		encryptedPart, err := c.enc.EncryptData(data)
		if err != nil {
			fmt.Printf("Failed to encrypt file: %s\n", err)
			return
		}

		// Добавляем зашифрованную часть в encryptedData
		encryptedData = append(encryptedData, encryptedPart...)

		// Обновляем хеш данных
		hasher.Write(encryptedPart)
	}

	// Получаем окончательный хеш данных
	hash := hasher.Sum(nil)

	// Сохраняем зашифрованные данные в файловой системе
	encryptedFilePath := filepath.Join(c.opt.FileStoragePath, fmt.Sprintf("%x", hash))
	err = os.WriteFile(encryptedFilePath, encryptedData, 0644)
	if err != nil {
		fmt.Printf("Failed to write file: %s\n", err)
		return
	}

	// Добавляем метаданные файла в сервис
	fileData := map[string]string{
		"path":      fmt.Sprintf("%x", hash), // Сохраняем хеш вместо пути к файлу
		"meta_info": title,
	}
	err = c.service.AddData(c.userID, "FilesData", fileData)
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

	fmt.Printf("Title: %s, File: %s\n", title, filePath)
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

func (c *Client) editLoginPassword() {
	fmt.Println("editLoginPassword")
	tableName, id := c.selectData()
	if tableName == "" || id == "" {
		return
	}
	c.rl.SetPrompt("Choose a new title (meta-information): ")
	title, _ := c.rl.Readline()
	c.rl.SetPrompt("Enter new login: ")
	login, _ := c.rl.Readline()
	c.rl.SetPrompt("Enter new password: ")
	c.rl.Config.EnableMask = true
	password, _ := c.rl.Readline()
	c.rl.Config.EnableMask = false
	newData := map[string]string{
		"login":     login,
		"password":  password,
		"meta_info": title,
	}
	err := c.service.UpdateData(c.userID, id, newData)
	if err != nil {
		fmt.Printf("Failed to edit data: %s\n", err)
	} else {
		fmt.Printf("Login: %s, Password: %s\n", login, password)
		fmt.Println("Data edited successfully!")
	}
}
func (c *Client) editTextData() {
	tableName, id := c.selectData()
	if tableName == "" || id == "" {
		return
	}
	c.rl.SetPrompt("Choose a new title (meta-information): ")
	title, _ := c.rl.Readline()
	c.rl.SetPrompt("Enter new text data: ")
	text, _ := c.rl.Readline()
	newData := map[string]string{
		"data":      text,
		"meta_info": title,
	}
	err := c.service.UpdateData(c.userID, id, newData)
	if err != nil {
		fmt.Printf("Failed to edit data: %s\n", err)
	} else {
		fmt.Printf("Title: %s, Text: %s\n", title, text)
		fmt.Println("Data edited successfully!")
	}
}
func (c *Client) editBinaryData() {
	tableName, id := c.selectData()
	if tableName == "" || id == "" {
		return
	}
	c.rl.SetPrompt("Choose a new title (meta-information): ")
	title, _ := c.rl.Readline()
	c.rl.SetPrompt("Specify the new file path: ")
	filePath, _ := c.rl.Readline()
	newData := map[string]string{
		"path":      filePath,
		"meta_info": title,
	}
	err := c.service.UpdateData(c.userID, id, newData)
	if err != nil {
		fmt.Printf("Failed to edit data: %s\n", err)
	} else {
		fmt.Printf("Title: %s, File: %s\n", title, filePath)
		fmt.Println("Data edited successfully!")
	}
}
func (c *Client) editBankCardData() {
	tableName, id := c.selectData()
	if tableName == "" || id == "" {
		return
	}
	c.rl.SetPrompt("Choose a new title (meta-information): ")
	title, _ := c.rl.Readline()
	c.rl.SetPrompt("Enter new card number: ")
	cardNumber, _ := c.rl.Readline()
	c.rl.SetPrompt("Enter new expiry date (MM/YY): ")
	expiryDate, _ := c.rl.Readline()
	c.rl.SetPrompt("Enter new CVV: ")
	cvv, _ := c.rl.Readline()
	newData := map[string]string{
		"card_number":     cardNumber,
		"expiration_date": expiryDate,
		"cvv":             cvv,
		"meta_info":       title,
	}
	err := c.service.UpdateData(c.userID, id, newData)
	if err != nil {
		fmt.Printf("Failed to edit data: %s\n", err)
	} else {
		fmt.Printf("Title: %s, Card Number: %s, Expiry Date: %s, CVV: %s\n", title, cardNumber, expiryDate, cvv)
		fmt.Println("Data edited successfully!")
	}
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

	data, _ := c.service.GetAllData(c.userID, tableName, "id", "meta_info")
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
		fmt.Println("Multiple entries found. Please enter the id of the entry you want to delete.")
		line, _ = c.rl.Readline()
		for _, entry := range entriesToDelete {
			if entry["id"] == line {
				fmt.Println("Are you sure you want to delete this entry? (yes/no)")
				line, _ = c.rl.Readline()
				if strings.ToLower(line) == "yes" {

					err := c.service.DeleteData(c.userID, tableName, entry["id"], entry["meta_info"])
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
		line, _ = c.rl.Readline()
		if strings.ToLower(line) == "yes" {
			err := c.service.DeleteData(c.userID, tableName, entriesToDelete[0]["id"], entriesToDelete[0]["meta_info"])
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
