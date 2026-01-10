package database

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wangbo/gocache/datastruct"
	"github.com/wangbo/gocache/dict"
)

// DB represents a single database instance
type DB struct {
	index      int
	data       *dict.ConcurrentDict
	ttlMap     *dict.ConcurrentDict
	versionMap *dict.ConcurrentDict
	mu         sync.RWMutex
}

// MakeDB creates a new database instance
func MakeDB() *DB {
	return &DB{
		index:      0,
		data:       dict.MakeConcurrentDict(16),
		ttlMap:     dict.MakeConcurrentDict(16),
		versionMap: dict.MakeConcurrentDict(16),
	}
}

// Exec executes a command and returns a reply
func (db *DB) Exec(cmdLine [][]byte) (result [][]byte, err error) {
	if len(cmdLine) == 0 {
		return nil, errors.New("empty command")
	}

	cmd := strings.ToLower(string(cmdLine[0]))
	args := cmdLine[1:]

	switch cmd {
	case "set":
		return execSet(db, args)
	case "get":
		return execGet(db, args)
	case "del":
		return execDel(db, args)
	case "exists":
		return execExists(db, args)
	case "keys":
		return execKeys(db, args)
	case "incr":
		return execIncr(db, args)
	case "incrby":
		return execIncrBy(db, args)
	case "decr":
		return execDecr(db, args)
	case "decrby":
		return execDecrBy(db, args)
	case "mget":
		return execMGet(db, args)
	case "mset":
		return execMSet(db, args)
	case "strlen":
		return execStrLen(db, args)
	case "append":
		return execAppend(db, args)
	case "getrange":
		return execGetRange(db, args)
	case "expire":
		return execExpire(db, args)
	case "pexpire":
		return execPExpire(db, args)
	case "ttl":
		return execTTL(db, args)
	case "pttl":
		return execPTTL(db, args)
	case "persist":
		return execPersist(db, args)
	// Hash commands
	case "hset":
		return execHSet(db, args)
	case "hget":
		return execHGet(db, args)
	case "hdel":
		return execHDel(db, args)
	case "hexists":
		return execHExists(db, args)
	case "hgetall":
		return execHGetAll(db, args)
	case "hkeys":
		return execHKeys(db, args)
	case "hvals":
		return execHVals(db, args)
	case "hlen":
		return execHLen(db, args)
	case "hsetnx":
		return execHSetNX(db, args)
	case "hincrby":
		return execHIncrBy(db, args)
	case "hmget":
		return execHMGet(db, args)
	case "hmset":
		return execHMSet(db, args)
	// List commands
	case "lpush":
		return execLPush(db, args)
	case "rpush":
		return execRPush(db, args)
	case "lpop":
		return execLPop(db, args)
	case "rpop":
		return execRPop(db, args)
	case "lindex":
		return execLIndex(db, args)
	case "lset":
		return execLSet(db, args)
	case "lrange":
		return execLRange(db, args)
	case "ltrim":
		return execLTrim(db, args)
	case "lrem":
		return execLRem(db, args)
	case "linsert":
		return execLInsert(db, args)
	case "llen":
		return execLLen(db, args)
	// Set commands
	case "sadd":
		return execSAdd(db, args)
	case "srem":
		return execSRem(db, args)
	case "sismember":
		return execSIsMember(db, args)
	case "smembers":
		return execSMembers(db, args)
	case "scard":
		return execSCard(db, args)
	case "spop":
		return execSPop(db, args)
	case "srandmember":
		return execSRandMember(db, args)
	case "smove":
		return execSMove(db, args)
	case "sdiff":
		return execSDiff(db, args)
	case "sinter":
		return execSInter(db, args)
	case "sunion":
		return execSUnion(db, args)
	case "sdiffstore":
		return execSDiffStore(db, args)
	case "sinterstore":
		return execSInterStore(db, args)
	case "sunionstore":
		return execSUnionStore(db, args)
	// SortedSet commands
	case "zadd":
		return execZAdd(db, args)
	case "zrem":
		return execZRem(db, args)
	case "zscore":
		return execZScore(db, args)
	case "zincrby":
		return execZIncrBy(db, args)
	case "zcard":
		return execZCard(db, args)
	case "zrank":
		return execZRank(db, args)
	case "zrevrank":
		return execZRevRank(db, args)
	case "zrange":
		return execZRange(db, args)
	case "zrevrange":
		return execZRevRange(db, args)
	case "zrangebyscore":
		return execZRangeByScore(db, args)
	case "zcount":
		return execZCount(db, args)
	default:
		return nil, errors.New("unknown command: " + cmd)
	}
}

// GetEntity retrieves the data entity for a given key
// It checks TTL and removes expired keys automatically
func (db *DB) GetEntity(key string) (*datastruct.DataEntity, bool) {
	// Check if key is expired
	db.expireIfNeeded(key)

	val, ok := db.data.Get(key)
	if !ok {
		return nil, false
	}
	entity, ok := val.(*datastruct.DataEntity)
	if !ok {
		return nil, false
	}
	return entity, true
}

// PutEntity stores a data entity
func (db *DB) PutEntity(key string, entity *datastruct.DataEntity) int {
	return db.data.Put(key, entity)
}

// PutIfExists updates entity only if key exists
func (db *DB) PutIfExists(key string, entity *datastruct.DataEntity) int {
	return db.data.PutIfExists(key, entity)
}

// PutIfAbsent inserts entity only if key does not exist
func (db *DB) PutIfAbsent(key string, entity *datastruct.DataEntity) int {
	return db.data.PutIfAbsent(key, entity)
}

// Remove removes a key from the database
func (db *DB) Remove(key string) int {
	result := db.data.Remove(key)
	db.ttlMap.Remove(key)
	db.versionMap.Remove(key)
	return result
}

