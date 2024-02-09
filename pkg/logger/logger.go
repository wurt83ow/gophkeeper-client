package logger

import (
	"log"
	"os"
)

// LoggerInterface defines the methods that your logger should implement.
type LoggerInterface interface {
	Printf(format string, v ...interface{})
}

// Logger represents a logger instance.
type Logger struct {
	*log.Logger
}

// NewLogger creates a new instance of LoggerInterface.
func NewLogger() LoggerInterface {
	logFile, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}

	return &Logger{
		Logger: log.New(logFile, "", log.LstdFlags),
	}
}
