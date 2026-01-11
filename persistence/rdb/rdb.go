package rdb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/wangbo/gocache/database"
	"github.com/wangbo/gocache/datastruct"
)

// RDB file format constants
const (
	RedisMagicString    = "REDIS"
	RDBVersion          = 9 // RDB version 9
	OpcodeEOF           = 255
	OpcodeSelectDB      = 254
	OpcodeResizeDB      = 251
	OpcodeAux           = 250
	OpcodeExpireTimeMS  = 253
	OpcodeExpireTime    = 252
	OpcodeFreq          = 246
	OpcodeUnused       = 245
)

// Value type encodings
const (
	TypeString       = 0
	TypeList         = 1
	TypeSet          = 2
	TypeZSet         = 3
	TypeHash         = 4
	TypeHashZipMap   = 9
	TypeListZipList  = 10
	TypeSetIntSet    = 11
	TypeZSetZipList  = 12
	TypeHashZipList  = 13
	TypeQuickList    = 14
	TypeStreamListPacks = 15
	TypeModule       = 7
	TypeModule2      = 6
)

// Length encoding constants
const (
	Len6Bit    = 0
	Len14Bit   = 1
	Len32Bit   = 2
	EncVal     = 3
	Compress   = 1
	EncInt8    = 0
	EncInt16   = 1
	EncInt32   = 2
	EncLZF     = 3
)

// Generator generates RDB files from database state
type Generator struct {
	db          *database.DB
	output      io.Writer
	auxFields   map[string]string
}

// MakeGenerator creates a new RDB generator
func MakeGenerator(db *database.DB) *Generator {
	return &Generator{
		db:        db,
		auxFields: make(map[string]string),
	}
}

// AddAuxField adds an auxiliary field to the RDB file
func (g *Generator) AddAuxField(key, value string) {
	g.auxFields[key] = value
}

// Generate generates an RDB file to the given writer
func (g *Generator) Generate(output io.Writer) error {
	g.output = output

	// Write magic string and version
	if err := g.writeHeader(); err != nil {
		return err
	}

	// Write auxiliary fields
	for key, value := range g.auxFields {
		if err := g.writeAuxField(key, value); err != nil {
			return err
		}
	}

	// Write database selector (DB 0)
	if err := g.writeSelectDB(0); err != nil {
		return err
	}

	// Write resize db (optional)
	// if err := g.writeResizeDB(); err != nil {
	// 	return err
	// }

	// Write all key-value pairs
	keys := g.db.Keys()
	for _, key := range keys {
		entity, ok := g.db.GetEntity(key)
		if !ok || entity == nil {
			continue
		}

		// Check for TTL
		ttl := g.db.TTL(key)
		if ttl > 0 {
			// Write with millisecond precision expiry
			if err := g.writeExpireTimeMS(int64(ttl)); err != nil {
				return err
			}
		}

		// Write value based on type
		if err := g.writeValue(key, entity); err != nil {
			return err
		}
	}

	// Write EOF marker
	if err := g.writeEOF(); err != nil {
		return err
	}

	// Write CRC64 checksum (8 bytes)
	// For simplicity, we write zeros
	checksum := make([]byte, 8)
	if _, err := g.output.Write(checksum); err != nil {
		return err
	}

	return nil
}

// writeHeader writes the RDB file header
func (g *Generator) writeHeader() error {
	// Magic string "REDIS" + version
	if _, err := g.output.Write([]byte(RedisMagicString)); err != nil {
		return err
	}

	version := make([]byte, 4)
	version[0] = byte(RDBVersion)
	version[1] = 0
	version[2] = 0
	version[3] = 0

	if _, err := g.output.Write(version); err != nil {
		return err
	}

	return nil
}

// writeAuxField writes an auxiliary field
func (g *Generator) writeAuxField(key, value string) error {
	// Write opcode
	if err := g.writeByte(OpcodeAux); err != nil {
		return err
	}

	// Write key
	if err := g.writeString(key); err != nil {
		return err
	}

	// Write value
	if err := g.writeString(value); err != nil {
		return err
	}

	return nil
}