// Exists checks if a key exists
func (db *DB) Exists(key string) bool {
	db.expireIfNeeded(key)
	_, ok := db.data.Get(key)
	return ok
}

// Expire sets a TTL for a key (in seconds)
func (db *DB) Expire(key string, ttl time.Duration) int {
	if _, ok := db.data.Get(key); !ok {
		return 0
	}
	db.ttlMap.Put(key, time.Now().Add(ttl))
	return 1
}

// Persist removes the TTL from a key
func (db *DB) Persist(key string) int {
	if _, ok := db.ttlMap.Get(key); !ok {
		return 0
	}
	db.ttlMap.Remove(key)
	return 1
}

// TTL returns the remaining TTL in seconds
// Returns -2 if key does not exist, -1 if key exists but has no expiry
func (db *DB) TTL(key string) time.Duration {
	db.expireIfNeeded(key)

	if _, ok := db.data.Get(key); !ok {
		return -2
	}

	val, ok := db.ttlMap.Get(key)
	if !ok {
		return -1
	}

	expireTime := val.(time.Time)
	remaining := time.Until(expireTime)
	if remaining < 0 {
		// Already expired
		db.Remove(key)
		return -2
	}

	return remaining
}

// expireIfNeeded checks and removes expired key
func (db *DB) expireIfNeeded(key string) {
	val, ok := db.ttlMap.Get(key)
	if !ok {
		return
	}

	expireTime := val.(time.Time)
	if time.Now().After(expireTime) {
		db.Remove(key)
	}
}

// ExecCommand is a convenience method to execute command from strings
func (db *DB) ExecCommand(cmd string, args ...string) ([][]byte, error) {
	cmdLine := make([][]byte, 0, len(args)+1)
	cmdLine = append(cmdLine, []byte(cmd))
	for _, arg := range args {
		cmdLine = append(cmdLine, []byte(arg))
	}
	return db.Exec(cmdLine)
}

// Command implementations

func execSet(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("wrong number of arguments")
	}

	key := string(args[0])
	value := args[1]

	// Simple SET without options for now
	entity := datastruct.MakeString(value)
	db.PutEntity(key, entity)

	return [][]byte{[]byte("OK")}, nil
}

func execGet(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments")
	}

	key := string(args[0])
	entity, ok := db.GetEntity(key)
	if !ok {
		return [][]byte{nil}, nil
	}

	str, ok := entity.Data.(*datastruct.String)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	return [][]byte{str.Get()}, nil
}

func execDel(db *DB, args [][]byte) ([][]byte, error) {
	count := 0
	for _, arg := range args {
		key := string(arg)
		count += db.Remove(key)
	}
	return [][]byte{[]byte(strconv.Itoa(count))}, nil
}

func execExists(db *DB, args [][]byte) ([][]byte, error) {
	count := 0
	for _, arg := range args {
		key := string(arg)
		if db.Exists(key) {
			count++
		}
	}
	return [][]byte{[]byte(strconv.Itoa(count))}, nil
}

func execKeys(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments")
	}

	// Simple implementation: return all keys
	// TODO: support pattern matching
	keys := db.data.Keys()
	result := make([][]byte, len(keys))
	for i, key := range keys {
		result[i] = []byte(key)
	}
	return result, nil
}

func execIncr(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments")
	}

	key := string(args[0])
	entity, ok := db.GetEntity(key)
	var str *datastruct.String

	if ok {
		var ok2 bool
		str, ok2 = entity.Data.(*datastruct.String)
		if !ok2 {
			return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
	} else {
		// Key doesn't exist, start from 0
		str = &datastruct.String{Value: []byte("0")}
		entity = &datastruct.DataEntity{Data: str}
		db.PutEntity(key, entity)
	}

	newVal, err := str.Increment(1)
	if err != nil {
		return nil, err
	}

	return [][]byte{[]byte(strconv.FormatInt(newVal, 10))}, nil
}

func execIncrBy(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("wrong number of arguments")
	}

	key := string(args[0])
	delta, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return nil, errors.New("ERR value is not an integer or out of range")
	}

	entity, ok := db.GetEntity(key)
	var str *datastruct.String

	if ok {
		var ok2 bool
		str, ok2 = entity.Data.(*datastruct.String)
		if !ok2 {
			return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
	} else {
		str = &datastruct.String{Value: []byte("0")}
		entity = &datastruct.DataEntity{Data: str}
		db.PutEntity(key, entity)
	}

	newVal, err := str.Increment(delta)
	if err != nil {
		return nil, err
	}

	return [][]byte{[]byte(strconv.FormatInt(newVal, 10))}, nil
}

func execDecr(db *DB, args [][]byte) ([][]byte, error) {
	return execIncrBy(db, append([][]byte{args[0]}, []byte("-1")))
}

func execDecrBy(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("wrong number of arguments")
	}

	delta, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return nil, errors.New("ERR value is not an integer or out of range")
	}

	// Negate the delta for INCRBY
	negDelta := -delta
	args[1] = []byte(strconv.FormatInt(negDelta, 10))

	return execIncrBy(db, args)
}

func execMGet(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 1 {
		return nil, errors.New("wrong number of arguments")
	}

	result := make([][]byte, len(args))
	for i, arg := range args {
		key := string(arg)
		entity, ok := db.GetEntity(key)
		if !ok {
			result[i] = nil
			continue
		}

		str, ok := entity.Data.(*datastruct.String)
		if !ok {
			result[i] = nil
			continue
		}

		result[i] = str.Get()
	}
	return result, nil
}

