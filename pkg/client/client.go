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

	"github.com/chzyer/readline"
)

type GophKeeper struct {
	rl *readline.Instance
}

func NewGophKeeper() *GophKeeper {
	rl, err := readline.New("> ")
	if err != nil {
		log.Fatal(err)
	}
	return &GophKeeper{rl: rl}
}

func (gk *GophKeeper) Close() {
	gk.rl.Close()
}

func (gk *GophKeeper) addLoginPassword() {
	gk.rl.SetPrompt("Enter login: ")
	login, _ := gk.rl.Readline()
	gk.rl.SetPrompt("Enter password: ")
	gk.rl.Config.EnableMask = true
	password, _ := gk.rl.Readline()
	gk.rl.Config.EnableMask = false
	fmt.Printf("Login: %s, Password: %s\n", login, password)
	fmt.Println("Data added successfully!")
}

func (gk *GophKeeper) addTextData() {
	gk.rl.SetPrompt("Choose a title (meta-information): ")
	title, _ := gk.rl.Readline()
	gk.rl.SetPrompt("Enter text data: ")
	text, _ := gk.rl.Readline()
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
		block, err := aes.NewCipher([]byte("example key 1234")) // Replace with your 16, 24, or 32 byte key
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

	fmt.Printf("Title: %s, Card Number: %s, Expiry Date: %s, CVV: %s\n", title, cardNumber, expiryDate, cvv)
	fmt.Println("Data added successfully!")
}
