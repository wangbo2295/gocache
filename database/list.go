package database

import (
	"errors"
	"strconv"
	"strings"

	"github.com/wangbo/gocache/datastruct"
)

// List command implementations

func execLPush(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("wrong number of arguments for LPUSH")
	}

	key := string(args[0])
	values := args[1:]

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		entity = datastruct.MakeList()
	}

	list, ok := entity.Data.(*datastruct.List)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	length := list.LPush(values...)
	db.PutEntity(key, entity)

	return [][]byte{[]byte(strconv.FormatInt(int64(length), 10))}, nil
}

func execRPush(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("wrong number of arguments for RPUSH")
	}

	key := string(args[0])
	values := args[1:]

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		entity = datastruct.MakeList()
	}

	list, ok := entity.Data.(*datastruct.List)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	length := list.RPush(values...)
	db.PutEntity(key, entity)

	return [][]byte{[]byte(strconv.FormatInt(int64(length), 10))}, nil
}

func execLPop(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments for LPOP")
	}

	key := string(args[0])

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{nil}, nil
	}

	list, ok := entity.Data.(*datastruct.List)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	value := list.LPop()
	if value == nil {
		db.Remove(key)
		return [][]byte{nil}, nil
	}

	if list.Len() == 0 {
		db.Remove(key)
	} else {
		db.PutEntity(key, entity)
	}

	return [][]byte{value}, nil
}

func execRPop(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments for RPOP")
	}

	key := string(args[0])

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{nil}, nil
	}

	list, ok := entity.Data.(*datastruct.List)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	value := list.RPop()
	if value == nil {
		db.Remove(key)
		return [][]byte{nil}, nil
	}

	if list.Len() == 0 {
		db.Remove(key)
	} else {
		db.PutEntity(key, entity)
	}

	return [][]byte{value}, nil
}

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
		return [][]byte{nil}, nil
	}

	list, ok := entity.Data.(*datastruct.List)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	value := list.LIndex(index)
	if value == nil {
		return [][]byte{nil}, nil
	}

	return [][]byte{value}, nil
}

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

	list, ok := entity.Data.(*datastruct.List)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	err = list.LSet(index, value)
	if err != nil {
		return nil, err
	}

	db.PutEntity(key, entity)
	return [][]byte{[]byte("OK")}, nil
}

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

	list, ok := entity.Data.(*datastruct.List)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	values := list.LRange(start, stop)

	result := make([][]byte, len(values))
	for i, val := range values {
		result[i] = val
	}

	return result, nil
}

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

	list, ok := entity.Data.(*datastruct.List)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	list.LTrim(start, stop)

	if list.Len() == 0 {
		db.Remove(key)
	} else {
		db.PutEntity(key, entity)
	}

	return [][]byte{[]byte("OK")}, nil
}

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

	list, ok := entity.Data.(*datastruct.List)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	removed := list.LRem(count, value)

	if list.Len() == 0 {
		db.Remove(key)
	} else {
		db.PutEntity(key, entity)
	}

	return [][]byte{[]byte(strconv.FormatInt(int64(removed), 10))}, nil
}

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

	list, ok := entity.Data.(*datastruct.List)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	length := list.LInsert(before, pivot, value)
	if length == -1 {
		return [][]byte{[]byte("-1")}, nil
	}

	db.PutEntity(key, entity)
	return [][]byte{[]byte(strconv.FormatInt(int64(length), 10))}, nil
}

func execLLen(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments for LLEN")
	}

	key := string(args[0])

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{[]byte("0")}, nil
	}

	list, ok := entity.Data.(*datastruct.List)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	length := list.Len()
	return [][]byte{[]byte(strconv.FormatInt(int64(length), 10))}, nil
}