// writeSelectDB writes the database selector opcode
func (g *Generator) writeSelectDB(dbID int) error {
	if err := g.writeByte(OpcodeSelectDB); err != nil {
		return err
	}
	return g.writeLength(uint64(dbID))
}

// writeExpireTimeMS writes the expire time in milliseconds
func (g *Generator) writeExpireTimeMS(ttl int64) error {
	if err := g.writeByte(OpcodeExpireTimeMS); err != nil {
		return err
	}

	// Write as 64-bit unsigned integer (milliseconds)
	expire := make([]byte, 8)
	binary.LittleEndian.PutUint64(expire, uint64(ttl))
	_, err := g.output.Write(expire)
	return err
}

// writeEOF writes the EOF opcode
func (g *Generator) writeEOF() error {
	return g.writeByte(OpcodeEOF)
}

// writeLength writes a length-encoded integer
func (g *Generator) writeLength(length uint64) error {
	if length < 64 {
		// 6-bit length
		return g.writeByte(byte(length) << 2)
	} else if length < 16384 {
		// 14-bit length
		b1 := byte((length>>8)<<2 | Len14Bit)
		b2 := byte(length)
		if err := g.writeByte(b1); err != nil {
			return err
		}
		return g.writeByte(b2)
	} else {
		// 32-bit length
		if err := g.writeByte(byte(Len32Bit << 2)); err != nil {
			return err
		}
		return g.writeUint32(uint32(length))
	}
}

// writeString writes a string
func (g *Generator) writeString(s string) error {
	// For simplicity, always use string encoding
	if len(s) < 64 {
		return g.writeStringEncoding([]byte(s))
	}
	return g.writeStringEncoding([]byte(s))
}

// writeStringEncoding writes a string with length encoding
func (g *Generator) writeStringEncoding(data []byte) error {
	if err := g.writeLength(uint64(len(data))); err != nil {
		return err
	}
	_, err := g.output.Write(data)
	return err
}

// writeValue writes a value based on its type
func (g *Generator) writeValue(key string, entity *datastruct.DataEntity) error {
	switch data := entity.Data.(type) {
	case *datastruct.String:
		return g.writeStringValue(key, data)
	case *datastruct.Hash:
		return g.writeHashValue(key, data)
	case *datastruct.List:
		return g.writeListValue(key, data)
	case *datastruct.Set:
		return g.writeSetValue(key, data)
	case *datastruct.SortedSet:
		return g.writeZSetValue(key, data)
	default:
		return fmt.Errorf("unknown type: %T", data)
	}
}

// writeStringValue writes a string value
func (g *Generator) writeStringValue(key string, data *datastruct.String) error {
	// Write type
	if err := g.writeByte(TypeString); err != nil {
		return err
	}

	// Write key
	if err := g.writeString(key); err != nil {
		return err
	}

	// Write value
	return g.writeStringEncoding(data.Value)
}

// writeHashValue writes a hash value
func (g *Generator) writeHashValue(key string, data *datastruct.Hash) error {
	// Write type
	if err := g.writeByte(TypeHash); err != nil {
		return err
	}

	// Write key
	if err := g.writeString(key); err != nil {
		return err
	}

	// Write number of fields
	allData := data.GetAll()
	if err := g.writeLength(uint64(len(allData))); err != nil {
		return err
	}

	// Write field-value pairs
	for field, value := range allData {
		if err := g.writeStringEncoding([]byte(field)); err != nil {
			return err
		}
		if err := g.writeStringEncoding(value); err != nil {
			return err
		}
	}

	return nil
}

// writeListValue writes a list value
func (g *Generator) writeListValue(key string, data *datastruct.List) error {
	// Write type
	if err := g.writeByte(TypeList); err != nil {
		return err
	}

	// Write key
	if err := g.writeString(key); err != nil {
		return err
	}

	// Write number of elements
	elements := data.GetAll()
	if err := g.writeLength(uint64(len(elements))); err != nil {
		return err
	}

	// Write elements
	for _, elem := range elements {
		if err := g.writeStringEncoding(elem); err != nil {
			return err
		}
	}

	return nil
}

