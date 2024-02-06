package logger

import (
	"log"
	"os"
)

// LoggerInterface определяет методы, которые должен реализовать ваш логгер
type LoggerInterface interface {
	Printf(format string, v ...interface{})
}

type Logger struct {
	*log.Logger
}

func NewLogger() LoggerInterface {
	logFile, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Ошибка при открытии файла журнала: %v", err)
	}

	return &Logger{
		Logger: log.New(logFile, "", log.LstdFlags),
	}
}
