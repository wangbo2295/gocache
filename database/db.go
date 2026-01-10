package database

import (
	"errors"
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