// writeSetValue writes a set value
func (g *Generator) writeSetValue(key string, data *datastruct.Set) error {
	// Write type
	if err := g.writeByte(TypeSet); err != nil {
		return err
	}

	// Write key
	if err := g.writeString(key); err != nil {
		return err
	}

	// Write number of members
	members := data.Members()
	if err := g.writeLength(uint64(len(members))); err != nil {
		return err
	}

	// Write members
	for _, member := range members {
		if err := g.writeStringEncoding(member); err != nil {
			return err
		}
	}

	return nil
}

// writeZSetValue writes a sorted set value
func (g *Generator) writeZSetValue(key string, data *datastruct.SortedSet) error {
	// Write type
	if err := g.writeByte(TypeZSet); err != nil {
		return err
	}

	// Write key
	if err := g.writeString(key); err != nil {
		return err
	}

	// Write number of elements
	if err := g.writeLength(uint64(data.Len())); err != nil {
		return err
	}

	// Write member-score pairs
	for i := 0; i < data.Len(); i++ {
		member := data.GetMemberByRank(i)
		score := data.GetScoreByRank(i)

		// Write score as double
		if err := g.writeDouble(score); err != nil {
			return err
		}

		// Write member
		if err := g.writeStringEncoding(member); err != nil {
			return err
		}
	}

	return nil
}

// Helper functions for writing primitive types

func (g *Generator) writeByte(b byte) error {
	_, err := g.output.Write([]byte{b})
	return err
}

func (g *Generator) writeUint32(v uint32) error {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, v)
	_, err := g.output.Write(buf)
	return err
}

func (g *Generator) writeDouble(f float64) error {
	// Write double in little-endian format
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, f)
	_, err := g.output.Write(buf.Bytes())
	return err
}

// SaveToFile saves the database to an RDB file
func SaveToFile(db *database.DB, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create RDB file: %w", err)
	}
	defer file.Close()

	generator := MakeGenerator(db)

	// Add Redis version info
	generator.AddAuxField("redis-ver", "6.0.0")
	generator.AddAuxField("redis-bits", "64")
	generator.AddAuxField("ctime", fmt.Sprintf("%d", time.Now().Unix()))

	if err := generator.Generate(file); err != nil {
		return fmt.Errorf("failed to generate RDB: %w", err)
	}

	// Sync to disk
	return file.Sync()
}

// RDBSaver implements persistence.DBSaver interface
type RDBSaver struct{}

// SaveDB saves the database to an RDB file
func (s *RDBSaver) SaveDB(db interface{}, filename string) error {
	// Type assert to *database.DB
	dbTyped, ok := db.(*database.DB)
	if !ok {
		return fmt.Errorf("invalid database type")
	}
	return SaveToFile(dbTyped, filename)
}

// SaveDBToWriter saves the database to an io.Writer
func (s *RDBSaver) SaveDBToWriter(db interface{}, writer io.Writer) error {
	// Type assert to *database.DB
	dbTyped, ok := db.(*database.DB)
	if !ok {
		return fmt.Errorf("invalid database type")
	}
	return SaveToWriter(dbTyped, writer)
}

// SaveToWriter saves the database to an io.Writer
func SaveToWriter(db *database.DB, writer io.Writer) error {
	generator := MakeGenerator(db)

	// Add Redis version info
	generator.AddAuxField("redis-ver", "6.0.0")
	generator.AddAuxField("redis-bits", "64")
	generator.AddAuxField("ctime", fmt.Sprintf("%d", time.Now().Unix()))

	if err := generator.Generate(writer); err != nil {
		return fmt.Errorf("failed to generate RDB: %w", err)
	}

	return nil
}

// RDBLoaderImpl implements replication.RDBLoader interface
type RDBLoaderImpl struct{}

// LoadRDBFromBytes loads RDB data from bytes into database
func (l *RDBLoaderImpl) LoadRDBFromBytes(db interface{}, data []byte) error {
	// Type assert to *database.DB
	dbTyped, ok := db.(*database.DB)
	if !ok {
		return fmt.Errorf("invalid database type")
	}
	return LoadFromBytes(dbTyped, data)
}
