package rdb

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"unsafe"

	"github.com/wangbo/gocache/database"
)

// Loader loads database from RDB file
type Loader struct {
	input io.Reader
	db    *database.DB
}

// MakeLoader creates a new RDB loader
func MakeLoader(db *database.DB) *Loader {
	return &Loader{
		db: db,
	}
}

// LoadFromFile loads database from RDB file
func LoadFromFile(db *database.DB, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open RDB file: %w", err)
	}
	defer file.Close()

	loader := MakeLoader(db)
	loader.input = file

	return loader.Load()
}

// LoadFromBytes loads database from RDB bytes
func LoadFromBytes(db *database.DB, data []byte) error {
	reader := bytes.NewReader(data)
	loader := MakeLoader(db)
	loader.input = reader

	return loader.Load()
}

// Load reads and parses the RDB file
func (l *Loader) Load() error {
	// Read header
	if err := l.readHeader(); err != nil {
		return fmt.Errorf("read header: %w", err)
	}

	// Read entries until EOF
	for {
		opcode, err := l.readByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("read opcode: %w", err)
		}

		switch opcode {
		case OpcodeEOF:
			// Read and verify checksum
			if err := l.readChecksum(); err != nil {
				return fmt.Errorf("read checksum: %w", err)
			}
			return nil
		case OpcodeSelectDB:
			dbID, err := l.readLength()
			if err != nil {
				return fmt.Errorf("read db id: %w", err)
			}
			_ = dbID // We only support single DB for now
		case OpcodeAux:
			if err := l.readAuxField(); err != nil {
				return fmt.Errorf("read aux field: %w", err)
			}
		case OpcodeExpireTimeMS:
			expiryMS, err := l.readExpireTimeMS()
			if err != nil {
				return fmt.Errorf("read expire time: %w", err)
			}
			// Next value will have this expiry
			// For simplicity, we skip expiry handling for now
			_ = expiryMS
		case TypeString:
			if err := l.readStringValue(); err != nil {
				return fmt.Errorf("read string value: %w", err)
			}
		case TypeHash:
			if err := l.readHashValue(); err != nil {
				return fmt.Errorf("read hash value: %w", err)
			}
		case TypeList:
			if err := l.readListValue(); err != nil {
				return fmt.Errorf("read list value: %w", err)
			}
		case TypeSet:
			if err := l.readSetValue(); err != nil {
				return fmt.Errorf("read set value: %w", err)
			}
		case TypeZSet:
			if err := l.readZSetValue(); err != nil {
				return fmt.Errorf("read zset value: %w", err)
			}
		default:
			return fmt.Errorf("unknown opcode: %d", opcode)
		}
	}

	return nil
}

// readHeader reads and verifies the RDB header
func (l *Loader) readHeader() error {
	magic := make([]byte, 5)
	if _, err := io.ReadFull(l.input, magic); err != nil {
		return err
	}

	if string(magic) != RedisMagicString {
		return fmt.Errorf("invalid RDB file: bad magic string")
	}

	version := make([]byte, 4)
	if _, err := io.ReadFull(l.input, version); err != nil {
		return err
	}

	// Verify version is reasonable
	if version[0] > RDBVersion {
		return fmt.Errorf("unsupported RDB version: %d", version[0])
	}

	return nil
}

// readChecksum reads and verifies the CRC64 checksum
func (l *Loader) readChecksum() error {
	checksum := make([]byte, 8)
	if _, err := io.ReadFull(l.input, checksum); err != nil {
		return err
	}
	// For simplicity, we skip checksum verification
	return nil
}

// readAuxField reads an auxiliary field
func (l *Loader) readAuxField() error {
	key, err := l.readString()
	if err != nil {
		return err
	}
	_, err = l.readString()
	if err != nil {
		return err
	}
	// For simplicity, we ignore aux fields
	_ = key
	return nil
}

// readExpireTimeMS reads expire time in milliseconds
func (l *Loader) readExpireTimeMS() (int64, error) {
	expire := make([]byte, 8)
	if _, err := io.ReadFull(l.input, expire); err != nil {
		return 0, err
	}
	return int64(binary.LittleEndian.Uint64(expire)), nil
}

// readLength reads a length-encoded integer
func (l *Loader) readLength() (uint64, error) {
	b, err := l.readByte()
	if err != nil {
		return 0, err
	}

	encType := (b & 0xC0) >> 6
	length := uint64(b & 0x3F)

	switch encType {
	case Len6Bit:
		// For 6-bit encoding, length is stored in high 6 bits (shifted left by 2)
		// We need to shift right to get the actual length
		return length >> 2, nil
	case Len14Bit:
		b2, err := l.readByte()
		if err != nil {
			return 0, err
		}
		// Combine: 6 bits from first byte (shifted) + 8 bits from second byte
		// The 6 bits are also shifted, so we need to handle that
		result := (length << 8) | uint64(b2)
		return result >> 2, nil
	case Len32Bit:
		buf := make([]byte, 4)
		if _, err := io.ReadFull(l.input, buf); err != nil {
			return 0, err
		}
		return uint64(binary.LittleEndian.Uint32(buf)), nil
	case EncVal:
		// Special encoding - not implemented for now
		return 0, errors.New("special encoding not implemented")
	default:
		return 0, fmt.Errorf("unknown length encoding: %d", encType)
	}
}