func execMSet(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 2 || len(args)%2 != 0 {
		return nil, errors.New("wrong number of arguments")
	}

	for i := 0; i < len(args); i += 2 {
		key := string(args[i])
		value := args[i+1]
		entity := datastruct.MakeString(value)
		db.PutEntity(key, entity)
	}

	return [][]byte{[]byte("OK")}, nil
}

func execStrLen(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments")
	}

	key := string(args[0])
	entity, ok := db.GetEntity(key)
	if !ok {
		return [][]byte{[]byte("0")}, nil
	}

	str, ok := entity.Data.(*datastruct.String)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	length := str.StrLen()
	return [][]byte{[]byte(strconv.Itoa(length))}, nil
}

func execAppend(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("wrong number of arguments")
	}

	key := string(args[0])
	value := args[1]

	entity, ok := db.GetEntity(key)
	var str *datastruct.String

	if ok {
		var ok2 bool
		str, ok2 = entity.Data.(*datastruct.String)
		if !ok2 {
			return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
	} else {
		str = &datastruct.String{}
		entity = &datastruct.DataEntity{Data: str}
		db.PutEntity(key, entity)
	}

	newLen := str.Append(value)
	return [][]byte{[]byte(strconv.Itoa(newLen))}, nil
}

func execGetRange(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 3 {
		return nil, errors.New("wrong number of arguments")
	}

	key := string(args[0])
	start, err := strconv.Atoi(string(args[1]))
	if err != nil {
		return nil, errors.New("ERR value is not an integer or out of range")
	}
	end, err := strconv.Atoi(string(args[2]))
	if err != nil {
		return nil, errors.New("ERR value is not an integer or out of range")
	}

	entity, ok := db.GetEntity(key)
	if !ok {
		return [][]byte{[]byte("")}, nil
	}

	str, ok := entity.Data.(*datastruct.String)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	result := str.GetRange(start, end)
	return [][]byte{result}, nil
}

func execExpire(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("wrong number of arguments")
	}

	key := string(args[0])
	seconds, err := strconv.Atoi(string(args[1]))
	if err != nil {
		return nil, errors.New("ERR value is not an integer or out of range")
	}

	ttl := time.Duration(seconds) * time.Second
	result := db.Expire(key, ttl)
	return [][]byte{[]byte(strconv.Itoa(result))}, nil
}

func execPExpire(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("wrong number of arguments")
	}

	key := string(args[0])
	milliseconds, err := strconv.Atoi(string(args[1]))
	if err != nil {
		return nil, errors.New("ERR value is not an integer or out of range")
	}

	ttl := time.Duration(milliseconds) * time.Millisecond
	result := db.Expire(key, ttl)
	return [][]byte{[]byte(strconv.Itoa(result))}, nil
}

func execTTL(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments")
	}

	key := string(args[0])
	ttl := db.TTL(key)

	// Handle special cases
	if ttl == -2 {
		return [][]byte{[]byte("-2")}, nil
	}
	if ttl == -1 {
		return [][]byte{[]byte("-1")}, nil
	}

	// Convert to seconds
	seconds := int64(ttl.Seconds())
	return [][]byte{[]byte(strconv.FormatInt(seconds, 10))}, nil
}

func execPTTL(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments")
	}

	key := string(args[0])
	ttl := db.TTL(key)

	// Handle special cases
	if ttl == -2 {
		return [][]byte{[]byte("-2")}, nil
	}
	if ttl == -1 {
		return [][]byte{[]byte("-1")}, nil
	}

	// Convert to milliseconds
	milliseconds := int64(ttl.Milliseconds())
	return [][]byte{[]byte(strconv.FormatInt(milliseconds, 10))}, nil
}

func execPersist(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments")
	}

	key := string(args[0])
	result := db.Persist(key)
	return [][]byte{[]byte(strconv.Itoa(result))}, nil
}

// ==================== Hash Commands ====================

func execHSet(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 3 {
		return nil, errors.New("wrong number of arguments")
	}

	key := string(args[0])
	field := string(args[1])
	value := args[2]

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		entity = datastruct.MakeHash()
	}

	hash, ok := entity.Data.(*datastruct.Hash)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	hash.Set(field, value)
	db.PutEntity(key, entity)
	return [][]byte{[]byte("1")}, nil
}

func execHGet(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("wrong number of arguments")
	}

	key := string(args[0])
	field := string(args[1])

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{nil}, nil
	}

	hash, ok := entity.Data.(*datastruct.Hash)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	val, ok := hash.Get(field)
	if !ok {
		return [][]byte{nil}, nil
	}
	return [][]byte{val}, nil
}

func execHDel(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("wrong number of arguments")
	}

	key := string(args[0])
	fields := make([]string, 0, len(args)-1)
	for i := 1; i < len(args); i++ {
		fields = append(fields, string(args[i]))
	}

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{[]byte("0")}, nil
	}

	hash, ok := entity.Data.(*datastruct.Hash)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	count := hash.Remove(fields...)
	if count > 0 {
		if hash.Len() == 0 {
			db.Remove(key)
		} else {
			db.PutEntity(key, entity)
		}
	}
	return [][]byte{[]byte(strconv.Itoa(count))}, nil
}

func execHExists(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("wrong number of arguments")
	}

	key := string(args[0])
	field := string(args[1])

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{[]byte("0")}, nil
	}

	hash, ok := entity.Data.(*datastruct.Hash)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	if hash.Exists(field) {
		return [][]byte{[]byte("1")}, nil
	}
	return [][]byte{[]byte("0")}, nil
}

func execHGetAll(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments")
	}

	key := string(args[0])

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{}, nil
	}

	hash, ok := entity.Data.(*datastruct.Hash)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	all := hash.GetAll()
	result := make([][]byte, 0, len(all)*2)
	for k, v := range all {
		result = append(result, []byte(k))
		result = append(result, v)
	}
	return result, nil
}

