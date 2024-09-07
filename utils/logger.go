package utils

import (
	"fmt"
	"log"
	"os"
)

type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	WithPrefix(prefix string) Logger
}

type StandardLogger struct {
	logger *log.Logger
	prefix string
}

func NewStandardLogger(prefix string) *StandardLogger {
	return &StandardLogger{
		logger: log.New(os.Stdout, prefix, log.LstdFlags),
		prefix: prefix,
	}
}

// Info logs informational messages
func (l *StandardLogger) Info(msg string, args ...interface{}) {
	l.logger.Printf("[INFO] "+msg, args...)
}

// Error logs error messages
func (l *StandardLogger) Error(msg string, args ...interface{}) {
	l.logger.Printf("[ERROR] "+msg, args...)
}

// Warn logs warning messages
func (l *StandardLogger) Warn(msg string, args ...interface{}) {
	l.logger.Printf("[WARN] "+msg, args...)
}

func (l *StandardLogger) WithPrefix(prefix string) Logger {
	return NewStandardLogger(fmt.Sprintf("%s: %s", l.prefix, prefix))
}
