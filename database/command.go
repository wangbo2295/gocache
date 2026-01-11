package database

import (
	"github.com/wangbo/gocache/protocol"
)

// CommandExecutor is the interface for command handlers
// Each command implements this interface to handle its execution logic
type CommandExecutor interface {
	// Execute runs the command with given arguments
	// Args are the command arguments (not including the command name itself)
	Execute(db *DB, args [][]byte) ([][]byte, error)

	// IsWriteCommand returns true if this command modifies data
	IsWriteCommand() bool
}

// CommandType represents a command type enumeration
type CommandType int

const (
	// String commands
	CmdSet CommandType = iota
	CmdGet
	CmdMSet
	CmdMGet
	CmdDel
	CmdExists
	CmdKeys
	CmdIncr
	CmdIncrBy
	CmdDecr
	CmdDecrBy
	CmdStrLen
	CmdAppend
	CmdGetRange

	// Hash commands
	CmdHSet
	CmdHGet
	CmdHDel
	CmdHExists
	CmdHGetAll
	CmdHKeys
	CmdHVals
	CmdHLen
	CmdHSetNX
	CmdHIncrBy
	CmdHMGet
	CmdHMSet

	// List commands
	CmdLPush
	CmdRPush
	CmdLPop
	CmdRPop
	CmdLIndex
	CmdLSet
	CmdLRange
	CmdLTrim
	CmdLRem
	CmdLInsert
	CmdLLen

	// Set commands
	CmdSAdd
	CmdSRem
	CmdSIsMember
	CmdSMembers
	CmdSCard
	CmdSPop
	CmdSRandMember
	CmdSMove
	CmdSDiff
	CmdSInter
	CmdSUnion
	CmdSDiffStore
	CmdSInterStore
	CmdSUnionStore

	// Sorted Set commands
	CmdZAdd
	CmdZRem
	CmdZScore
	CmdZIncrBy
	CmdZCard
	CmdZRank
	CmdZRevRank
	CmdZRange
	CmdZRevRange
	CmdZRangeByScore
	CmdZCount

	// TTL commands
	CmdExpire
	CmdPExpire
	CmdExpireAt
	CmdPExpireAt
	CmdTTL
	CmdPTTL
	CmdPersist

	// Transaction commands
	CmdMulti
	CmdExec
	CmdDiscard
	CmdWatch
	CmdUnwatch

	// Management commands
	CmdPing
	CmdInfo
	CmdMemory
	CmdSave
	CmdBgSave
	CmdSlaveOf
	CmdSync
	CmdPSync

	// Database commands
	CmdSelect
	CmdType
	CmdMove

	// Security and monitoring commands
	CmdAuth
	CmdSlowLog
	CmdMonitor
)

