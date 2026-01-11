package database

import (
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/wangbo/gocache/datastruct"
)

// SortedSet command implementations

func execZAdd(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 3 || len(args)%2 != 1 {
		return nil, errors.New("wrong number of arguments for ZADD")
	}

	key := string(args[0])

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		entity = datastruct.MakeSortedSet()
	}

	zset, ok := entity.Data.(*datastruct.SortedSet)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

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

	zset, ok := entity.Data.(*datastruct.SortedSet)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	removed := zset.Remove(members...)

	if zset.Len() == 0 {
		db.Remove(key)
	} else {
		db.PutEntity(key, entity)
	}

	return [][]byte{[]byte(strconv.FormatInt(int64(removed), 10))}, nil
}

func execZScore(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("wrong number of arguments for ZSCORE")
	}

	key := string(args[0])
	member := args[1]

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{nil}, nil
	}

	zset, ok := entity.Data.(*datastruct.SortedSet)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	score := zset.Score(member)
	if math.IsNaN(score) {
		return [][]byte{nil}, nil
	}

	return [][]byte{[]byte(strconv.FormatFloat(score, 'f', -1, 64))}, nil
}

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

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		entity = datastruct.MakeSortedSet()
	}

	zset, ok := entity.Data.(*datastruct.SortedSet)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	newScore := zset.IncrBy(increment, member)
	db.PutEntity(key, entity)

	return [][]byte{[]byte(strconv.FormatFloat(newScore, 'f', -1, 64))}, nil
}

func execZCard(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments for ZCARD")
	}

	key := string(args[0])

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{[]byte("0")}, nil
	}

	zset, ok := entity.Data.(*datastruct.SortedSet)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	return [][]byte{[]byte(strconv.FormatInt(int64(zset.Len()), 10))}, nil
}

func execZRank(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("wrong number of arguments for ZRANK")
	}

	key := string(args[0])
	member := args[1]

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{nil}, nil
	}

	zset, ok := entity.Data.(*datastruct.SortedSet)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	rank := zset.Rank(member)
	if rank == -1 {
		return [][]byte{nil}, nil
	}

	return [][]byte{[]byte(strconv.FormatInt(int64(rank), 10))}, nil
}

func execZRevRank(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("wrong number of arguments for ZREVRANK")
	}

	key := string(args[0])
	member := args[1]

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{nil}, nil
	}

	zset, ok := entity.Data.(*datastruct.SortedSet)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	rank := zset.RevRank(member)
	if rank == -1 {
		return [][]byte{nil}, nil
	}

	return [][]byte{[]byte(strconv.FormatInt(int64(rank), 10))}, nil
}

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

	zset, ok := entity.Data.(*datastruct.SortedSet)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	result := zset.Range(start, stop, withScores)
	return result, nil
}

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

	zset, ok := entity.Data.(*datastruct.SortedSet)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	result := zset.RevRange(start, stop, withScores)
	return result, nil
}

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

	zset, ok := entity.Data.(*datastruct.SortedSet)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	if count > 0 || offset > 0 {
		result := zset.RangeByScoreWithLimit(min, max, offset, count, withScores, false)
		return result, nil
	}

	result := zset.RangeByScore(min, max, withScores)
	return result, nil
}

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

	zset, ok := entity.Data.(*datastruct.SortedSet)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	count := zset.Count(min, max)
	return [][]byte{[]byte(strconv.FormatInt(int64(count), 10))}, nil
}
