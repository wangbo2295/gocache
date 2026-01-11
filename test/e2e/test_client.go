package e2e

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// RespType represents Redis RESP protocol data types
type RespType int

const (
	SimpleString RespType = iota
	Error
	Integer
	BulkString
	Array
)

// Reply represents a Redis protocol reply
type Reply struct {
	Type    RespType
	Data    interface{}
	Error   error
}

// TestClient is a Redis protocol client for testing
type TestClient struct {
	conn       net.Conn
	reader     *bufio.Reader
	addr       string
	timeout    time.Duration
	retries    int
}

// NewTestClient creates a new test client
func NewTestClient(addr string) *TestClient {
	return &TestClient{
		addr:    addr,
		timeout: 5 * time.Second,
		retries: 3,
	}
}

// Connect connects to the Redis server
func (c *TestClient) Connect() error {
	var err error
	for i := 0; i < c.retries; i++ {
		c.conn, err = net.DialTimeout("tcp", c.addr, c.timeout)
		if err == nil {
			c.reader = bufio.NewReader(c.conn)
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("failed to connect after %d retries: %v", c.retries, err)
}

// Close closes the connection
func (c *TestClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Reconnect reconnects to the server
func (c *TestClient) Reconnect() error {
	if c.conn != nil {
		c.conn.Close()
	}
	return c.Connect()
}

// Send sends a command and returns the reply
func (c *TestClient) Send(cmd string, args ...string) (*Reply, error) {
	if c.conn == nil {
		return nil, errors.New("not connected")
	}

	// Build command array
	cmdArray := make([]interface{}, 0, len(args)+1)
	cmdArray = append(cmdArray, cmd)
	for _, arg := range args {
		cmdArray = append(cmdArray, arg)
	}

	// Send command
	data := encodeArray(cmdArray)
	_, err := c.conn.Write([]byte(data))
	if err != nil {
		return nil, fmt.Errorf("failed to send command: %v", err)
	}

	// Read reply
	return c.readReply()
}

// SendBytes sends a command with byte arguments
func (c *TestClient) SendBytes(cmd [][]byte) (*Reply, error) {
	if c.conn == nil {
		return nil, errors.New("not connected")
	}

	if len(cmd) == 0 {
		return nil, errors.New("empty command")
	}

	// Convert to interface array
	cmdArray := make([]interface{}, len(cmd))
	for i, arg := range cmd {
		cmdArray[i] = string(arg)
	}

	// Send command
	data := encodeArray(cmdArray)
	_, err := c.conn.Write([]byte(data))
	if err != nil {
		return nil, fmt.Errorf("failed to send command: %v", err)
	}

	// Read reply
	return c.readReply()
}

// Execute sends a command and returns the reply as string array
func (c *TestClient) Execute(cmd string, args ...string) ([]string, error) {
	reply, err := c.Send(cmd, args...)
	if err != nil {
		return nil, err
	}

	if reply.Error != nil {
		return nil, reply.Error
	}

	return reply.ToStringArray(), nil
}

// ExecuteBytes sends a command with byte arguments
func (c *TestClient) ExecuteBytes(cmd [][]byte) ([]string, error) {
	reply, err := c.SendBytes(cmd)
	if err != nil {
		return nil, err
	}

	if reply.Error != nil {
		return nil, reply.Error
	}

	return reply.ToStringArray(), nil
}

// readReply reads a reply from the server
func (c *TestClient) readReply() (*Reply, error) {
	line, err := readLine(c.reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read reply: %v", err)
	}

	if len(line) == 0 {
		return nil, errors.New("empty reply")
	}

	switch line[0] {
	case '+': // Simple String
		return &Reply{
			Type: SimpleString,
			Data: string(line[1:]),
		}, nil

	case '-': // Error
		errMsg := string(line[1:])
		return &Reply{
			Type:  Error,
			Data:  errMsg,
			Error: errors.New(errMsg),
		}, errors.New(errMsg)

	case ':': // Integer
		val, err := strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid integer: %v", err)
		}
		return &Reply{
			Type: Integer,
			Data: val,
		}, nil

	case '$': // Bulk String
		size, err := strconv.Atoi(string(line[1:]))
		if err != nil {
			return nil, fmt.Errorf("invalid bulk string length: %v", err)
		}

		if size < 0 {
			// Null bulk string
			return &Reply{
				Type: BulkString,
				Data: nil,
			}, nil
		}

		// Read the data
		data := make([]byte, size)
		_, err = c.reader.Read(data)
		if err != nil {
			return nil, fmt.Errorf("failed to read bulk string data: %v", err)
		}

		// Read trailing \r\n
		c.reader.ReadLine()

		return &Reply{
			Type: BulkString,
			Data: string(data),
		}, nil

	case '*': // Array
		count, err := strconv.Atoi(string(line[1:]))
		if err != nil {
			return nil, fmt.Errorf("invalid array count: %v", err)
		}

		if count < 0 {
			// Null array
			return &Reply{
				Type: Array,
				Data: nil,
			}, nil
		}

		// Read array elements
		elements := make([]interface{}, count)
		for i := 0; i < count; i++ {
			reply, err := c.readReply()
			if err != nil {
				return nil, err
			}
			elements[i] = reply.Data
		}

		return &Reply{
			Type: Array,
			Data: elements,
		}, nil

	default:
		return nil, fmt.Errorf("unknown reply type: %c", line[0])
	}
}

// readLine reads a line ending with \r\n
func readLine(r *bufio.Reader) ([]byte, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}

	if len(line) < 2 || line[len(line)-2] != '\r' {
		return nil, errors.New("invalid line ending")
	}

	return []byte(line[:len(line)-2]), nil
}

// encodeArray encodes an array for RESP protocol
func encodeArray(items []interface{}) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("*%d\r\n", len(items)))

	for _, item := range items {
		switch v := item.(type) {
		case string:
			sb.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))
		case []byte:
			sb.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))
		case int:
			s := strconv.Itoa(v)
			sb.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(s), s))
		case int64:
			s := strconv.FormatInt(v, 10)
			sb.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(s), s))
		default:
			s := fmt.Sprintf("%v", v)
			sb.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(s), s))
		}
	}

	return sb.String()
}

