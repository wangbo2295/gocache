package database

import (
	"errors"
	"strconv"

	"github.com/wangbo/gocache/datastruct"
)

// String command implementations

// Pre-allocated responses to reduce allocations
var (
	okResponse     = [][]byte{[]byte("OK")}
	zeroResponse   = [][]byte{[]byte("0")}
	oneResponse    = [][]byte{[]byte("1")}
	emptyResponse  = [][]byte{[]byte("")}
	nilResponse    = [][]byte{nil}
)

func execSet(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("wrong number of arguments")
	}

	key := string(args[0])
	value := args[1]

	entity := datastruct.MakeString(value)
	db.PutEntity(key, entity)

	// Clear any existing TTL (SET overwrites key completely)
	db.Persist(key)

	// Use pre-allocated OK response
	return okResponse, nil
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

	// Use pre-allocated responses for common cases
	switch count {
	case 0:
		return zeroResponse, nil
	case 1:
		return oneResponse, nil
	default:
		return [][]byte{[]byte(strconv.Itoa(count))}, nil
	}
}

func execKeys(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments")
	}

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
	newVal, err := db.atomicIncr(key, 1)
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

	newVal, err := db.atomicIncr(key, delta)
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

	// Use pre-allocated OK response
	return okResponse, nil
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
		return emptyResponse, nil
	}

	str, ok := entity.Data.(*datastruct.String)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	result := str.GetRange(start, end)
	if len(result) == 0 {
		return emptyResponse, nil
	}
	return [][]byte{result}, nil
}