func execHKeys(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments")
	}

	key := string(args[0])

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{}, nil
	}

	hash, ok := entity.Data.(*datastruct.Hash)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	keys := hash.Keys()
	result := make([][]byte, len(keys))
	for i, key := range keys {
		result[i] = []byte(key)
	}
	return result, nil
}

func execHVals(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments")
	}

	key := string(args[0])

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{}, nil
	}

	hash, ok := entity.Data.(*datastruct.Hash)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	values := hash.Values()
	return values, nil
}

func execHLen(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments")
	}

	key := string(args[0])

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{[]byte("0")}, nil
	}

	hash, ok := entity.Data.(*datastruct.Hash)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	return [][]byte{[]byte(strconv.Itoa(hash.Len()))}, nil
}

func execHSetNX(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 3 {
		return nil, errors.New("wrong number of arguments")
	}

	key := string(args[0])
	field := string(args[1])
	value := args[2]

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		entity = datastruct.MakeHash()
	}

	hash, ok := entity.Data.(*datastruct.Hash)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	if hash.SetNX(field, value) {
		db.PutEntity(key, entity)
		return [][]byte{[]byte("1")}, nil
	}
	return [][]byte{[]byte("0")}, nil
}

func execHIncrBy(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 3 {
		return nil, errors.New("wrong number of arguments")
	}

	key := string(args[0])
	field := string(args[1])
	increment, err := strconv.ParseInt(string(args[2]), 10, 64)
	if err != nil {
		return nil, errors.New("ERR value is not an integer or out of range")
	}

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		entity = datastruct.MakeHash()
	}

	hash, ok := entity.Data.(*datastruct.Hash)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	val, err := hash.IncrBy(field, increment)
	if err != nil {
		return nil, err
	}

	db.PutEntity(key, entity)
	return [][]byte{[]byte(strconv.FormatInt(val, 10))}, nil
}

func execHMGet(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("wrong number of arguments")
	}

	key := string(args[0])
	fields := make([]string, 0, len(args)-1)
	for i := 1; i < len(args); i++ {
		fields = append(fields, string(args[i]))
	}

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		// Return nil values for all fields
		result := make([][]byte, len(fields))
		return result, nil
	}

	hash, ok := entity.Data.(*datastruct.Hash)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	result := make([][]byte, len(fields))
	for i, field := range fields {
		val, ok := hash.Get(field)
		if !ok {
			result[i] = nil
		} else {
			result[i] = val
		}
	}
	return result, nil
}

func execHMSet(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 3 || (len(args)-1)%2 != 0 {
		return nil, errors.New("wrong number of arguments")
	}

	key := string(args[0])

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		entity = datastruct.MakeHash()
	}

	hash, ok := entity.Data.(*datastruct.Hash)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	// Parse field-value pairs
	for i := 1; i < len(args); i += 2 {
		field := string(args[i])
		value := args[i+1]
		hash.Set(field, value)
	}

	db.PutEntity(key, entity)
	return [][]byte{[]byte("OK")}, nil
}

// List command implementations

// execLPush inserts one or more values at the head of the list
// LPUSH key value [value ...]
func execLPush(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("wrong number of arguments for LPUSH")
	}

	key := string(args[0])
	values := args[1:]

	// Get or create entity
	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		entity = datastruct.MakeList()
	}

	// Type check
	list, ok := entity.Data.(*datastruct.List)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	// Push values
	length := list.LPush(values...)
	db.PutEntity(key, entity)

	return [][]byte{[]byte(strconv.FormatInt(int64(length), 10))}, nil
}

// execRPush inserts one or more values at the tail of the list
// RPUSH key value [value ...]
func execRPush(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("wrong number of arguments for RPUSH")
	}

	key := string(args[0])
	values := args[1:]

	// Get or create entity
	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		entity = datastruct.MakeList()
	}

	// Type check
	list, ok := entity.Data.(*datastruct.List)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	// Push values
	length := list.RPush(values...)
	db.PutEntity(key, entity)

	return [][]byte{[]byte(strconv.FormatInt(int64(length), 10))}, nil
}

// execLPop removes and returns the first element of the list
// LPOP key
func execLPop(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments for LPOP")
	}

	key := string(args[0])

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{[]byte("(nil)")}, nil
	}

	// Type check
	list, ok := entity.Data.(*datastruct.List)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	// Pop value
	value := list.LPop()
	if value == nil {
		// List became empty, remove key
		db.Remove(key)
		return [][]byte{[]byte("(nil)")}, nil
	}

	// Update database
	if list.Len() == 0 {
		db.Remove(key)
	} else {
		db.PutEntity(key, entity)
	}

	return [][]byte{value}, nil
}

// execRPop removes and returns the last element of the list
// RPOP key
func execRPop(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments for RPOP")
	}

	key := string(args[0])

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{[]byte("(nil)")}, nil
	}

	// Type check
	list, ok := entity.Data.(*datastruct.List)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	// Pop value
	value := list.RPop()
	if value == nil {
		// List became empty, remove key
		db.Remove(key)
		return [][]byte{[]byte("(nil)")}, nil
	}

	// Update database
	if list.Len() == 0 {
		db.Remove(key)
	} else {
		db.PutEntity(key, entity)
	}

	return [][]byte{value}, nil
}

// execLIndex returns the element at index in the list
// LINDEX key index
func execLIndex(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("wrong number of arguments for LINDEX")
	}

	key := string(args[0])
	index, err := strconv.Atoi(string(args[1]))
	if err != nil {
		return nil, errors.New("ERR value is not an integer or out of range")
	}

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{[]byte("(nil)")}, nil
	}

	// Type check
	list, ok := entity.Data.(*datastruct.List)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	// Get value
	value := list.LIndex(index)
	if value == nil {
		return [][]byte{[]byte("(nil)")}, nil
	}

	return [][]byte{value}, nil
}

