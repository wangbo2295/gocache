package database

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/wangbo/gocache/config"
	"github.com/wangbo/gocache/persistence"
	"github.com/wangbo/gocache/replication"
)

// Management command implementations

func execPing(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) == 0 {
		return [][]byte{[]byte("PONG")}, nil
	}
	return [][]byte{args[0]}, nil
}

func execInfo(db *DB, args [][]byte) ([][]byte, error) {
	section := "default"
	if len(args) > 0 {
		section = strings.ToLower(string(args[0]))
	}

	var info string
	switch section {
	case "memory", "stats":
		info = execInfoMemoryString(db)
	default:
		info = execInfoDefaultString(db)
	}

	return [][]byte{[]byte(info)}, nil
}

func execInfoDefaultString(db *DB) string {
	var builder strings.Builder

	builder.WriteString("# Server\r\n")
	builder.WriteString("redis_version:6.2.0\r\n")
	builder.WriteString("go_cache_version:1.0.0\r\n")
	builder.WriteString("os:" + runtimeOS() + "\r\n")
	builder.WriteString("arch:" + runtimeArch() + "\r\n")
	builder.WriteString("process_id:" + strconv.FormatInt(int64(getPID()), 10) + "\r\n")
	builder.WriteString("tcp_port:" + strconv.Itoa(config.Config.Port) + "\r\n")
	builder.WriteString("uptime_in_seconds:" + strconv.FormatInt(int64(getUptime()), 10) + "\r\n")
	builder.WriteString("uptime_in_days:0\r\n")
	builder.WriteString("\r\n")

	builder.WriteString("# Clients\r\n")
	builder.WriteString("connected_clients:1\r\n")
	builder.WriteString("maxclients:10000\r\n")
	builder.WriteString("\r\n")

	builder.WriteString("# Memory\r\n")
	builder.WriteString("used_memory:" + strconv.FormatInt(db.GetUsedMemory(), 10) + "\r\n")
	builder.WriteString("used_memory_human:" + formatBytes(db.GetUsedMemory()) + "\r\n")
	builder.WriteString("maxmemory:" + strconv.FormatInt(config.Config.MaxMemory, 10) + "\r\n")
	builder.WriteString("maxmemory_human:" + formatBytes(config.Config.MaxMemory) + "\r\n")
	builder.WriteString("maxmemory_policy:" + config.Config.MaxMemoryPolicy + "\r\n")
	builder.WriteString("\r\n")

	builder.WriteString("# Stats\r\n")
	builder.WriteString("total_connections_received:1\r\n")
	builder.WriteString("total_commands_processed:10\r\n")
	builder.WriteString("instantaneous_ops_per_sec:0\r\n")
	builder.WriteString("\r\n")

	builder.WriteString("# Replication\r\n")
	builder.WriteString("role:" + replication.State.GetRole().String() + "\r\n")
	if replication.State.IsMaster() {
		builder.WriteString("connected_slaves:" + strconv.Itoa(replication.State.GetSlaveCount()) + "\r\n")
	} else {
		masterHost, masterPort := replication.State.GetMasterInfo()
		builder.WriteString("master_host:" + masterHost + "\r\n")
		builder.WriteString("master_port:" + strconv.Itoa(masterPort) + "\r\n")
		builder.WriteString("master_link_status:up\r\n")
	}
	builder.WriteString("replid:" + strconv.FormatUint(replication.State.GetReplicationID(), 10) + "\r\n")
	builder.WriteString("repl_offset:" + strconv.FormatUint(replication.State.GetReplicationOffset(), 10) + "\r\n")
	builder.WriteString("\r\n")

	builder.WriteString("# Persistence\r\n")
	builder.WriteString("loading:0\r\n")
	builder.WriteString("aof_enabled:" + strconv.FormatBool(config.Config.AppendOnly) + "\r\n")
	if !db.lastSaveTime.IsZero() {
		builder.WriteString("rdb_last_save_time:" + strconv.FormatInt(db.lastSaveTime.Unix(), 10) + "\r\n")
		builder.WriteString("rdb_last_save_time_elapsed:" + strconv.FormatInt(int64(time.Since(db.lastSaveTime).Seconds()), 10) + "\r\n")
	} else {
		builder.WriteString("rdb_last_save_time:0\r\n")
	}
	if db.bgSaveInProgress {
		builder.WriteString("bgsave_in_progress:1\r\n")
	} else {
		builder.WriteString("bgsave_in_progress:0\r\n")
	}
	builder.WriteString("\r\n")

	builder.WriteString("# Slow Log\r\n")
	builder.WriteString("slowlog_len:" + strconv.Itoa(db.GetSlowLogLen()) + "\r\n")
	builder.WriteString("slowlog_max_len:" + strconv.Itoa(db.slowLogMaxLen) + "\r\n")
	builder.WriteString("\r\n")

	return builder.String()
}

