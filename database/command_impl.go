package database

// Command executor registry
var commandExecutors = map[CommandType]CommandExecutor{}

// BaseCommand provides a default implementation for IsWriteCommand
type BaseCommand struct {
	isWrite bool
}

func (c *BaseCommand) IsWriteCommand() bool {
	return c.isWrite
}

// FunctionCommand wraps a function as a CommandExecutor
type FunctionCommand struct {
	BaseCommand
	executeFunc func(db *DB, args [][]byte) ([][]byte, error)
}

func (c *FunctionCommand) Execute(db *DB, args [][]byte) ([][]byte, error) {
	return c.executeFunc(db, args)
}

// NewWriteCommand creates a write command executor
func NewWriteCommand(fn func(db *DB, args [][]byte) ([][]byte, error)) CommandExecutor {
	return &FunctionCommand{
		BaseCommand: BaseCommand{isWrite: true},
		executeFunc: fn,
	}
}

// NewReadCommand creates a read command executor
func NewReadCommand(fn func(db *DB, args [][]byte) ([][]byte, error)) CommandExecutor {
	return &FunctionCommand{
		BaseCommand: BaseCommand{isWrite: false},
		executeFunc: fn,
	}
}

// Initialize command executors using the existing exec functions
func initCommandExecutors() {
	// String commands
	commandExecutors[CmdSet] = NewWriteCommand(execSet)
	commandExecutors[CmdGet] = NewReadCommand(execGet)
	commandExecutors[CmdMSet] = NewWriteCommand(execMSet)
	commandExecutors[CmdMGet] = NewReadCommand(execMGet)
	commandExecutors[CmdDel] = NewWriteCommand(execDel)
	commandExecutors[CmdExists] = NewReadCommand(execExists)
	commandExecutors[CmdKeys] = NewReadCommand(execKeys)
	commandExecutors[CmdIncr] = NewWriteCommand(execIncr)
	commandExecutors[CmdIncrBy] = NewWriteCommand(execIncrBy)
	commandExecutors[CmdDecr] = NewWriteCommand(execDecr)
	commandExecutors[CmdDecrBy] = NewWriteCommand(execDecrBy)
	commandExecutors[CmdStrLen] = NewReadCommand(execStrLen)
	commandExecutors[CmdAppend] = NewWriteCommand(execAppend)
	commandExecutors[CmdGetRange] = NewReadCommand(execGetRange)

	// Hash commands
	commandExecutors[CmdHSet] = NewWriteCommand(execHSet)
	commandExecutors[CmdHGet] = NewReadCommand(execHGet)
	commandExecutors[CmdHDel] = NewWriteCommand(execHDel)
	commandExecutors[CmdHExists] = NewReadCommand(execHExists)
	commandExecutors[CmdHGetAll] = NewReadCommand(execHGetAll)
	commandExecutors[CmdHKeys] = NewReadCommand(execHKeys)
	commandExecutors[CmdHVals] = NewReadCommand(execHVals)
	commandExecutors[CmdHLen] = NewReadCommand(execHLen)
	commandExecutors[CmdHSetNX] = NewWriteCommand(execHSetNX)
	commandExecutors[CmdHIncrBy] = NewWriteCommand(execHIncrBy)
	commandExecutors[CmdHMGet] = NewReadCommand(execHMGet)
	commandExecutors[CmdHMSet] = NewWriteCommand(execHMSet)

	// List commands
	commandExecutors[CmdLPush] = NewWriteCommand(execLPush)
	commandExecutors[CmdRPush] = NewWriteCommand(execRPush)
	commandExecutors[CmdLPop] = NewWriteCommand(execLPop)
	commandExecutors[CmdRPop] = NewWriteCommand(execRPop)
	commandExecutors[CmdLIndex] = NewReadCommand(execLIndex)
	commandExecutors[CmdLSet] = NewWriteCommand(execLSet)
	commandExecutors[CmdLRange] = NewReadCommand(execLRange)
	commandExecutors[CmdLTrim] = NewWriteCommand(execLTrim)
	commandExecutors[CmdLRem] = NewWriteCommand(execLRem)
	commandExecutors[CmdLInsert] = NewWriteCommand(execLInsert)
	commandExecutors[CmdLLen] = NewReadCommand(execLLen)

	// Set commands
	commandExecutors[CmdSAdd] = NewWriteCommand(execSAdd)
	commandExecutors[CmdSRem] = NewWriteCommand(execSRem)
	commandExecutors[CmdSIsMember] = NewReadCommand(execSIsMember)
	commandExecutors[CmdSMembers] = NewReadCommand(execSMembers)
	commandExecutors[CmdSCard] = NewReadCommand(execSCard)
	commandExecutors[CmdSPop] = NewWriteCommand(execSPop)
	commandExecutors[CmdSRandMember] = NewReadCommand(execSRandMember)
	commandExecutors[CmdSMove] = NewWriteCommand(execSMove)
	commandExecutors[CmdSDiff] = NewReadCommand(execSDiff)
	commandExecutors[CmdSInter] = NewReadCommand(execSInter)
	commandExecutors[CmdSUnion] = NewReadCommand(execSUnion)
	commandExecutors[CmdSDiffStore] = NewWriteCommand(execSDiffStore)
	commandExecutors[CmdSInterStore] = NewWriteCommand(execSInterStore)
	commandExecutors[CmdSUnionStore] = NewWriteCommand(execSUnionStore)

	// Sorted Set commands
	commandExecutors[CmdZAdd] = NewWriteCommand(execZAdd)
	commandExecutors[CmdZRem] = NewWriteCommand(execZRem)
	commandExecutors[CmdZScore] = NewReadCommand(execZScore)
	commandExecutors[CmdZIncrBy] = NewWriteCommand(execZIncrBy)
	commandExecutors[CmdZCard] = NewReadCommand(execZCard)
	commandExecutors[CmdZRank] = NewReadCommand(execZRank)
	commandExecutors[CmdZRevRank] = NewReadCommand(execZRevRank)
	commandExecutors[CmdZRange] = NewReadCommand(execZRange)
	commandExecutors[CmdZRevRange] = NewReadCommand(execZRevRange)
	commandExecutors[CmdZRangeByScore] = NewReadCommand(execZRangeByScore)
	commandExecutors[CmdZCount] = NewReadCommand(execZCount)

	// TTL commands
	commandExecutors[CmdExpire] = NewWriteCommand(execExpire)
	commandExecutors[CmdPExpire] = NewWriteCommand(execPExpire)
	commandExecutors[CmdExpireAt] = NewWriteCommand(execExpireAt)
	commandExecutors[CmdPExpireAt] = NewWriteCommand(execPExpireAt)
	commandExecutors[CmdTTL] = NewReadCommand(execTTL)
	commandExecutors[CmdPTTL] = NewReadCommand(execPTTL)
	commandExecutors[CmdPersist] = NewWriteCommand(execPersist)

	// Transaction commands
	commandExecutors[CmdMulti] = NewReadCommand(execMulti)
	commandExecutors[CmdExec] = NewReadCommand(execExec)
	commandExecutors[CmdDiscard] = NewReadCommand(execDiscard)
	commandExecutors[CmdWatch] = NewReadCommand(execWatch)
	commandExecutors[CmdUnwatch] = NewReadCommand(execUnwatch)

	// Management commands
	commandExecutors[CmdPing] = NewReadCommand(execPing)
	commandExecutors[CmdInfo] = NewReadCommand(execInfo)
	commandExecutors[CmdMemory] = NewReadCommand(execMemory)
	commandExecutors[CmdSave] = NewReadCommand(execSave)
	commandExecutors[CmdBgSave] = NewReadCommand(execBgSave)
	commandExecutors[CmdSlaveOf] = NewReadCommand(execSlaveOf)
	commandExecutors[CmdSync] = NewReadCommand(execSync)
	commandExecutors[CmdPSync] = NewReadCommand(execPSync)

	// Database commands
	commandExecutors[CmdSelect] = NewReadCommand(execSelect)
	commandExecutors[CmdType] = NewReadCommand(execType)
	commandExecutors[CmdMove] = NewWriteCommand(execMove)

	// Security and monitoring commands
	commandExecutors[CmdAuth] = NewReadCommand(execAuth)
	commandExecutors[CmdSlowLog] = NewReadCommand(execSlowLog)
	commandExecutors[CmdMonitor] = NewReadCommand(execMonitor)
}

func init() {
	initCommandExecutors()
}

// GetCommandExecutor returns the executor for a given command type
func GetCommandExecutor(cmdType CommandType) (CommandExecutor, bool) {
	executor, ok := commandExecutors[cmdType]
	return executor, ok
}

// RegisterCommandExecutor allows registering custom command executors
func RegisterCommandExecutor(cmdType CommandType, executor CommandExecutor) {
	commandExecutors[cmdType] = executor
}
