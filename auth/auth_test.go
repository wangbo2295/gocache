package auth

import (
	"sync"
	"testing"
)

func TestNewAuthenticator(t *testing.T) {
	auth := NewAuthenticator()

	if auth == nil {
		t.Fatal("NewAuthenticator returned nil")
	}

	if auth.IsEnabled() {
		t.Error("New authenticator should not be enabled by default")
	}

	if auth.GetPassword() != "" {
		t.Error("New authenticator should have empty password")
	}
}

func TestSetPassword(t *testing.T) {
	auth := NewAuthenticator()

	// Test setting password
	auth.SetPassword("testpassword")

	if !auth.IsEnabled() {
		t.Error("Authenticator should be enabled after setting password")
	}

	if auth.GetPassword() != "testpassword" {
		t.Errorf("Expected password 'testpassword', got '%s'", auth.GetPassword())
	}

	// Test changing password
	auth.SetPassword("newpassword")

	if auth.GetPassword() != "newpassword" {
		t.Errorf("Expected password 'newpassword', got '%s'", auth.GetPassword())
	}
}

func TestAuthenticate(t *testing.T) {
	auth := NewAuthenticator()

	// Test authentication when not enabled
	if !auth.Authenticate("anypassword") {
		t.Error("Should return true when authentication is not enabled")
	}

	// Enable authentication
	auth.SetPassword("correctpassword")

	// Test correct password
	if !auth.Authenticate("correctpassword") {
		t.Error("Should return true for correct password")
	}

	// Test incorrect password
	if auth.Authenticate("wrongpassword") {
		t.Error("Should return false for incorrect password")
	}

	// Test empty password
	if auth.Authenticate("") {
		t.Error("Should return false for empty password")
	}
}

func TestIsAuthenticated(t *testing.T) {
	auth := NewAuthenticator()

	clientID := "client123"

	// Test when auth is not enabled
	if !auth.IsAuthenticated(clientID) {
		t.Error("Should return true when authentication is not enabled")
	}

	// Enable authentication
	auth.SetPassword("password")

	// Test unauthenticated client
	if auth.IsAuthenticated(clientID) {
		t.Error("Client should not be authenticated initially")
	}

	// Mark client as authenticated
	auth.MarkAuthenticated(clientID)

	// Test authenticated client
	if !auth.IsAuthenticated(clientID) {
		t.Error("Client should be authenticated after marking")
	}

	// Test non-existent client
	anotherClient := "client456"
	if auth.IsAuthenticated(anotherClient) {
		t.Error("Non-existent client should not be authenticated")
	}
}

func TestMarkAuthenticated(t *testing.T) {
	auth := NewAuthenticator()
	auth.SetPassword("password")

	clientID := "client789"

	// Mark as authenticated
	auth.MarkAuthenticated(clientID)

	if !auth.IsAuthenticated(clientID) {
		t.Error("Client should be authenticated after marking")
	}

	// Mark again (idempotent)
	auth.MarkAuthenticated(clientID)

	if !auth.IsAuthenticated(clientID) {
		t.Error("Client should still be authenticated")
	}
}

func TestLogout(t *testing.T) {
	auth := NewAuthenticator()
	auth.SetPassword("password")

	clientID := "client999"

	// Mark as authenticated
	auth.MarkAuthenticated(clientID)

	if !auth.IsAuthenticated(clientID) {
		t.Error("Client should be authenticated")
	}

	// Logout
	auth.Logout(clientID)

	if auth.IsAuthenticated(clientID) {
		t.Error("Client should not be authenticated after logout")
	}

	// Logout again (should be safe)
	auth.Logout(clientID)

	if auth.IsAuthenticated(clientID) {
		t.Error("Client should still not be authenticated")
	}
}

