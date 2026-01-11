package protocol

// Command constants (all uppercase for consistency)
const (
	// String commands
	CmdSet      = "SET"
	CmdGet      = "GET"
	CmdMSet     = "MSET"
	CmdMGet     = "MGET"
	CmdDel      = "DEL"
	CmdExists   = "EXISTS"
	CmdKeys     = "KEYS"
	CmdIncr     = "INCR"
	CmdIncrBy   = "INCRBY"
	CmdDecr     = "DECR"
	CmdDecrBy   = "DECRBY"
	CmdStrLen   = "STRLEN"
	CmdAppend   = "APPEND"
	CmdGetRange = "GETRANGE"

	// Hash commands
	CmdHSet    = "HSET"
	CmdHGet    = "HGET"
	CmdHDel    = "HDEL"
	CmdHExists = "HEXISTS"
	CmdHGetAll = "HGETALL"
	CmdHKeys   = "HKEYS"
	CmdHVals   = "HVALS"
	CmdHLen    = "HLEN"
	CmdHSetNX  = "HSETNX"
	CmdHIncrBy = "HINCRBY"
	CmdHMGet   = "HMGET"
	CmdHMSet   = "HMSET"

	// List commands
	CmdLPush   = "LPUSH"
	CmdRPush   = "RPUSH"
	CmdLPop    = "LPOP"
	CmdRPop    = "RPOP"
	CmdLIndex  = "LINDEX"
	CmdLSet    = "LSET"
	CmdLRange  = "LRANGE"
	CmdLTrim   = "LTRIM"
	CmdLRem    = "LREM"
	CmdLInsert = "LINSERT"
	CmdLLen    = "LLEN"

	// Set commands
	CmdSAdd        = "SADD"
	CmdSRem        = "SREM"
	CmdSIsMember   = "SISMEMBER"
	CmdSMembers    = "SMEMBERS"
	CmdSCard       = "SCARD"
	CmdSPop        = "SPOP"
	CmdSRandMember = "SRANDMEMBER"
	CmdSMove       = "SMOVE"
	CmdSDiff       = "SDIFF"
	CmdSDiffStore  = "SDIFFSTORE"
	CmdSInter      = "SINTER"
	CmdSInterStore = "SINTERSTORE"
	CmdSUnion      = "SUNION"
	CmdSUnionStore = "SUNIONSTORE"

	// Sorted Set commands
	CmdZAdd          = "ZADD"
	CmdZRem          = "ZREM"
	CmdZScore        = "ZSCORE"
	CmdZIncrBy       = "ZINCRBY"
	CmdZCard         = "ZCARD"
	CmdZRank         = "ZRANK"
	CmdZRevRank      = "ZREVRANK"
	CmdZRange        = "ZRANGE"
	CmdZRevRange     = "ZREVRANGE"
	CmdZRangeByScore = "ZRANGEBYSCORE"
	CmdZCount        = "ZCOUNT"

	// TTL commands
	CmdExpire   = "EXPIRE"
	CmdPExpire  = "PEXPIRE"
	CmdExpireAt = "EXPIREAT"
	CmdPExpireAt = "PEXPIREAT"
	CmdTTL      = "TTL"
	CmdPTTL     = "PTTL"
	CmdPersist  = "PERSIST"

	// Transaction commands
	CmdMulti   = "MULTI"
	CmdExec    = "EXEC"
	CmdDiscard = "DISCARD"
	CmdWatch   = "WATCH"
	CmdUnwatch = "UNWATCH"

	// Management commands
	CmdPing    = "PING"
	CmdInfo    = "INFO"
	CmdMemory  = "MEMORY"
	CmdSave    = "SAVE"
	CmdBgSave  = "BGSAVE"
	CmdSlaveOf = "SLAVEOF"
	CmdSync    = "SYNC"
	CmdPSync   = "PSYNC"

	// Database commands
	CmdSelect = "SELECT"
	CmdType   = "TYPE"
	CmdMove   = "MOVE"
	CmdAuth    = "AUTH"
	CmdSlowLog = "SLOWLOG"
	CmdMonitor = "MONITOR"
)

