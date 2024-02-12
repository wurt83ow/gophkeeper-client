package encription

import (
	"os"
	"testing"
)

// Тест проверяет шифрование и дешифрование данных.
func TestEncryptDecrypt(t *testing.T) {
	key := "supersecretkey"
	data := "hello world"

	enc := NewEnc(key)

	// Шифруем данные
	encryptedData, err := enc.Encrypt(data)
	if err != nil {
		t.Fatal("Error encrypting data:", err)
	}

	// Дешифруем данные
	decryptedData, err := enc.Decrypt(encryptedData)
	if err != nil {
		t.Fatal("Error decrypting data:", err)
	}

	// Проверяем, что данные после дешифрования совпадают с оригинальными данными
	if decryptedData != data {
		t.Errorf("Expected decrypted data to be %s, got %s", data, decryptedData)
	}
}

// Тест проверяет шифрование и дешифрование файлов.
func TestEncryptDecryptFile(t *testing.T) {
	key := "supersecretkey"
	inputFilePath := "test.txt"
	outputFilePath := "encrypted_test.txt"

	// Создаем временный файл с данными
	data := "hello world"
	if err := os.WriteFile(inputFilePath, []byte(data), 0644); err != nil {
		t.Fatal("Error creating test file:", err)
	}
	defer os.Remove(inputFilePath)

	enc := NewEnc(key)

	// Шифруем файл
	encryptedFilePath, _, err := enc.EncryptFile(inputFilePath, "")
	if err != nil {
		t.Fatal("Error encrypting file:", err)
	}

	// Дешифруем файл
	err = enc.DecryptFile(encryptedFilePath, outputFilePath)
	if err != nil {
		t.Fatal("Error decrypting file:", err)
	}

	// Сравниваем содержимое исходного и дешифрованного файла
	originalData, err := os.ReadFile(inputFilePath)
	if err != nil {
		t.Fatal("Error reading original file:", err)
	}

	decryptedData, err := os.ReadFile(outputFilePath)
	if err != nil {
		t.Fatal("Error reading decrypted file:", err)
	}

	if string(originalData) != string(decryptedData) {
		t.Error("Decrypted data does not match original data")
	}
}