// execLSet sets the element at index to value
// LSET key index value
func execLSet(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 3 {
		return nil, errors.New("wrong number of arguments for LSET")
	}

	key := string(args[0])
	index, err := strconv.Atoi(string(args[1]))
	if err != nil {
		return nil, errors.New("ERR value is not an integer or out of range")
	}
	value := args[2]

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return nil, errors.New("ERR no such key")
	}

	// Type check
	list, ok := entity.Data.(*datastruct.List)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	// Set value
	err = list.LSet(index, value)
	if err != nil {
		return nil, err
	}

	db.PutEntity(key, entity)
	return [][]byte{[]byte("OK")}, nil
}

// execLRange returns a slice of elements from start to stop (inclusive)
// LRANGE key start stop
func execLRange(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 3 {
		return nil, errors.New("wrong number of arguments for LRANGE")
	}

	key := string(args[0])
	start, err := strconv.Atoi(string(args[1]))
	if err != nil {
		return nil, errors.New("ERR value is not an integer or out of range")
	}
	stop, err := strconv.Atoi(string(args[2]))
	if err != nil {
		return nil, errors.New("ERR value is not an integer or out of range")
	}

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{}, nil
	}

	// Type check
	list, ok := entity.Data.(*datastruct.List)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	// Get range
	values := list.LRange(start, stop)

	// Convert to result
	result := make([][]byte, len(values))
	for i, val := range values {
		result[i] = val
	}

	return result, nil
}

// execLTrim trims the list to only contain elements from start to stop (inclusive)
// LTRIM key start stop
func execLTrim(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 3 {
		return nil, errors.New("wrong number of arguments for LTRIM")
	}

	key := string(args[0])
	start, err := strconv.Atoi(string(args[1]))
	if err != nil {
		return nil, errors.New("ERR value is not an integer or out of range")
	}
	stop, err := strconv.Atoi(string(args[2]))
	if err != nil {
		return nil, errors.New("ERR value is not an integer or out of range")
	}

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{[]byte("OK")}, nil
	}

	// Type check
	list, ok := entity.Data.(*datastruct.List)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	// Trim list
	list.LTrim(start, stop)

	// Remove key if empty
	if list.Len() == 0 {
		db.Remove(key)
	} else {
		db.PutEntity(key, entity)
	}

	return [][]byte{[]byte("OK")}, nil
}

// execLRem removes the first count occurrences of elements equal to value
// LREM key count value
func execLRem(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 3 {
		return nil, errors.New("wrong number of arguments for LREM")
	}

	key := string(args[0])
	count, err := strconv.Atoi(string(args[1]))
	if err != nil {
		return nil, errors.New("ERR value is not an integer or out of range")
	}
	value := args[2]

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{[]byte("0")}, nil
	}

	// Type check
	list, ok := entity.Data.(*datastruct.List)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	// Remove elements
	removed := list.LRem(count, value)

	// Remove key if empty
	if list.Len() == 0 {
		db.Remove(key)
	} else {
		db.PutEntity(key, entity)
	}

	return [][]byte{[]byte(strconv.FormatInt(int64(removed), 10))}, nil
}

// execLInsert inserts value before or after pivot value
// LINSERT key BEFORE|AFTER pivot value
func execLInsert(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 4 {
		return nil, errors.New("wrong number of arguments for LINSERT")
	}

	key := string(args[0])
	position := strings.ToLower(string(args[1]))
	pivot := args[2]
	value := args[3]

	var before bool
	switch position {
	case "before":
		before = true
	case "after":
		before = false
	default:
		return nil, errors.New("ERR syntax error")
	}

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{[]byte("0")}, nil
	}

	// Type check
	list, ok := entity.Data.(*datastruct.List)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	// Insert value
	length := list.LInsert(before, pivot, value)
	if length == -1 {
		// Pivot not found
		return [][]byte{[]byte("-1")}, nil
	}

	db.PutEntity(key, entity)
	return [][]byte{[]byte(strconv.FormatInt(int64(length), 10))}, nil
}

// execLLen returns the length of the list
// LLEN key
func execLLen(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments for LLEN")
	}

	key := string(args[0])

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{[]byte("0")}, nil
	}

	// Type check
	list, ok := entity.Data.(*datastruct.List)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	length := list.Len()
	return [][]byte{[]byte(strconv.FormatInt(int64(length), 10))}, nil
}

// Set command implementations

// execSAdd adds one or more members to a set
// SADD key member [member ...]
func execSAdd(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("wrong number of arguments for SADD")
	}

	key := string(args[0])
	members := args[1:]

	// Get or create entity
	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		entity = datastruct.MakeSet()
	}

	// Type check
	set, ok := entity.Data.(*datastruct.Set)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	// Add members
	added := set.Add(members...)
	db.PutEntity(key, entity)

	return [][]byte{[]byte(strconv.FormatInt(int64(added), 10))}, nil
}

// execSRem removes one or more members from a set
// SREM key member [member ...]
func execSRem(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("wrong number of arguments for SREM")
	}

	key := string(args[0])
	members := args[1:]

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{[]byte("0")}, nil
	}

	// Type check
	set, ok := entity.Data.(*datastruct.Set)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	// Remove members
	removed := set.Remove(members...)

	// Remove key if empty
	if set.Len() == 0 {
		db.Remove(key)
	} else {
		db.PutEntity(key, entity)
	}

	return [][]byte{[]byte(strconv.FormatInt(int64(removed), 10))}, nil
}