// String returns the string representation of the command type
func (c CommandType) String() string {
	switch c {
	case CmdSet:
		return protocol.CmdSet
	case CmdGet:
		return protocol.CmdGet
	case CmdMSet:
		return protocol.CmdMSet
	case CmdMGet:
		return protocol.CmdMGet
	case CmdDel:
		return protocol.CmdDel
	case CmdExists:
		return protocol.CmdExists
	case CmdKeys:
		return protocol.CmdKeys
	case CmdIncr:
		return protocol.CmdIncr
	case CmdIncrBy:
		return protocol.CmdIncrBy
	case CmdDecr:
		return protocol.CmdDecr
	case CmdDecrBy:
		return protocol.CmdDecrBy
	case CmdStrLen:
		return protocol.CmdStrLen
	case CmdAppend:
		return protocol.CmdAppend
	case CmdGetRange:
		return protocol.CmdGetRange
	case CmdHSet:
		return protocol.CmdHSet
	case CmdHGet:
		return protocol.CmdHGet
	case CmdHDel:
		return protocol.CmdHDel
	case CmdHExists:
		return protocol.CmdHExists
	case CmdHGetAll:
		return protocol.CmdHGetAll
	case CmdHKeys:
		return protocol.CmdHKeys
	case CmdHVals:
		return protocol.CmdHVals
	case CmdHLen:
		return protocol.CmdHLen
	case CmdHSetNX:
		return protocol.CmdHSetNX
	case CmdHIncrBy:
		return protocol.CmdHIncrBy
	case CmdHMGet:
		return protocol.CmdHMGet
	case CmdHMSet:
		return protocol.CmdHMSet
	case CmdLPush:
		return protocol.CmdLPush
	case CmdRPush:
		return protocol.CmdRPush
	case CmdLPop:
		return protocol.CmdLPop
	case CmdRPop:
		return protocol.CmdRPop
	case CmdLIndex:
		return protocol.CmdLIndex
	case CmdLSet:
		return protocol.CmdLSet
	case CmdLRange:
		return protocol.CmdLRange
	case CmdLTrim:
		return protocol.CmdLTrim
	case CmdLRem:
		return protocol.CmdLRem
	case CmdLInsert:
		return protocol.CmdLInsert
	case CmdLLen:
		return protocol.CmdLLen
	case CmdSAdd:
		return protocol.CmdSAdd
	case CmdSRem:
		return protocol.CmdSRem
	case CmdSIsMember:
		return protocol.CmdSIsMember
	case CmdSMembers:
		return protocol.CmdSMembers
	case CmdSCard:
		return protocol.CmdSCard
	case CmdSPop:
		return protocol.CmdSPop
	case CmdSRandMember:
		return protocol.CmdSRandMember
	case CmdSMove:
		return protocol.CmdSMove
	case CmdSDiff:
		return protocol.CmdSDiff
	case CmdSInter:
		return protocol.CmdSInter
	case CmdSUnion:
		return protocol.CmdSUnion
	case CmdSDiffStore:
		return protocol.CmdSDiffStore
	case CmdSInterStore:
		return protocol.CmdSInterStore
	case CmdSUnionStore:
		return protocol.CmdSUnionStore
	case CmdZAdd:
		return protocol.CmdZAdd
	case CmdZRem:
		return protocol.CmdZRem
	case CmdZScore:
		return protocol.CmdZScore
	case CmdZIncrBy:
		return protocol.CmdZIncrBy
	case CmdZCard:
		return protocol.CmdZCard
	case CmdZRank:
		return protocol.CmdZRank
	case CmdZRevRank:
		return protocol.CmdZRevRank
	case CmdZRange:
		return protocol.CmdZRange
	case CmdZRevRange:
		return protocol.CmdZRevRange
	case CmdZRangeByScore:
		return protocol.CmdZRangeByScore
	case CmdZCount:
		return protocol.CmdZCount
	case CmdExpire:
		return protocol.CmdExpire
	case CmdPExpire:
		return protocol.CmdPExpire
	case CmdExpireAt:
		return protocol.CmdExpireAt
	case CmdPExpireAt:
		return protocol.CmdPExpireAt
	case CmdTTL:
		return protocol.CmdTTL
	case CmdPTTL:
		return protocol.CmdPTTL
	case CmdPersist:
		return protocol.CmdPersist
	case CmdMulti:
		return protocol.CmdMulti
	case CmdExec:
		return protocol.CmdExec
	case CmdDiscard:
		return protocol.CmdDiscard
	case CmdWatch:
		return protocol.CmdWatch
	case CmdUnwatch:
		return protocol.CmdUnwatch
	case CmdPing:
		return protocol.CmdPing
	case CmdInfo:
		return protocol.CmdInfo
	case CmdMemory:
		return protocol.CmdMemory
	case CmdSave:
		return protocol.CmdSave
	case CmdBgSave:
		return protocol.CmdBgSave
	case CmdSlaveOf:
		return protocol.CmdSlaveOf
	case CmdSync:
		return protocol.CmdSync
	case CmdPSync:
		return protocol.CmdPSync
	case CmdSelect:
		return protocol.CmdSelect
	case CmdType:
		return protocol.CmdType
	case CmdMove:
		return protocol.CmdMove
	case CmdAuth:
		return protocol.CmdAuth
	case CmdSlowLog:
		return protocol.CmdSlowLog
	case CmdMonitor:
		return protocol.CmdMonitor
	default:
		return "UNKNOWN"
	}
}

// IsWriteCommand returns true if this command modifies data
func (c CommandType) IsWriteCommand() bool {
	return protocol.IsWriteCommand(c.String())
}

