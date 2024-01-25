package encription

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
)

type Enc struct {
	key []byte
}

func NewEnc(key string) *Enc {
	hash := sha256.New()
	hash.Write([]byte(key))
	return &Enc{
		key: hash.Sum(nil),
	}
}

// func (e *Enc) Decrypt(encryptedText string) (string, error) {
// 	ciphertext, err := base64.URLEncoding.DecodeString(encryptedText)
// 	if err != nil {
// 		return "", err
// 	}

// 	block, err := aes.NewCipher(e.key)
// 	if err != nil {
// 		return "", err
// 	}

// 	if len(ciphertext) < aes.BlockSize {
// 		return "", errors.New("ciphertext too short")
// 	}

// 	iv := ciphertext[:aes.BlockSize]
// 	ciphertext = ciphertext[aes.BlockSize:]

// 	stream := cipher.NewCFBDecrypter(block, iv)
// 	stream.XORKeyStream(ciphertext, ciphertext)

// 	return string(ciphertext), nil
// }

func isBase64(s string) bool {
	_, err := base64.URLEncoding.DecodeString(s)
	return err == nil
}

func (e *Enc) Decrypt(encryptedText string) (string, error) {

	if !isBase64(encryptedText) {
		return "", errors.New("invalid base64 data")
	}
	ciphertext, err := base64.URLEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 data: %w", err)
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}

func (e *Enc) Encrypt(data string) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(data))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(data))

	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func (e *Enc) DecryptFile(inputPath string, outputPath string) error {
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return err
	}

	nonce := make([]byte, 16)
	if _, err := io.ReadFull(inputFile, nonce); err != nil {
		return err
	}

	stream := cipher.NewCTR(block, nonce)

	buf := make([]byte, 1024)
	for {
		n, err := inputFile.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		stream.XORKeyStream(buf[:n], buf[:n])

		if _, err := outputFile.Write(buf[:n]); err != nil {
			return err
		}
	}

	return nil
}

func (e *Enc) EncryptFile(inputPath string, outputPath string) (string, []byte, error) {
	bs := []byte{}

	inputFile, err := os.Open(inputPath)
	if err != nil {
		return "", bs, err
	}
	defer inputFile.Close()

	hasher := sha256.New()
	buf := make([]byte, 1024)
	for {
		n, err := inputFile.Read(buf)
		if err != nil && err != io.EOF {
			return "", bs, err
		}
		if n == 0 {
			break
		}

		hasher.Write(buf[:n])
	}

	hash := hasher.Sum(nil)
	encryptedFilePath := filepath.Join(outputPath, fmt.Sprintf("%x", hash))

	outputFile, err := os.Create(encryptedFilePath)
	if err != nil {
		return "", bs, err
	}
	defer outputFile.Close()

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", bs, err
	}

	nonce := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", bs, err
	}

	// Записываем nonce в начало файла
	if _, err := outputFile.Write(nonce); err != nil {
		return "", bs, err
	}

	stream := cipher.NewCTR(block, nonce)

	// Сбрасываем указатель файла на начало
	inputFile.Seek(0, 0)

	for {
		n, err := inputFile.Read(buf)
		if err != nil && err != io.EOF {
			return "", bs, err
		}
		if n == 0 {
			break
		}

		stream.XORKeyStream(buf[:n], buf[:n])

		if _, err := outputFile.Write(buf[:n]); err != nil {
			return "", bs, err
		}
	}

	return encryptedFilePath, hash, nil
}

func (e *Enc) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (e *Enc) CompareHashAndPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
func (e *Enc) GetHash(data []byte) string {
	hasher := sha256.New()
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}