// execSIsMember checks if member is in the set
// SISMEMBER key member
func execSIsMember(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("wrong number of arguments for SISMEMBER")
	}

	key := string(args[0])
	member := args[1]

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{[]byte("0")}, nil
	}

	// Type check
	set, ok := entity.Data.(*datastruct.Set)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	if set.IsMember(member) {
		return [][]byte{[]byte("1")}, nil
	}
	return [][]byte{[]byte("0")}, nil
}

// execSMembers returns all members of the set
// SMEMBERS key
func execSMembers(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments for SMEMBERS")
	}

	key := string(args[0])

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{}, nil
	}

	// Type check
	set, ok := entity.Data.(*datastruct.Set)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	members := set.Members()
	result := make([][]byte, len(members))
	for i, member := range members {
		result[i] = member
	}

	return result, nil
}

// execSCard returns the number of members in the set
// SCARD key
func execSCard(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments for SCARD")
	}

	key := string(args[0])

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{[]byte("0")}, nil
	}

	// Type check
	set, ok := entity.Data.(*datastruct.Set)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	return [][]byte{[]byte(strconv.FormatInt(int64(set.Len()), 10))}, nil
}

// execSPop removes and returns a random member from the set
// SPOP key
func execSPop(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments for SPOP")
	}

	key := string(args[0])

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{[]byte("(nil)")}, nil
	}

	// Type check
	set, ok := entity.Data.(*datastruct.Set)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	// Pop random member
	member := set.Pop()
	if member == nil {
		// Set became empty, remove key
		db.Remove(key)
		return [][]byte{[]byte("(nil)")}, nil
	}

	// Update database
	if set.Len() == 0 {
		db.Remove(key)
	} else {
		db.PutEntity(key, entity)
	}

	return [][]byte{member}, nil
}

// execSRandMember returns a random member from the set without removing it
// SRANDMEMBER key
func execSRandMember(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments for SRANDMEMBER")
	}

	key := string(args[0])

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{[]byte("(nil)")}, nil
	}

	// Type check
	set, ok := entity.Data.(*datastruct.Set)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	member := set.GetRandom()
	if member == nil {
		return [][]byte{[]byte("(nil)")}, nil
	}

	return [][]byte{member}, nil
}

// execSMove moves a member from one set to another
// SMOVE source destination member
func execSMove(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 3 {
		return nil, errors.New("wrong number of arguments for SMOVE")
	}

	srcKey := string(args[0])
	dstKey := string(args[1])
	member := args[2]

	// Get source set
	srcEntity, ok := db.GetEntity(srcKey)
	if !ok || srcEntity.Data == nil {
		return [][]byte{[]byte("0")}, nil
	}

	// Type check source
	srcSet, ok := srcEntity.Data.(*datastruct.Set)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	// Get or create destination set
	dstEntity, ok := db.GetEntity(dstKey)
	if !ok || dstEntity.Data == nil {
		dstEntity = datastruct.MakeSet()
	}

	// Type check destination
	dstSet, ok := dstEntity.Data.(*datastruct.Set)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	// Move member
	moved := srcSet.Move(dstSet, member)
	if !moved {
		return [][]byte{[]byte("0")}, nil
	}

	// Update source
	if srcSet.Len() == 0 {
		db.Remove(srcKey)
	} else {
		db.PutEntity(srcKey, srcEntity)
	}

	// Update destination
	db.PutEntity(dstKey, dstEntity)

	return [][]byte{[]byte("1")}, nil
}

// execSDiff returns the difference between the first set and all other sets
// SDIFF key [key ...]
func execSDiff(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 1 {
		return nil, errors.New("wrong number of arguments for SDIFF")
	}

	keys := make([]string, len(args))
	for i, arg := range args {
		keys[i] = string(arg)
	}

	// Get first set
	entity, ok := db.GetEntity(keys[0])
	if !ok || entity.Data == nil {
		return [][]byte{}, nil
	}

	// Type check
	set, ok := entity.Data.(*datastruct.Set)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	// Get other sets
	others := make([]*datastruct.Set, 0, len(keys)-1)
	for _, key := range keys[1:] {
		entity, ok := db.GetEntity(key)
		if ok && entity.Data != nil {
			otherSet, ok := entity.Data.(*datastruct.Set)
			if !ok {
				return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
			}
			others = append(others, otherSet)
		}
	}

	// Calculate diff
	result := set.Diff(others)
	return result, nil
}

// execSDiffStore calculates the difference and stores it in destination
// SDIFFSTORE destination key [key ...]
func execSDiffStore(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("wrong number of arguments for SDIFFSTORE")
	}

	dstKey := string(args[0])
	srcKeys := args[1:]

	// Calculate diff using SDiff logic
	diff, err := execSDiff(db, srcKeys)
	if err != nil {
		return nil, err
	}

	// Create or update destination set
	dstEntity := datastruct.MakeSet()
	dstSet, _ := dstEntity.Data.(*datastruct.Set)
	dstSet.Add(diff...)

	db.PutEntity(dstKey, dstEntity)

	return [][]byte{[]byte(strconv.FormatInt(int64(len(diff)), 10))}, nil
}

