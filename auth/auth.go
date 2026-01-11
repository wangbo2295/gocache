package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
)

// Authenticator handles authentication
type Authenticator struct {
	password       string
	passwordHash   string
	authenticated   map[string]bool // client ID -> authenticated
	enabled        bool
	mu             sync.RWMutex
}

// NewAuthenticator creates a new authenticator
func NewAuthenticator() *Authenticator {
	return &Authenticator{
		authenticated: make(map[string]bool),
		enabled:       false,
	}
}

// SetPassword sets the password and enables authentication
func (a *Authenticator) SetPassword(password string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.password = password
	a.passwordHash = hashPassword(password)
	a.enabled = true
}

// GetPassword returns the current password
func (a *Authenticator) GetPassword() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.password
}

// IsEnabled returns true if authentication is enabled
func (a *Authenticator) IsEnabled() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.enabled
}

// Authenticate checks if the given password is correct
func (a *Authenticator) Authenticate(password string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.enabled {
		return true // No auth required
	}

	return hashPassword(password) == a.passwordHash
}

// IsAuthenticated checks if a client is authenticated
func (a *Authenticator) IsAuthenticated(clientID string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.enabled {
		return true // No auth required
	}

	return a.authenticated[clientID]
}

// MarkAuthenticated marks a client as authenticated
func (a *Authenticator) MarkAuthenticated(clientID string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.authenticated[clientID] = true
}

// Logout removes authentication for a client
func (a *Authenticator) Logout(clientID string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	delete(a.authenticated, clientID)
}

// hashPassword hashes a password using SHA-256
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}