func execInfoMemoryString(db *DB) string {
	var builder strings.Builder

	builder.WriteString("# Memory\r\n")
	builder.WriteString("used_memory:" + strconv.FormatInt(db.GetUsedMemory(), 10) + "\r\n")
	builder.WriteString("used_memory_human:" + formatBytes(db.GetUsedMemory()) + "\r\n")
	builder.WriteString("maxmemory:" + strconv.FormatInt(config.Config.MaxMemory, 10) + "\r\n")
	builder.WriteString("maxmemory_human:" + formatBytes(config.Config.MaxMemory) + "\r\n")
	builder.WriteString("maxmemory_policy:" + config.Config.MaxMemoryPolicy + "\r\n")
	builder.WriteString("\r\n")

	if db.evictionPolicy != nil {
		builder.WriteString("# Eviction Policy\r\n")
		builder.WriteString("policy_type:" + config.Config.MaxMemoryPolicy + "\r\n")
		builder.WriteString("\r\n")
	}

	return builder.String()
}

func execMemory(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 1 {
		return nil, errors.New("wrong number of arguments for MEMORY")
	}

	subCmd := strings.ToLower(string(args[0]))

	switch subCmd {
	case "usage":
		if len(args) != 2 {
			return nil, errors.New("wrong number of arguments for MEMORY USAGE")
		}
		key := string(args[1])

		entity, ok := db.GetEntity(key)
		if !ok || entity == nil {
			return [][]byte{[]byte("0")}, nil
		}

		size := entity.EstimateSize()
		return [][]byte{[]byte(strconv.FormatInt(size, 10))}, nil

	case "stats":
		info := make([][]byte, 0)
		info = append(info, []byte("used_memory:"+strconv.FormatInt(db.GetUsedMemory(), 10)))
		info = append(info, []byte("used_memory_human:"+formatBytes(db.GetUsedMemory())))
		info = append(info, []byte("maxmemory:"+strconv.FormatInt(config.Config.MaxMemory, 10)))
		info = append(info, []byte("maxmemory_human:"+formatBytes(config.Config.MaxMemory)))
		return info, nil

	default:
		return nil, errors.New("unknown MEMORY subcommand")
	}
}

func formatBytes(bytes int64) string {
	if bytes < 1024 {
		return strconv.FormatInt(bytes, 10) + "b"
	}
	units := []string{"kb", "mb", "gb", "tb"}
	value := float64(bytes)
	for _, unit := range units {
		value /= 1024
		if value < 1024 {
			return strconv.FormatFloat(value, 'f', 2, 64) + unit
		}
	}
	return strconv.FormatFloat(value, 'f', 2, 64) + "pb"
}

func getPID() int {
	return 1000
}

func getUptime() int64 {
	return 3600
}

func runtimeOS() string {
	return "darwin"
}

func runtimeArch() string {
	return "amd64"
}

// execSave synchronously saves the database to disk
func execSave(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, errors.New("wrong number of arguments for SAVE")
	}

	// Get RDB filename from config
	rdbFilename := config.Config.DBFilename
	if rdbFilename == "" {
		rdbFilename = "dump.rdb"
	}

	// Save database using registered saver
	if err := persistence.SaveDatabase(db, rdbFilename); err != nil {
		return nil, err
	}

	// Update last save time in DB
	db.lastSaveTime = time.Now()

	return [][]byte{[]byte("OK")}, nil
}

// execBgSave asynchronously saves the database to disk
func execBgSave(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, errors.New("wrong number of arguments for BGSAVE")
	}

	db.bgSaveMu.Lock()
	defer db.bgSaveMu.Unlock()

	// Check if already saving in background
	if db.bgSaveInProgress {
		return nil, errors.New("Background save already in progress")
	}

	// Get RDB filename from config
	rdbFilename := config.Config.DBFilename
	if rdbFilename == "" {
		rdbFilename = "dump.rdb"
	}

	// Start background save
	db.bgSaveInProgress = true
	db.bgSaveStartTime = time.Now()

	go func() {
		defer func() {
			db.bgSaveMu.Lock()
			db.bgSaveInProgress = false
			db.lastSaveTime = time.Now()
			db.bgSaveMu.Unlock()
		}()

		if err := persistence.SaveDatabase(db, rdbFilename); err != nil {
			// Log error (in real implementation)
			return
		}
	}()

	return [][]byte{[]byte("Background saving started")}, nil
}

