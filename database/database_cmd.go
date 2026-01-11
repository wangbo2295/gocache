package database

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/wangbo/gocache/datastruct"
)

// Database command implementations

func execSelect(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments for SELECT")
	}

	index, err := strconv.Atoi(string(args[0]))
	if err != nil {
		return nil, errors.New("ERR invalid database index")
	}

	if index < 0 || index >= 16 {
		return nil, errors.New("ERR DB index is out of range")
	}

	// Note: This is a placeholder. In a real implementation with multiple databases,
	// you would have a map of databases and switch the active database.
	// For now, we just return OK as we're using a single database.
	// The actual database switching would be handled at the server/handler level.

	return okResponse, nil
}

func execType(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of arguments for TYPE")
	}

	key := string(args[0])
	entity, ok := db.GetEntity(key)
	if !ok {
		return [][]byte{[]byte("none")}, nil
	}

	switch entity.Data.(type) {
	case *datastruct.String:
		return [][]byte{[]byte("string")}, nil
	case *datastruct.Hash:
		return [][]byte{[]byte("hash")}, nil
	case *datastruct.List:
		return [][]byte{[]byte("list")}, nil
	case *datastruct.Set:
		return [][]byte{[]byte("set")}, nil
	case *datastruct.SortedSet:
		return [][]byte{[]byte("zset")}, nil
	default:
		return [][]byte{[]byte("none")}, nil
	}
}

func execMove(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("wrong number of arguments for MOVE")
	}

	key := string(args[0])
	destinationDB, err := strconv.Atoi(string(args[1]))
	if err != nil {
		return nil, errors.New("ERR value is not an integer or out of range")
	}

	if destinationDB < 0 || destinationDB >= 16 {
		return nil, errors.New("ERR DB index is out of range")
	}

	// Check if key exists
	_, ok := db.GetEntity(key)
	if !ok {
		return zeroResponse, nil
	}

	// Note: In a real implementation with multiple databases, we would:
	// 1. Get the entity from current database
	// 2. Remove it from current database
	// 3. Put it in the destination database
	// For now, since we have only one database, we just return 0 (not moved)

	// Placeholder: Return 0 to indicate not moved (single DB implementation)
	return zeroResponse, nil
}

// Helper function to get type name for an entity
func getEntityTypeName(entity *datastruct.DataEntity) string {
	if entity == nil || entity.Data == nil {
		return "none"
	}

	switch entity.Data.(type) {
	case *datastruct.String:
		return "string"
	case *datastruct.Hash:
		return "hash"
	case *datastruct.List:
		return "list"
	case *datastruct.Set:
		return "set"
	case *datastruct.SortedSet:
		return "zset"
	default:
		return "none"
	}
}

// FormatEntityInfo formats entity information for debugging/monitoring
func FormatEntityInfo(key string, entity *datastruct.DataEntity) string {
	if entity == nil {
		return fmt.Sprintf("Key: %s, Type: none", key)
	}

	typeName := getEntityTypeName(entity)
	return fmt.Sprintf("Key: %s, Type: %s", key, typeName)
}
