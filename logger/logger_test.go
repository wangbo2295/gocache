package logger

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestSetLevel(t *testing.T) {
	tests := []struct {
		name  string
		level string
		want  LogLevel
	}{
		{"debug", "debug", DEBUG},
		{"DEBUG", "DEBUG", DEBUG},
		{"info", "info", INFO},
		{"INFO", "INFO", INFO},
		{"warn", "warn", WARN},
		{"WARN", "WARN", WARN},
		{"warning", "warning", WARN},
		{"error", "error", ERROR},
		{"ERROR", "ERROR", ERROR},
		{"invalid", "invalid", INFO}, // Default to INFO for invalid
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetLevel(tt.level)
			if std.level != tt.want {
				t.Errorf("SetLevel(%q) = %v, want %v", tt.level, std.level, tt.want)
			}
		})
	}
}

func TestLogLevel(t *testing.T) {
	tests := []struct {
		level LogLevel
		want  string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.level.String(); got != tt.want {
				t.Errorf("LogLevel.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLogOutput(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel("debug")

	tests := []struct {
		name   string
		level  LogLevel
		logFn  func(string, ...interface{})
		format string
		args   []interface{}
		check  func(string) bool
	}{
		{
			name:   "Debug",
			level:  DEBUG,
			logFn:  Debug,
			format: "test debug message",
			args:   nil,
			check:  func(s string) bool { return strings.Contains(s, "DEBUG") && strings.Contains(s, "test debug message") },
		},
		{
			name:   "Info",
			level:  INFO,
			logFn:  Info,
			format: "test info message",
			args:   nil,
			check:  func(s string) bool { return strings.Contains(s, "INFO") && strings.Contains(s, "test info message") },
		},
		{
			name:   "Warn",
			level:  WARN,
			logFn:  Warn,
			format: "test warn message",
			args:   nil,
			check:  func(s string) bool { return strings.Contains(s, "WARN") && strings.Contains(s, "test warn message") },
		},
		{
			name:   "Error",
			level:  ERROR,
			logFn:  Error,
			format: "test error message",
			args:   nil,
			check:  func(s string) bool { return strings.Contains(s, "ERROR") && strings.Contains(s, "test error message") },
		},
		{
			name:   "Info with formatting",
			level:  INFO,
			logFn:  Info,
			format: "user %s logged in from %s",
			args:   []interface{}{"alice", "192.168.1.1"},
			check:  func(s string) bool { return strings.Contains(s, "alice") && strings.Contains(s, "192.168.1.1") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			std.level = tt.level

			if tt.args == nil {
				tt.logFn(tt.format)
			} else {
				tt.logFn(tt.format, tt.args...)
			}

			output := buf.String()
			if !tt.check(output) {
				t.Errorf("Log output check failed for %s. Got: %s", tt.name, output)
			}
		})
	}
}

func TestLogLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)

	// Set level to INFO, DEBUG messages should not appear
	SetLevel("info")

	Debug("debug message")
	Info("info message")

	output := buf.String()
	if strings.Contains(output, "debug message") {
		t.Error("DEBUG message should not appear when level is INFO")
	}
	if !strings.Contains(output, "info message") {
		t.Error("INFO message should appear when level is INFO")
	}
}

func TestSetFile(t *testing.T) {
	tmpDir := os.TempDir()
	logFile := tmpDir + "/test.log"

	err := SetFile(logFile)
	if err != nil {
		t.Fatalf("SetFile() failed: %v", err)
	}
	defer Close()

	// Write a log message
	Info("test message to file")

	// Close the file to flush
	Close()

	// Read the file and check content
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "test message to file") {
		t.Errorf("Log file does not contain expected message. Got: %s", string(content))
	}

	// Clean up
	os.Remove(logFile)
}

func TestSetEmptyFile(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)

	// Set file then clear it
	tmpDir := os.TempDir()
	logFile := tmpDir + "/test.log"
	SetFile(logFile)
	Close()

	// Now set empty filename (should revert to stdout)
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	SetFile("")
	Info("message to stdout")

	w.Close()
	os.Stdout = old

	var buf2 bytes.Buffer
	buf2.ReadFrom(r)

	if buf2.Len() == 0 {
		t.Error("Expected output to stdout after setting empty filename")
	}

	// Clean up
	os.Remove(logFile)
}
