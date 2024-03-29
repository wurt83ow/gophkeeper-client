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

// Enc represents an encryption service.
type Enc struct {
	key []byte // key is the encryption key.
}

// NewEnc creates a new instance of Enc with the provided key.
func NewEnc(key string) *Enc {
	hash := sha256.New()
	hash.Write([]byte(key))
	return &Enc{
		key: hash.Sum(nil),
	}
}

// IsBase64 checks if the provided string is base64 encoded.
func (e *Enc) IsBase64(s string) bool {
	_, err := base64.URLEncoding.DecodeString(s)
	return err == nil
}

// Decrypt decrypts the provided base64-encoded encrypted text.
func (e *Enc) Decrypt(encryptedText string) (string, error) {
	if !e.IsBase64(encryptedText) {
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

// Encrypt encrypts the provided data.
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

// DecryptFile decrypts a file.
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

// EncryptFile encrypts a file.
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

	// Write the nonce to the beginning of the file
	if _, err := outputFile.Write(nonce); err != nil {
		return "", bs, err
	}

	stream := cipher.NewCTR(block, nonce)

	// Reset file pointer to the beginning
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

// HashPassword hashes the provided password using bcrypt.
func (e *Enc) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CompareHashAndPassword compares a hashed password with its possible plaintext equivalent.
func (e *Enc) CompareHashAndPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// GetHash returns the hash of the provided data.
func (e *Enc) GetHash(data []byte) string {
	hasher := sha256.New()
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}