func TestMultipleClients(t *testing.T) {
	auth := NewAuthenticator()
	auth.SetPassword("password")

	clients := []string{"client1", "client2", "client3"}

	// Mark all clients as authenticated
	for _, client := range clients {
		auth.MarkAuthenticated(client)
	}

	// Verify all are authenticated
	for _, client := range clients {
		if !auth.IsAuthenticated(client) {
			t.Errorf("Client %s should be authenticated", client)
		}
	}

	// Logout one client
	auth.Logout("client2")

	// Verify only client2 is logged out
	if !auth.IsAuthenticated("client1") {
		t.Error("client1 should still be authenticated")
	}

	if auth.IsAuthenticated("client2") {
		t.Error("client2 should not be authenticated")
	}

	if !auth.IsAuthenticated("client3") {
		t.Error("client3 should still be authenticated")
	}
}

func TestConcurrentAccess(t *testing.T) {
	auth := NewAuthenticator()
	auth.SetPassword("password")

	var wg sync.WaitGroup
	numOperations := 1000

	// Concurrent authentications
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				clientID := string(rune('a' + j%26))
				auth.MarkAuthenticated(clientID)
				auth.IsAuthenticated(clientID)
				auth.Logout(clientID)
			}
		}()
	}

	wg.Wait()

	// Verify authenticator is still in consistent state
	if !auth.IsEnabled() {
		t.Error("Authenticator should still be enabled")
	}

	if auth.GetPassword() != "password" {
		t.Error("Password should not have changed")
	}
}

func TestHashPassword(t *testing.T) {
	// Test that same password produces same hash
	auth1 := NewAuthenticator()
	auth1.SetPassword("testpass")

	auth2 := NewAuthenticator()
	auth2.SetPassword("testpass")

	// Both should accept the same password
	if !auth1.Authenticate("testpass") {
		t.Error("auth1 should accept correct password")
	}

	if !auth2.Authenticate("testpass") {
		t.Error("auth2 should accept correct password")
	}

	// Different passwords should produce different hashes
	auth3 := NewAuthenticator()
	auth3.SetPassword("differentpass")

	if auth1.Authenticate("differentpass") {
		t.Error("Should not accept different password")
	}
}

func TestPasswordSecurity(t *testing.T) {
	auth := NewAuthenticator()

	password := "mypassword123"
	auth.SetPassword(password)

	// Verify password is stored (not just the hash)
	if auth.GetPassword() != password {
		t.Error("Password should be stored for comparison")
	}

	// Verify authentication works
	if !auth.Authenticate(password) {
		t.Error("Should authenticate with correct password")
	}

	// Verify wrong hash doesn't work
	if auth.Authenticate("wrongpass") {
		t.Error("Should not authenticate with wrong password")
	}
}

func TestEdgeCases(t *testing.T) {
	auth := NewAuthenticator()

	// Empty password
	auth.SetPassword("")
	if !auth.Authenticate("") {
		t.Error("Should authenticate with empty password when set")
	}

	// Special characters in password
	specialPassword := "!@#$%^&*()_+-=[]{}|;':\",./<>?"
	auth.SetPassword(specialPassword)

	if !auth.Authenticate(specialPassword) {
		t.Error("Should handle special characters in password")
	}

	// Very long password
	longPassword := string(make([]byte, 1000))
	for i := range longPassword {
		longPassword = longPassword[:i] + "a" + longPassword[i+1:]
	}
	auth.SetPassword(longPassword)

	if !auth.Authenticate(longPassword) {
		t.Error("Should handle long passwords")
	}
}

func TestDisableAuth(t *testing.T) {
	auth := NewAuthenticator()

	// Set password
	auth.SetPassword("password")
	if !auth.IsEnabled() {
		t.Error("Should be enabled")
	}

	// Mark a client as authenticated
	auth.MarkAuthenticated("client1")
	if !auth.IsAuthenticated("client1") {
		t.Error("Client should be authenticated")
	}

	// Note: There's no explicit Disable method, but we can test
	// that creating a new authenticator starts disabled
	auth2 := NewAuthenticator()
	if auth2.IsEnabled() {
		t.Error("New authenticator should be disabled")
	}

	if !auth2.IsAuthenticated("any_client") {
		t.Error("Should return true for any client when disabled")
	}
}
