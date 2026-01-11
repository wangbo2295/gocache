package database

import (
	"errors"
	"strconv"
	"time"
)

// TTL command implementations

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

	if ttl == -2 {
		return [][]byte{[]byte("-2")}, nil
	}
	if ttl == -1 {
		return [][]byte{[]byte("-1")}, nil
	}

	seconds := int64(ttl.Seconds())
	return [][]byte{[]byte(strconv.FormatInt(seconds, 10))}, nil
}

func execPTTL(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments")
	}

	key := string(args[0])
	ttl := db.TTL(key)

	if ttl == -2 {
		return [][]byte{[]byte("-2")}, nil
	}
	if ttl == -1 {
		return [][]byte{[]byte("-1")}, nil
	}

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

func execExpireAt(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("wrong number of arguments")
	}

	key := string(args[0])
	timestamp, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return nil, errors.New("ERR value is not an integer or out of range")
	}

	expireTime := time.Unix(timestamp, 0)
	ttl := time.Until(expireTime)

	if ttl <= 0 {
		// Already expired or invalid, remove key if exists
		db.Remove(key)
		return zeroResponse, nil
	}

	result := db.Expire(key, ttl)
	return [][]byte{[]byte(strconv.Itoa(result))}, nil
}

func execPExpireAt(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("wrong number of arguments")
	}

	key := string(args[0])
	timestampMs, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return nil, errors.New("ERR value is not an integer or out of range")
	}

	expireTime := time.Unix(0, timestampMs*int64(time.Millisecond))
	ttl := time.Until(expireTime)

	if ttl <= 0 {
		// Already expired or invalid, remove key if exists
		db.Remove(key)
		return zeroResponse, nil
	}

	result := db.Expire(key, ttl)
	return [][]byte{[]byte(strconv.Itoa(result))}, nil
}
