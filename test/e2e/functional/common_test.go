package functional

import (
	"testing"

	"github.com/wangbo/gocache/test/e2e"
)

const (
	defaultAddr = "127.0.0.1:16379"
)

// setupTestClient creates and connects a test client
func setupTestClient(t *testing.T) *e2e.TestClient {
	client := e2e.NewTestClient(defaultAddr)
	if err := client.Connect(); err != nil {
		t.Skipf("Failed to connect to server at %s: %v (skipping test)", defaultAddr, err)
	}

	// Authenticate if server requires password
	// Ignore auth errors (server might not require password)
	client.Send("AUTH", "yourpassword")

	// Ensure we're not in a transaction state from previous tests
	client.Send("DISCARD")

	return client
}