// readString reads a string
func (l *Loader) readString() (string, error) {
	length, err := l.readLength()
	if err != nil {
		return "", err
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(l.input, data); err != nil {
		return "", err
	}

	return string(data), nil
}

// readStringEncoding reads a string with length encoding
func (l *Loader) readStringEncoding() ([]byte, error) {
	length, err := l.readLength()
	if err != nil {
		return nil, err
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(l.input, data); err != nil {
		return nil, err
	}

	return data, nil
}

// readStringValue reads a string value and stores it in database
func (l *Loader) readStringValue() error {
	key, err := l.readString()
	if err != nil {
		return err
	}

	value, err := l.readStringEncoding()
	if err != nil {
		return err
	}

	// Store in database
	l.db.ExecCommand("SET", key, string(value))
	return nil
}

// readHashValue reads a hash value and stores it in database
func (l *Loader) readHashValue() error {
	key, err := l.readString()
	if err != nil {
		return err
	}

	length, err := l.readLength()
	if err != nil {
		return err
	}

	// Build HMSET command
	args := []string{"HMSET", key}
	for i := uint64(0); i < length; i++ {
		field, err := l.readStringEncoding()
		if err != nil {
			return err
		}

		value, err := l.readStringEncoding()
		if err != nil {
			return err
		}

		args = append(args, string(field), string(value))
	}

	// Execute command
	cmdArgs := make([][]byte, len(args))
	for i, arg := range args {
		cmdArgs[i] = []byte(arg)
	}

	_, err = l.db.Exec(cmdArgs)
	return err
}

// readListValue reads a list value and stores it in database
func (l *Loader) readListValue() error {
	key, err := l.readString()
	if err != nil {
		return err
	}

	length, err := l.readLength()
	if err != nil {
		return err
	}

	// Build RPUSH command with all elements
	args := []string{"RPUSH", key}
	for i := uint64(0); i < length; i++ {
		elem, err := l.readStringEncoding()
		if err != nil {
			return err
		}
		args = append(args, string(elem))
	}

	// Execute command
	cmdArgs := make([][]byte, len(args))
	for i, arg := range args {
		cmdArgs[i] = []byte(arg)
	}

	_, err = l.db.Exec(cmdArgs)
	return err
}

// readSetValue reads a set value and stores it in database
func (l *Loader) readSetValue() error {
	key, err := l.readString()
	if err != nil {
		return err
	}

	length, err := l.readLength()
	if err != nil {
		return err
	}

	// Build SADD command with all members
	args := []string{"SADD", key}
	for i := uint64(0); i < length; i++ {
		member, err := l.readStringEncoding()
		if err != nil {
			return err
		}
		args = append(args, string(member))
	}

	// Execute command
	cmdArgs := make([][]byte, len(args))
	for i, arg := range args {
		cmdArgs[i] = []byte(arg)
	}

	_, err = l.db.Exec(cmdArgs)
	return err
}

// readZSetValue reads a sorted set value and stores it in database
func (l *Loader) readZSetValue() error {
	key, err := l.readString()
	if err != nil {
		return err
	}

	length, err := l.readLength()
	if err != nil {
		return err
	}

	// Build ZADD command with all member-score pairs
	args := []string{"ZADD", key}
	for i := uint64(0); i < length; i++ {
		score, err := l.readDouble()
		if err != nil {
			return err
		}

		member, err := l.readStringEncoding()
		if err != nil {
			return err
		}

		args = append(args, fmt.Sprintf("%f", score), string(member))
	}

	// Execute command
	cmdArgs := make([][]byte, len(args))
	for i, arg := range args {
		cmdArgs[i] = []byte(arg)
	}

	_, err = l.db.Exec(cmdArgs)
	return err
}

// readValue reads any value type
func (l *Loader) readValue() error {
	return errors.New("readValue not implemented")
}

// Helper functions for reading primitive types

func (l *Loader) readByte() (byte, error) {
	b := make([]byte, 1)
	_, err := io.ReadFull(l.input, b)
	return b[0], err
}

func (l *Loader) readDouble() (float64, error) {
	// Read double in little-endian format
	buf := make([]byte, 8)
	if _, err := io.ReadFull(l.input, buf); err != nil {
		return 0, err
	}

	bits := binary.LittleEndian.Uint64(buf)
	return float64frombits(bits), nil
}

func float64frombits(b uint64) float64 {
	return *(*float64)(unsafe.Pointer(&b))
}
