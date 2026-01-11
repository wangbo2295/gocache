package database

import (
	"errors"
	"strconv"

	"github.com/wangbo/gocache/datastruct"
)

// Set command implementations

func execSAdd(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("wrong number of arguments for SADD")
	}

	key := string(args[0])
	members := args[1:]

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		entity = datastruct.MakeSet()
	}

	set, ok := entity.Data.(*datastruct.Set)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	added := set.Add(members...)
	db.PutEntity(key, entity)

	return [][]byte{[]byte(strconv.FormatInt(int64(added), 10))}, nil
}

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

	set, ok := entity.Data.(*datastruct.Set)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	removed := set.Remove(members...)

	if set.Len() == 0 {
		db.Remove(key)
	} else {
		db.PutEntity(key, entity)
	}

	return [][]byte{[]byte(strconv.FormatInt(int64(removed), 10))}, nil
}

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

	set, ok := entity.Data.(*datastruct.Set)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	if set.IsMember(member) {
		return [][]byte{[]byte("1")}, nil
	}
	return [][]byte{[]byte("0")}, nil
}

func execSMembers(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments for SMEMBERS")
	}

	key := string(args[0])

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{}, nil
	}

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

func execSCard(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments for SCARD")
	}

	key := string(args[0])

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{[]byte("0")}, nil
	}

	set, ok := entity.Data.(*datastruct.Set)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	return [][]byte{[]byte(strconv.FormatInt(int64(set.Len()), 10))}, nil
}

func execSPop(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments for SPOP")
	}

	key := string(args[0])

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{nil}, nil
	}

	set, ok := entity.Data.(*datastruct.Set)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	member := set.Pop()
	if member == nil {
		db.Remove(key)
		return [][]byte{nil}, nil
	}

	if set.Len() == 0 {
		db.Remove(key)
	} else {
		db.PutEntity(key, entity)
	}

	return [][]byte{member}, nil
}

func execSRandMember(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments for SRANDMEMBER")
	}

	key := string(args[0])

	entity, ok := db.GetEntity(key)
	if !ok || entity.Data == nil {
		return [][]byte{nil}, nil
	}

	set, ok := entity.Data.(*datastruct.Set)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	member := set.GetRandom()
	if member == nil {
		return [][]byte{nil}, nil
	}

	return [][]byte{member}, nil
}

func execSMove(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 3 {
		return nil, errors.New("wrong number of arguments for SMOVE")
	}

	srcKey := string(args[0])
	dstKey := string(args[1])
	member := args[2]

	srcEntity, ok := db.GetEntity(srcKey)
	if !ok || srcEntity.Data == nil {
		return [][]byte{[]byte("0")}, nil
	}

	srcSet, ok := srcEntity.Data.(*datastruct.Set)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	dstEntity, ok := db.GetEntity(dstKey)
	if !ok || dstEntity.Data == nil {
		dstEntity = datastruct.MakeSet()
	}

	dstSet, ok := dstEntity.Data.(*datastruct.Set)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	moved := srcSet.Move(dstSet, member)
	if !moved {
		return [][]byte{[]byte("0")}, nil
	}

	if srcSet.Len() == 0 {
		db.Remove(srcKey)
	} else {
		db.PutEntity(srcKey, srcEntity)
	}

	db.PutEntity(dstKey, dstEntity)

	return [][]byte{[]byte("1")}, nil
}

func execSDiff(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 1 {
		return nil, errors.New("wrong number of arguments for SDIFF")
	}

	keys := make([]string, len(args))
	for i, arg := range args {
		keys[i] = string(arg)
	}

	entity, ok := db.GetEntity(keys[0])
	if !ok || entity.Data == nil {
		return [][]byte{}, nil
	}

	set, ok := entity.Data.(*datastruct.Set)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

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

	result := set.Diff(others)
	return result, nil
}

func execSDiffStore(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("wrong number of arguments for SDIFFSTORE")
	}

	dstKey := string(args[0])
	srcKeys := args[1:]

	diff, err := execSDiff(db, srcKeys)
	if err != nil {
		return nil, err
	}

	dstEntity := datastruct.MakeSet()
	dstSet, _ := dstEntity.Data.(*datastruct.Set)
	dstSet.Add(diff...)

	db.PutEntity(dstKey, dstEntity)

	return [][]byte{[]byte(strconv.FormatInt(int64(len(diff)), 10))}, nil
}

func execSInter(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 1 {
		return nil, errors.New("wrong number of arguments for SINTER")
	}

	keys := make([]string, len(args))
	for i, arg := range args {
		keys[i] = string(arg)
	}

	entity, ok := db.GetEntity(keys[0])
	if !ok || entity.Data == nil {
		return [][]byte{}, nil
	}

	set, ok := entity.Data.(*datastruct.Set)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	if len(keys) == 1 {
		return set.Members(), nil
	}

	others := make([]*datastruct.Set, 0, len(keys)-1)
	for _, key := range keys[1:] {
		entity, ok := db.GetEntity(key)
		if !ok || entity.Data == nil {
			return [][]byte{}, nil
		}
		otherSet, ok := entity.Data.(*datastruct.Set)
		if !ok {
			return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		others = append(others, otherSet)
	}

	result := set.Intersect(others)
	return result, nil
}

func execSInterStore(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("wrong number of arguments for SINTERSTORE")
	}

	dstKey := string(args[0])
	srcKeys := args[1:]

	intersection, err := execSInter(db, srcKeys)
	if err != nil {
		return nil, err
	}

	dstEntity := datastruct.MakeSet()
	dstSet, _ := dstEntity.Data.(*datastruct.Set)
	dstSet.Add(intersection...)

	db.PutEntity(dstKey, dstEntity)

	return [][]byte{[]byte(strconv.FormatInt(int64(len(intersection)), 10))}, nil
}

func execSUnion(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 1 {
		return nil, errors.New("wrong number of arguments for SUNION")
	}

	keys := make([]string, len(args))
	for i, arg := range args {
		keys[i] = string(arg)
	}

	entity, ok := db.GetEntity(keys[0])
	if !ok || entity.Data == nil {
		if len(keys) == 1 {
			return [][]byte{}, nil
		}
		entity = datastruct.MakeSet()
	}

	set, ok := entity.Data.(*datastruct.Set)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

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

	result := set.Union(others)
	return result, nil
}

func execSUnionStore(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("wrong number of arguments for SUNIONSTORE")
	}

	dstKey := string(args[0])
	srcKeys := args[1:]

	union, err := execSUnion(db, srcKeys)
	if err != nil {
		return nil, err
	}

	dstEntity := datastruct.MakeSet()
	dstSet, _ := dstEntity.Data.(*datastruct.Set)
	dstSet.Add(union...)

	db.PutEntity(dstKey, dstEntity)

	return [][]byte{[]byte(strconv.FormatInt(int64(len(union)), 10))}, nil
}