// CommandRegistry maps command names to their types
var CommandRegistry = map[string]CommandType{
	// String commands
	protocol.CmdSet:      CmdSet,
	protocol.CmdGet:      CmdGet,
	protocol.CmdMSet:     CmdMSet,
	protocol.CmdMGet:     CmdMGet,
	protocol.CmdDel:      CmdDel,
	protocol.CmdExists:   CmdExists,
	protocol.CmdKeys:     CmdKeys,
	protocol.CmdIncr:     CmdIncr,
	protocol.CmdIncrBy:   CmdIncrBy,
	protocol.CmdDecr:     CmdDecr,
	protocol.CmdDecrBy:   CmdDecrBy,
	protocol.CmdStrLen:   CmdStrLen,
	protocol.CmdAppend:   CmdAppend,
	protocol.CmdGetRange: CmdGetRange,

	// Hash commands
	protocol.CmdHSet:    CmdHSet,
	protocol.CmdHGet:    CmdHGet,
	protocol.CmdHDel:    CmdHDel,
	protocol.CmdHExists: CmdHExists,
	protocol.CmdHGetAll: CmdHGetAll,
	protocol.CmdHKeys:   CmdHKeys,
	protocol.CmdHVals:   CmdHVals,
	protocol.CmdHLen:    CmdHLen,
	protocol.CmdHSetNX:  CmdHSetNX,
	protocol.CmdHIncrBy: CmdHIncrBy,
	protocol.CmdHMGet:   CmdHMGet,
	protocol.CmdHMSet:   CmdHMSet,

	// List commands
	protocol.CmdLPush:   CmdLPush,
	protocol.CmdRPush:   CmdRPush,
	protocol.CmdLPop:    CmdLPop,
	protocol.CmdRPop:    CmdRPop,
	protocol.CmdLIndex:  CmdLIndex,
	protocol.CmdLSet:    CmdLSet,
	protocol.CmdLRange:  CmdLRange,
	protocol.CmdLTrim:   CmdLTrim,
	protocol.CmdLRem:    CmdLRem,
	protocol.CmdLInsert: CmdLInsert,
	protocol.CmdLLen:    CmdLLen,

	// Set commands
	protocol.CmdSAdd:        CmdSAdd,
	protocol.CmdSRem:        CmdSRem,
	protocol.CmdSIsMember:   CmdSIsMember,
	protocol.CmdSMembers:    CmdSMembers,
	protocol.CmdSCard:       CmdSCard,
	protocol.CmdSPop:        CmdSPop,
	protocol.CmdSRandMember: CmdSRandMember,
	protocol.CmdSMove:       CmdSMove,
	protocol.CmdSDiff:       CmdSDiff,
	protocol.CmdSInter:      CmdSInter,
	protocol.CmdSUnion:      CmdSUnion,
	protocol.CmdSDiffStore:  CmdSDiffStore,
	protocol.CmdSInterStore: CmdSInterStore,
	protocol.CmdSUnionStore: CmdSUnionStore,

	// Sorted Set commands
	protocol.CmdZAdd:          CmdZAdd,
	protocol.CmdZRem:          CmdZRem,
	protocol.CmdZScore:        CmdZScore,
	protocol.CmdZIncrBy:       CmdZIncrBy,
	protocol.CmdZCard:         CmdZCard,
	protocol.CmdZRank:         CmdZRank,
	protocol.CmdZRevRank:      CmdZRevRank,
	protocol.CmdZRange:        CmdZRange,
	protocol.CmdZRevRange:     CmdZRevRange,
	protocol.CmdZRangeByScore: CmdZRangeByScore,
	protocol.CmdZCount:        CmdZCount,

	// TTL commands
	protocol.CmdExpire:    CmdExpire,
	protocol.CmdPExpire:   CmdPExpire,
	protocol.CmdExpireAt:  CmdExpireAt,
	protocol.CmdPExpireAt: CmdPExpireAt,
	protocol.CmdTTL:       CmdTTL,
	protocol.CmdPTTL:      CmdPTTL,
	protocol.CmdPersist:   CmdPersist,

	// Transaction commands
	protocol.CmdMulti:   CmdMulti,
	protocol.CmdExec:    CmdExec,
	protocol.CmdDiscard: CmdDiscard,
	protocol.CmdWatch:   CmdWatch,
	protocol.CmdUnwatch: CmdUnwatch,

	// Management commands
	protocol.CmdPing:    CmdPing,
	protocol.CmdInfo:    CmdInfo,
	protocol.CmdMemory:  CmdMemory,
	protocol.CmdSave:    CmdSave,
	protocol.CmdBgSave:  CmdBgSave,
	protocol.CmdSlaveOf: CmdSlaveOf,
	protocol.CmdSync:    CmdSync,
	protocol.CmdPSync:   CmdPSync,

	// Database commands
	protocol.CmdSelect: CmdSelect,
	protocol.CmdType:   CmdType,
	protocol.CmdMove:   CmdMove,

	// Security and monitoring commands
	protocol.CmdAuth:    CmdAuth,
	protocol.CmdSlowLog: CmdSlowLog,
	protocol.CmdMonitor: CmdMonitor,
}

// ParseCommandType parses a command name string to CommandType
func ParseCommandType(cmdName string) (CommandType, bool) {
	// Convert to uppercase for case-insensitive lookup
	cmdUpper := protocol.ToUpper(cmdName)
	cmdType, ok := CommandRegistry[cmdUpper]
	return cmdType, ok
}
