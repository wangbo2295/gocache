package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger represents a simple logger with configurable output and level
type Logger struct {
	mu       sync.Mutex
	level    LogLevel
	output   io.Writer
	filename string
	file     *os.File
}

// Global logger instance
var std = &Logger{
	level:  INFO,
	output: os.Stdout,
}

// SetLevel sets the global log level
func SetLevel(level string) {
	std.mu.Lock()
	defer std.mu.Unlock()

	switch strings.ToLower(level) {
	case "debug":
		std.level = DEBUG
	case "info":
		std.level = INFO
	case "warn", "warning":
		std.level = WARN
	case "error":
		std.level = ERROR
	default:
		std.level = INFO
	}
}

// SetOutput sets the output destination for the logger
func SetOutput(w io.Writer) {
	std.mu.Lock()
	defer std.mu.Unlock()

	std.output = w
}

// SetFile sets the log file output
func SetFile(filename string) error {
	std.mu.Lock()
	defer std.mu.Unlock()

	// Close existing file if open
	if std.file != nil {
		std.file.Close()
		std.file = nil
	}

	if filename == "" {
		std.output = os.Stdout
		std.filename = ""
		return nil
	}

	// Open file in append mode, create if not exists
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	std.file = file
	std.output = file
	std.filename = filename
	return nil
}

// Close closes the log file if one is open
func Close() error {
	std.mu.Lock()
	defer std.mu.Unlock()

	if std.file != nil {
		err := std.file.Close()
		std.file = nil
		return err
	}
	return nil
}

// log is the internal logging method
func log(level LogLevel, format string, args ...interface{}) {
	std.mu.Lock()
	defer std.mu.Unlock()

	// Check if the message should be logged based on level
	if level < std.level {
		return
	}

	// Format the message
	message := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	logLine := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level.String(), message)

	// Write to output
	std.output.Write([]byte(logLine))
}

// Debug logs a debug message
func Debug(format string, args ...interface{}) {
	log(DEBUG, format, args...)
}

// Info logs an info message
func Info(format string, args ...interface{}) {
	log(INFO, format, args...)
}

// Warn logs a warning message
func Warn(format string, args ...interface{}) {
	log(WARN, format, args...)
}

// Error logs an error message
func Error(format string, args ...interface{}) {
	log(ERROR, format, args...)
}

// Fatal logs a fatal message and exits the program
func Fatal(format string, args ...interface{}) {
	std.mu.Lock()
	defer std.mu.Unlock()

	message := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	logLine := fmt.Sprintf("[%s] [FATAL] %s\n", timestamp, message)

	std.output.Write([]byte(logLine))
	os.Exit(1)
}
