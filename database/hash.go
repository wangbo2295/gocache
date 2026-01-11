package database

import (
	"errors"
	"strconv"

	"github.com/wangbo/gocache/datastruct"
)

// Hash command implementations

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

	for i := 1; i < len(args); i += 2 {
		field := string(args[i])
		value := args[i+1]
		hash.Set(field, value)
	}

	db.PutEntity(key, entity)
	return [][]byte{[]byte("OK")}, nil
}
