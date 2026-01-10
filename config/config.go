package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Properties holds all configuration properties for GoCache
type Properties struct {
	// Server configuration
	Bind     string
	Port     int
	Databases int

	// Client configuration
	MaxClients int
	Timeout    int // 0 means no timeout

	// Persistence configuration
	AppendOnly      bool
	AppendFilename  string
	AppendFsync     string // always, everysec, no
	DBFilename      string

	// Logging configuration
	LogLevel string // debug, info, warn, error
	LogFile  string

	// Security
	RequirePass string
}

// Global configuration instance
var Config = &Properties{
	// Set default values
	Bind:          "127.0.0.1",
	Port:          16379,
	Databases:     16,
	MaxClients:    10000,
	Timeout:       0,
	AppendOnly:    false,
	AppendFilename: "appendonly.aof",
	AppendFsync:   "everysec",
	DBFilename:    "dump.rdb",
	LogLevel:      "info",
	LogFile:       "",
	RequirePass:   "",
}

// Load loads configuration from file
func Load(configPath string) error {
	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file doesn't exist, use defaults
			return nil
		}
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key-value pairs
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid config at line %d: %s", lineNum, line)
		}

		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = value[1 : len(value)-1]
		}

		// Set configuration
		if err := setConfig(key, value); err != nil {
			return fmt.Errorf("failed to set config %s at line %d: %w", key, lineNum, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	return nil
}

// setConfig sets a single configuration value
func setConfig(key, value string) error {
	switch key {
	case "bind":
		Config.Bind = value
	case "port":
		port, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid port: %s", value)
		}
		if port < 1 || port > 65535 {
			return fmt.Errorf("port out of range: %d", port)
		}
		Config.Port = port
	case "databases":
		dbs, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid databases: %s", value)
		}
		if dbs < 1 || dbs > 256 {
			return fmt.Errorf("databases out of range: %d", dbs)
		}
		Config.Databases = dbs
	case "maxclients":
		clients, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid maxclients: %s", value)
		}
		Config.MaxClients = clients
	case "timeout":
		timeout, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid timeout: %s", value)
		}
		Config.Timeout = timeout
	case "appendonly":
		Config.AppendOnly = strings.ToLower(value) == "yes"
	case "appendfilename":
		Config.AppendFilename = value
	case "appendfsync":
		fsync := strings.ToLower(value)
		if fsync != "always" && fsync != "everysec" && fsync != "no" {
			return fmt.Errorf("invalid appendfsync: %s (must be always, everysec, or no)", value)
		}
		Config.AppendFsync = fsync
	case "dbfilename":
		Config.DBFilename = value
	case "loglevel":
		level := strings.ToLower(value)
		if level != "debug" && level != "info" && level != "warn" && level != "error" {
			return fmt.Errorf("invalid loglevel: %s (must be debug, info, warn, or error)", value)
		}
		Config.LogLevel = level
	case "logfile":
		Config.LogFile = value
	case "requirepass":
		Config.RequirePass = value
	default:
		// Ignore unknown config keys for now
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}