// execSlaveOf sets the instance as a slave of the specified master
func execSlaveOf(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("wrong number of arguments for SLAVEOF")
	}

	host := string(args[0])
	portStr := string(args[1])

	// Handle "SLAVEOF NO ONE" - become a master
	if host == "NO" && portStr == "ONE" {
		replication.State.SetAsMaster()
		return [][]byte{[]byte("OK")}, nil
	}

	// Parse port
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, errors.New("invalid port number")
	}

	// Set as slave
	if err := replication.State.SetAsSlave(host, port); err != nil {
		return nil, err
	}

	// Initiate synchronization with master in background
	go func() {
		if err := performSynchronization(db); err != nil {
			fmt.Printf("Synchronization failed: %v\n", err)
		}
	}()

	return [][]byte{[]byte("OK")}, nil
}

// performSynchronization performs full synchronization with master
func performSynchronization(db *DB) error {
	// Perform full sync
	rdbData, err := replication.State.PerformFullSync()
	if err != nil {
		return fmt.Errorf("full sync failed: %w", err)
	}

	// Load RDB data into database
	if err := loadRDBFromBytes(db, rdbData); err != nil {
		return fmt.Errorf("failed to load RDB: %w", err)
	}

	fmt.Printf("Successfully synchronized with master\n")

	// Start replication loop to receive propagated commands
	adapter := replication.NewDBCommandAdapter(db)
	if err := replication.State.StartReplicationLoop(adapter); err != nil {
		return fmt.Errorf("failed to start replication loop: %w", err)
	}

	fmt.Printf("Replication loop started\n")
	return nil
}

// loadRDBFromBytes loads RDB data from bytes into database
func loadRDBFromBytes(db *DB, data []byte) error {
	// Use the replication package's RDB loader to avoid circular imports
	return replication.LoadRDBData(db, data)
}

// execSync initiates a full synchronization with the master
func execSync(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, errors.New("wrong number of arguments for SYNC")
	}

	// This command is received from a slave
	// TODO: Send RDB file to slave
	return [][]byte{[]byte("OK")}, nil
}

// execPSync initiates a partial synchronization with the master
func execPSync(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("wrong number of arguments for PSYNC")
	}

	// This command is received from a slave
	// args[0] = replication ID
	// args[1] = replication offset

	// TODO: Send incremental updates to slave if available
	// For now, fall back to full sync
	return [][]byte{[]byte("FULLRESYNC")}, nil
}

// execAuth authenticates the connection
// Note: AUTH is now handled at the server level (server/server.go:handleAuth)
// This function is kept for registry compatibility but should not be called directly
func execAuth(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments for AUTH")
	}

	// AUTH is handled at the connection level before commands reach the database
	// This function is only called if the server-level handling fails
	return nil, errors.New("AUTH should be handled at server level")
}

// execSlowLog manages the slow log
func execSlowLog(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 1 {
		return nil, errors.New("wrong number of arguments for SLOWLOG")
	}

	subCmd := strings.ToLower(string(args[0]))

	switch subCmd {
	case "get":
		// Return all slow log entries
		entries := db.GetSlowLogEntries()
		return formatSlowLogEntries(entries), nil

	case "len":
		// Return number of slow log entries
		return [][]byte{[]byte(strconv.Itoa(db.GetSlowLogLen()))}, nil

	case "reset":
		// Reset slow log
		db.ResetSlowLog()
		return [][]byte{[]byte("OK")}, nil

	default:
		return nil, errors.New("unknown SLOWLOG subcommand")
	}
}

// formatSlowLogEntries formats slow log entries for output
func formatSlowLogEntries(entries []*SlowLogEntry) [][]byte {
	result := make([][]byte, len(entries))

	for i, entry := range entries {
		// Format: (integer) (timestamp) (microseconds) (command)
		line := fmt.Sprintf("%d) (timestamp=%s) (microseconds=%d) %s",
			i+1,
			entry.Timestamp.Format("2006-01-02 15:04:05.000"),
			entry.Duration,
			string(entry.Command))
		result[i] = []byte(line)
	}

	return result
}

// execMonitor enables command monitoring
// Note: This is a special command that requires server-level handling
// The database layer just returns OK, actual monitoring is handled in server layer
func execMonitor(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, errors.New("wrong number of arguments for MONITOR")
	}

	// Return a special response to indicate monitoring mode
	// The server layer will handle this specially
	return [][]byte{[]byte("OK")}, nil
}