// execSInter returns the intersection of all specified sets
// SINTER key [key ...]
func execSInter(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 1 {
		return nil, errors.New("wrong number of arguments for SINTER")
	}

	keys := make([]string, len(args))
	for i, arg := range args {
		keys[i] = string(arg)
	}

	// Get first set
	entity, ok := db.GetEntity(keys[0])
	if !ok || entity.Data == nil {
		return [][]byte{}, nil
	}

	// Type check
	set, ok := entity.Data.(*datastruct.Set)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	// If only one set, return all members
	if len(keys) == 1 {
		return set.Members(), nil
	}

	// Get other sets
	others := make([]*datastruct.Set, 0, len(keys)-1)
	for _, key := range keys[1:] {
		entity, ok := db.GetEntity(key)
		if !ok || entity.Data == nil {
			// If any set doesn't exist, intersection is empty
			return [][]byte{}, nil
		}
		otherSet, ok := entity.Data.(*datastruct.Set)
		if !ok {
			return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		others = append(others, otherSet)
	}

	// Calculate intersection
	result := set.Intersect(others)
	return result, nil
}

// execSInterStore calculates the intersection and stores it in destination
// SINTERSTORE destination key [key ...]
func execSInterStore(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("wrong number of arguments for SINTERSTORE")
	}

	dstKey := string(args[0])
	srcKeys := args[1:]

	// Calculate intersection using SInter logic
	intersection, err := execSInter(db, srcKeys)
	if err != nil {
		return nil, err
	}

	// Create or update destination set
	dstEntity := datastruct.MakeSet()
	dstSet, _ := dstEntity.Data.(*datastruct.Set)
	dstSet.Add(intersection...)

	db.PutEntity(dstKey, dstEntity)

	return [][]byte{[]byte(strconv.FormatInt(int64(len(intersection)), 10))}, nil
}

// execSUnion returns the union of all specified sets
// SUNION key [key ...]
func execSUnion(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 1 {
		return nil, errors.New("wrong number of arguments for SUNION")
	}

	keys := make([]string, len(args))
	for i, arg := range args {
		keys[i] = string(arg)
	}

	// Get first set
	entity, ok := db.GetEntity(keys[0])
	if !ok || entity.Data == nil {
		// If first set doesn't exist, treat as empty and get union of others
		if len(keys) == 1 {
			return [][]byte{}, nil
		}

		// Get union of remaining sets
		entity = datastruct.MakeSet()
	}

	// Type check first set
	set, ok := entity.Data.(*datastruct.Set)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	// Get other sets
	others := make([]*datastruct.Set, 0, len(keys)-1)
	for _, key := range keys[1:] {
		entity, ok := db.GetEntity(key)
		if ok && entity.Data != nil {
			otherSet, ok := entity.Data.(*datastruct.Set)
			if !ok {
				return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
			}
			others = append(others, otherSet)
		}
	}

	// Calculate union
	result := set.Union(others)
	return result, nil
}

// execSUnionStore calculates the union and stores it in destination
// SUNIONSTORE destination key [key ...]
func execSUnionStore(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("wrong number of arguments for SUNIONSTORE")
	}

	dstKey := string(args[0])
	srcKeys := args[1:]

	// Calculate union using SUnion logic
	union, err := execSUnion(db, srcKeys)
	if err != nil {
		return nil, err
	}

	// Create or update destination set
	dstEntity := datastruct.MakeSet()
	dstSet, _ := dstEntity.Data.(*datastruct.Set)
	dstSet.Add(union...)

	db.PutEntity(dstKey, dstEntity)

	return [][]byte{[]byte(strconv.FormatInt(int64(len(union)), 10))}, nil
}

// SortedSet command implementations

// execZAdd adds or updates members with scores in the sorted set
// ZADD key score member [score member ...]
func execZAdd(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 3 || len(args)%2 != 1 {
		return nil, errors.New("wrong number of arguments for ZADD")
	}

	key := string(args[0])

	// Get or create entity
	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		entity = datastruct.MakeSortedSet()
	}

	// Type check
	zset, ok := entity.Data.(*datastruct.SortedSet)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	// Parse score-member pairs
	added := 0
	for i := 1; i < len(args); i += 2 {
		score, err := strconv.ParseFloat(string(args[i]), 64)
		if err != nil {
			return nil, errors.New("ERR value is not a valid float")
		}
		member := args[i+1]
		added += zset.Add(score, member)
	}

	db.PutEntity(key, entity)
	return [][]byte{[]byte(strconv.FormatInt(int64(added), 10))}, nil
}

// execZRem removes one or more members from the sorted set
// ZREM key member [member ...]
func execZRem(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("wrong number of arguments for ZREM")
	}

	key := string(args[0])
	members := args[1:]

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{[]byte("0")}, nil
	}

	// Type check
	zset, ok := entity.Data.(*datastruct.SortedSet)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	// Remove members
	removed := zset.Remove(members...)

	// Remove key if empty
	if zset.Len() == 0 {
		db.Remove(key)
	} else {
		db.PutEntity(key, entity)
	}

	return [][]byte{[]byte(strconv.FormatInt(int64(removed), 10))}, nil
}

// execZScore returns the score of a member in the sorted set
// ZSCORE key member
func execZScore(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("wrong number of arguments for ZSCORE")
	}

	key := string(args[0])
	member := args[1]

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{[]byte("(nil)")}, nil
	}

	// Type check
	zset, ok := entity.Data.(*datastruct.SortedSet)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	score := zset.Score(member)
	if math.IsNaN(score) {
		return [][]byte{[]byte("(nil)")}, nil
	}

	return [][]byte{[]byte(strconv.FormatFloat(score, 'f', -1, 64))}, nil
}

// execZIncrBy increments the score of a member by increment
// ZINCRBY increment member
func execZIncrBy(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 3 {
		return nil, errors.New("wrong number of arguments for ZINCRBY")
	}

	key := string(args[0])
	increment, err := strconv.ParseFloat(string(args[1]), 64)
	if err != nil {
		return nil, errors.New("ERR value is not a valid float")
	}
	member := args[2]

	// Get or create entity
	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		entity = datastruct.MakeSortedSet()
	}

	// Type check
	zset, ok := entity.Data.(*datastruct.SortedSet)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	// Increment score
	newScore := zset.IncrBy(increment, member)
	db.PutEntity(key, entity)

	return [][]byte{[]byte(strconv.FormatFloat(newScore, 'f', -1, 64))}, nil
}

