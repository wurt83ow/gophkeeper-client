package logger

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	// Создаем новый логгер
	logger := NewLogger()

	// Записываем сообщение в лог
	testMessage := "Test message"
	logger.Printf(testMessage)

	// Открываем файл лога и проверяем, что последняя строка соответствует нашему сообщению
	file, err := os.Open("log.txt")
	if err != nil {
		t.Fatalf("Error opening log file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lastLine string
	for scanner.Scan() {
		lastLine = scanner.Text()
	}

	if !strings.Contains(lastLine, testMessage) {
		t.Errorf("Log file did not contain the expected message. Got: %s, want: %s", lastLine, testMessage)
	}
}
