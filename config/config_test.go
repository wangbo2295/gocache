package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaultConfig(t *testing.T) {
	// Reset to defaults
	Config = &Properties{
		Bind:           "127.0.0.1",
		Port:           6379,
		Databases:      16,
		MaxClients:     10000,
		Timeout:        0,
		AppendOnly:     false,
		AppendFilename: "appendonly.aof",
		AppendFsync:    "everysec",
		DBFilename:     "dump.rdb",
		LogLevel:       "info",
		LogFile:        "",
		RequirePass:    "",
	}

	// Test loading non-existent file (should use defaults)
	err := Load("nonexistent.conf")
	if err != nil {
		t.Errorf("Load() failed for non-existent file: %v", err)
	}

	// Verify defaults are preserved
	if Config.Bind != "127.0.0.1" {
		t.Errorf("Expected Bind to be 127.0.0.1, got %s", Config.Bind)
	}
	if Config.Port != 6379 {
		t.Errorf("Expected Port to be 6379, got %d", Config.Port)
	}
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.conf")
	configContent := `# GoCache Configuration
bind 0.0.0.0
port 6380
databases 32
maxclients 20000
timeout 300
appendonly yes
appendfilename test.aof
appendfsync always
dbfilename test.rdb
loglevel debug
logfile test.log
requirepass mypassword
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load configuration
	Config = &Properties{}
	err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify loaded values
	tests := []struct {
		name     string
		expected interface{}
		actual   interface{}
	}{
		{"Bind", "0.0.0.0", Config.Bind},
		{"Port", 6380, Config.Port},
		{"Databases", 32, Config.Databases},
		{"MaxClients", 20000, Config.MaxClients},
		{"Timeout", 300, Config.Timeout},
		{"AppendOnly", true, Config.AppendOnly},
		{"AppendFilename", "test.aof", Config.AppendFilename},
		{"AppendFsync", "always", Config.AppendFsync},
		{"DBFilename", "test.rdb", Config.DBFilename},
		{"LogLevel", "debug", Config.LogLevel},
		{"LogFile", "test.log", Config.LogFile},
		{"RequirePass", "mypassword", Config.RequirePass},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expected != tt.actual {
				t.Errorf("Expected %s to be %v, got %v", tt.name, tt.expected, tt.actual)
			}
		})
	}
}

func TestLoadConfigWithComments(t *testing.T) {
	// Test that comments and empty lines are handled correctly
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.conf")
	configContent := `# This is a comment
bind 192.168.1.1

# Another comment
port 6379

`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	Config = &Properties{}
	err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if Config.Bind != "192.168.1.1" {
		t.Errorf("Expected Bind to be 192.168.1.1, got %s", Config.Bind)
	}
	if Config.Port != 6379 {
		t.Errorf("Expected Port to be 6379, got %d", Config.Port)
	}
}

func TestLoadConfigWithQuotes(t *testing.T) {
	// Test that quoted values are handled correctly
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.conf")
	configContent := `requirepass "my password with spaces"
logfile "/var/log/gocache.log"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	Config = &Properties{}
	err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if Config.RequirePass != "my password with spaces" {
		t.Errorf("Expected RequirePass to be 'my password with spaces', got %s", Config.RequirePass)
	}
	if Config.LogFile != "/var/log/gocache.log" {
		t.Errorf("Expected LogFile to be '/var/log/gocache.log', got %s", Config.LogFile)
	}
}

func TestLoadConfigInvalidPort(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.conf")
	configContent := `port 99999
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	Config = &Properties{}
	err := Load(configPath)
	if err == nil {
		t.Error("Expected error for invalid port, got nil")
	}
}

func TestLoadConfigInvalidAppendFsync(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.conf")
	configContent := `appendfsync invalid
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	Config = &Properties{}
	err := Load(configPath)
	if err == nil {
		t.Error("Expected error for invalid appendfsync, got nil")
	}
}

func TestLoadConfigUnknownKey(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.conf")
	configContent := `unknown_key value
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	Config = &Properties{}
	err := Load(configPath)
	if err == nil {
		t.Error("Expected error for unknown config key, got nil")
	}
}
