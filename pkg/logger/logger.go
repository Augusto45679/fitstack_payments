// Package logger provides structured logging utilities.
package logger

import (
	"context"
	"log"
	"os"
)

// Logger provides structured logging.
type Logger struct {
	*log.Logger
}

// NewLogger creates a new logger.
func NewLogger() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "[FitStack-Payments] ", log.LstdFlags|log.Lshortfile),
	}
}

// Info logs an info message.
func (l *Logger) Info(ctx context.Context, message string, fields ...interface{}) {
	l.Printf("[INFO] %s %v", message, fields)
}

// Error logs an error message.
func (l *Logger) Error(ctx context.Context, message string, err error, fields ...interface{}) {
	l.Printf("[ERROR] %s: %v %v", message, err, fields)
}

// Warn logs a warning message.
func (l *Logger) Warn(ctx context.Context, message string, fields ...interface{}) {
	l.Printf("[WARN] %s %v", message, fields)
}

// Debug logs a debug message.
func (l *Logger) Debug(ctx context.Context, message string, fields ...interface{}) {
	l.Printf("[DEBUG] %s %v", message, fields)
}