// execZCard returns the number of members in the sorted set
// ZCARD key
func execZCard(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments for ZCARD")
	}

	key := string(args[0])

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{[]byte("0")}, nil
	}

	// Type check
	zset, ok := entity.Data.(*datastruct.SortedSet)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	return [][]byte{[]byte(strconv.FormatInt(int64(zset.Len()), 10))}, nil
}

// execZRank returns the rank of a member (0-based, ascending)
// ZRANK key member
func execZRank(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("wrong number of arguments for ZRANK")
	}

	key := string(args[0])
	member := args[1]

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{[]byte("(nil)")}, nil
	}

	// Type check
	zset, ok := entity.Data.(*datastruct.SortedSet)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	rank := zset.Rank(member)
	if rank == -1 {
		return [][]byte{[]byte("(nil)")}, nil
	}

	return [][]byte{[]byte(strconv.FormatInt(int64(rank), 10))}, nil
}

// execZRevRank returns the rank of a member (0-based, descending)
// ZREVRANK key member
func execZRevRank(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("wrong number of arguments for ZREVRANK")
	}

	key := string(args[0])
	member := args[1]

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{[]byte("(nil)")}, nil
	}

	// Type check
	zset, ok := entity.Data.(*datastruct.SortedSet)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	rank := zset.RevRank(member)
	if rank == -1 {
		return [][]byte{[]byte("(nil)")}, nil
	}

	return [][]byte{[]byte(strconv.FormatInt(int64(rank), 10))}, nil
}

// execZRange returns members in the given range by rank (ascending)
// ZRANGE key start stop [WITHSCORES]
func execZRange(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 3 {
		return nil, errors.New("wrong number of arguments for ZRANGE")
	}

	key := string(args[0])
	start, err := strconv.Atoi(string(args[1]))
	if err != nil {
		return nil, errors.New("ERR value is not an integer or out of range")
	}
	stop, err := strconv.Atoi(string(args[2]))
	if err != nil {
		return nil, errors.New("ERR value is not an integer or out of range")
	}

	withScores := false
	if len(args) >= 4 {
		if strings.ToUpper(string(args[3])) == "WITHSCORES" {
			withScores = true
		}
	}

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{}, nil
	}

	// Type check
	zset, ok := entity.Data.(*datastruct.SortedSet)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	result := zset.Range(start, stop, withScores)
	return result, nil
}

// execZRevRange returns members in the given range by rank (descending)
// ZREVRANGE key start stop [WITHSCORES]
func execZRevRange(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 3 {
		return nil, errors.New("wrong number of arguments for ZREVRANGE")
	}

	key := string(args[0])
	start, err := strconv.Atoi(string(args[1]))
	if err != nil {
		return nil, errors.New("ERR value is not an integer or out of range")
	}
	stop, err := strconv.Atoi(string(args[2]))
	if err != nil {
		return nil, errors.New("ERR value is not an integer or out of range")
	}

	withScores := false
	if len(args) >= 4 {
		if strings.ToUpper(string(args[3])) == "WITHSCORES" {
			withScores = true
		}
	}

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{}, nil
	}

	// Type check
	zset, ok := entity.Data.(*datastruct.SortedSet)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	result := zset.RevRange(start, stop, withScores)
	return result, nil
}

// execZRangeByScore returns members with scores between min and max
// ZRANGEBYSCORE key min max [WITHSCORES] [LIMIT offset count]
func execZRangeByScore(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 3 {
		return nil, errors.New("wrong number of arguments for ZRANGEBYSCORE")
	}

	key := string(args[0])
	min, err := strconv.ParseFloat(string(args[1]), 64)
	if err != nil {
		return nil, errors.New("ERR value is not a valid float")
	}
	max, err := strconv.ParseFloat(string(args[2]), 64)
	if err != nil {
		return nil, errors.New("ERR value is not a valid float")
	}

	withScores := false
	offset := 0
	count := -1

	// Parse optional arguments
	for i := 3; i < len(args); i++ {
		arg := strings.ToUpper(string(args[i]))
		if arg == "WITHSCORES" {
			withScores = true
		} else if arg == "LIMIT" && i+2 < len(args) {
			offset, err = strconv.Atoi(string(args[i+1]))
			if err != nil {
				return nil, errors.New("ERR value is not an integer")
			}
			count, err = strconv.Atoi(string(args[i+2]))
			if err != nil {
				return nil, errors.New("ERR value is not an integer")
			}
			i += 2
		}
	}

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{}, nil
	}

	// Type check
	zset, ok := entity.Data.(*datastruct.SortedSet)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	// Check if LIMIT is specified
	if count > 0 || offset > 0 {
		result := zset.RangeByScoreWithLimit(min, max, offset, count, withScores, false)
		return result, nil
	}

	result := zset.RangeByScore(min, max, withScores)
	return result, nil
}

// execZCount returns the number of members with scores between min and max
// ZCOUNT key min max
func execZCount(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 3 {
		return nil, errors.New("wrong number of arguments for ZCOUNT")
	}

	key := string(args[0])
	min, err := strconv.ParseFloat(string(args[1]), 64)
	if err != nil {
		return nil, errors.New("ERR value is not a valid float")
	}
	max, err := strconv.ParseFloat(string(args[2]), 64)
	if err != nil {
		return nil, errors.New("ERR value is not a valid float")
	}

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{[]byte("0")}, nil
	}

	// Type check
	zset, ok := entity.Data.(*datastruct.SortedSet)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	count := zset.Count(min, max)
	return [][]byte{[]byte(strconv.FormatInt(int64(count), 10))}, nil
}