// WriteCommands is a map of write commands (commands that modify data)
var WriteCommands = map[string]bool{
	// String commands
	CmdSet:      true,
	CmdMSet:     true,
	CmdDel:      true,
	CmdIncr:     true,
	CmdIncrBy:   true,
	CmdDecr:     true,
	CmdDecrBy:   true,
	CmdAppend:   true,
	CmdGetRange: true,

	// Hash commands
	CmdHSet:    true,
	CmdHMSet:   true,
	CmdHSetNX:  true,
	CmdHDel:    true,
	CmdHIncrBy: true,

	// List commands
	CmdLPush:   true,
	CmdRPush:   true,
	CmdLPop:    true,
	CmdRPop:    true,
	CmdLSet:    true,
	CmdLTrim:   true,
	CmdLRem:    true,
	CmdLInsert: true,

	// Set commands
	CmdSAdd:  true,
	CmdSRem:  true,
	CmdSPop:  true,
	CmdSMove: true,

	// Sorted Set commands
	CmdZAdd:    true,
	CmdZRem:    true,
	CmdZIncrBy: true,

	// TTL commands
	CmdExpire:  true,
	CmdPExpire: true,
	CmdPersist: true,
}

// IntegerCommands is a map of commands that return integer results
var IntegerCommands = map[string]bool{
	// String commands
	CmdDel:     true,
	CmdExists:  true,
	CmdIncr:    true,
	CmdIncrBy:  true,
	CmdDecr:    true,
	CmdDecrBy:  true,
	CmdStrLen:  true,
	CmdAppend:  true,

	// Hash commands
	CmdHDel:    true,
	CmdHExists: true,
	CmdHLen:    true,
	CmdHSetNX:  true,
	CmdHIncrBy: true,

	// List commands
	CmdLPush:   true,
	CmdRPush:   true,
	CmdLPop:    true,
	CmdRPop:    true,
	CmdLLen:    true,
	CmdLInsert: true,
	CmdLRem:    true,

	// Set commands
	CmdSAdd:       true,
	CmdSRem:       true,
	CmdSCard:      true,
	CmdSIsMember:  true,
	CmdSMove:      true,
	CmdSDiffStore: true,
	CmdSInterStore: true,
	CmdSUnionStore: true,

	// Sorted Set commands
	CmdZAdd:    true,
	CmdZRem:    true,
	CmdZCard:   true,
	CmdZCount:  true,
	CmdZRank:   true,
	CmdZRevRank: true,
	CmdZIncrBy: true,

	// TTL commands
	CmdExpire:  true,
	CmdPExpire: true,
	CmdPersist: true,
	CmdTTL:     true,
	CmdPTTL:    true,
}

// ArrayCommands is a map of commands that always return array replies (even with 1 element)
var ArrayCommands = map[string]bool{
	// Hash commands
	CmdHGetAll: true,
	CmdHKeys:   true,
	CmdHVals:   true,
	CmdHMGet:   true,

	// List commands
	CmdLRange: true,

	// Set commands
	CmdSMembers:  true,
	CmdSDiff:     true,
	CmdSInter:    true,
	CmdSUnion:    true,

	// Sorted Set commands
	CmdZRange:        true,
	CmdZRevRange:     true,
	CmdZRangeByScore: true,

	// String commands
	CmdKeys: true,
	CmdMGet: true,
}

// StatusCommands is a map of commands that return status "OK" response
var StatusCommands = map[string]bool{
	CmdSet:     true,
	CmdMSet:    true,
	CmdHMSet:   true,
	CmdLSet:    true,
	CmdLTrim:   true,
	CmdMulti:   true,
	CmdDiscard: true,
	CmdWatch:   true,
	CmdUnwatch: true,
	CmdSave:    true,
	CmdBgSave:  true,
	CmdSlaveOf: true,
}

// IsWriteCommand checks if a command is a write command (case-insensitive)
func IsWriteCommand(cmd string) bool {
	return WriteCommands[ToUpper(cmd)]
}

// IsIntegerCommand checks if a command returns an integer result (case-insensitive)
func IsIntegerCommand(cmd string) bool {
	return IntegerCommands[ToUpper(cmd)]
}

// IsArrayCommand checks if a command returns an array result (case-insensitive)
func IsArrayCommand(cmd string) bool {
	return ArrayCommands[ToUpper(cmd)]
}

// IsStatusCommand checks if a command returns a status "OK" response (case-insensitive)
func IsStatusCommand(cmd string) bool {
	return StatusCommands[ToUpper(cmd)]
}

// ToUpper converts a string to uppercase (case-insensitive command handling)
// This is a simple implementation - for production, consider using strings.ToUpper
func ToUpper(s string) string {
	if len(s) == 0 {
		return s
	}

	// Simple ASCII-only toUpper (faster than strings.ToUpper for our use case)
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			result[i] = c - 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}