// ToStringArray converts reply data to string array
func (r *Reply) ToStringArray() []string {
	if r.Data == nil {
		return []string{"(nil)"}
	}

	switch v := r.Data.(type) {
	case string:
		if v == "" {
			return []string{""}
		}
		return []string{v}

	case int64:
		return []string{strconv.FormatInt(v, 10)}

	case []interface{}:
		result := make([]string, len(v))
		for i, elem := range v {
			if elem == nil {
				result[i] = "(nil)"
			} else {
				result[i] = fmt.Sprintf("%v", elem)
			}
		}
		return result

	default:
		return []string{fmt.Sprintf("%v", v)}
	}
}

// IsOK checks if reply is OK
func (r *Reply) IsOK() bool {
	return r.Type == SimpleString && r.Data == "OK"
}

// IsError checks if reply is an error
func (r *Reply) IsError() bool {
	return r.Type == Error
}

// IsNil checks if reply is nil
func (r *Reply) IsNil() bool {
	return r.Data == nil
}

// GetInt gets integer value from reply
func (r *Reply) GetInt() (int64, error) {
	if r.Type != Integer {
		return 0, fmt.Errorf("expected integer, got %v", r.Type)
	}
	return r.Data.(int64), nil
}

// GetString gets string value from reply
func (r *Reply) GetString() string {
	if r.Data == nil {
		return ""
	}

	// Handle different data types
	switch v := r.Data.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case int64:
		return fmt.Sprintf("%d", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// GetArray gets array value from reply
func (r *Reply) GetArray() []interface{} {
	if r.Data == nil {
		return nil
	}
	arr, ok := r.Data.([]interface{})
	if !ok {
		return nil
	}
	return arr
}

// SetTimeout sets connection timeout
func (c *TestClient) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// SetRetries sets retry count
func (c *TestClient) SetRetries(retries int) {
	c.retries = retries
}
