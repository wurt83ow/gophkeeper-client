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

// func (e *Enc) DecryptData(ciphertext []byte) ([]byte, error) {
// 	block, err := aes.NewCipher(e.key)
// 	if err != nil {
// 		return nil, err
// 	}

// 	gcm, err := cipher.NewGCM(block)
// 	if err != nil {
// 		return nil, err
// 	}

// 	nonceSize := gcm.NonceSize()
// 	if len(ciphertext) < nonceSize {
// 		return nil, errors.New("ciphertext too short")
// 	}

// 	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
// 	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return plaintext, nil
// }

// func (e *Enc) DecryptData(ciphertext []byte) ([]byte, error) {
// 	block, err := aes.NewCipher(e.key)
// 	if err != nil {
// 		return nil, err
// 	}

// 	gcm, err := cipher.NewGCM(block)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var plaintext []byte
// 	nonceSize := gcm.NonceSize()

// 	for len(ciphertext) > 0 {
// 		if len(ciphertext) < 2*nonceSize {
// 			return nil, errors.New("ciphertext too short")
// 		}

// 		nonce, block := ciphertext[:nonceSize], ciphertext[nonceSize:2*nonceSize]
// 		ciphertext = ciphertext[2*nonceSize:]

// 		blockDecrypted, err := gcm.Open(nil, nonce, block, nil)
// 		if err != nil {
// 			return nil, err
// 		}

// 		plaintext = append(plaintext, blockDecrypted...)
// 	}

// 	return plaintext, nil
// }

func (e *Enc) DecryptData(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func (e *Enc) DecryptFile(inputFile *os.File, outputFile *os.File) error {
	var decryptedData []byte
	buf := make([]byte, 1024)
	for {
		n, err := inputFile.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		data := buf[:n]
		decryptedPart, err := e.DecryptData(data)
		if err != nil {
			return err
		}

		// Добавляем расшифрованную часть в decryptedData
		decryptedData = append(decryptedData, decryptedPart...)
	}

	_, err := outputFile.Write(decryptedData)
	if err != nil {
		return err
	}

	return nil
}

func (e *Enc) EncryptData(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

func (e *Enc) EncryptFile(file *os.File) ([]byte, []byte, error) {
	var encryptedData []byte
	hasher := sha256.New()
	buf := make([]byte, 1024)
	for {
		n, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, nil, err
		}

		data := buf[:n]
		encryptedPart, err := e.EncryptData(data)
		if err != nil {
			return nil, nil, err
		}

		// Добавляем зашифрованную часть в encryptedData
		encryptedData = append(encryptedData, encryptedPart...)

		// Обновляем хеш данных
		hasher.Write(encryptedPart)
	}

	// Получаем окончательный хеш данных
	hash := hasher.Sum(nil)

	return encryptedData, hash, nil
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
