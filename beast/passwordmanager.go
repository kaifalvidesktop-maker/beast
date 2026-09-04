package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"strings"
	"sync"
	"time"
)

// ---------------------------------------------------
// ENCRYPTED PASSWORD MANAGER
// Master password derives an AES-256 key (via SHA-256).
// Nothing is ever written to disk — vault lives in RAM only
// and is wiped when BEAST closes or the vault is locked.
// ---------------------------------------------------

type PasswordEntry struct {
	ID        int
	Domain    string
	Username  string
	encrypted []byte // AES-GCM ciphertext, never exposed directly
	nonce     []byte
	CreatedAt time.Time
}

type PasswordVault struct {
	mu           sync.Mutex
	Entries      []*PasswordEntry
	NextID       int
	masterKey    []byte // derived key, held only while unlocked
	Unlocked     bool
	FailedTries  int
}

var passwordVault = &PasswordVault{
	Entries: []*PasswordEntry{},
	NextID:  1,
}

// deriveKey turns a plain-text master password into a 32-byte AES key
func deriveKey(masterPassword string) []byte {
	sum := sha256.Sum256([]byte(masterPassword))
	return sum[:]
}

// Unlock sets the vault's active key from the master password.
// It doesn't "verify" against anything stored (nothing is stored across
// restarts), so the FIRST unlock in a session establishes the key.
func (pv *PasswordVault) Unlock(masterPassword string) bool {
	pv.mu.Lock()
	defer pv.mu.Unlock()

	if strings.TrimSpace(masterPassword) == "" {
		return false
	}

	pv.masterKey = deriveKey(masterPassword)
	pv.Unlocked = true
	pv.FailedTries = 0
	return true
}

// Lock clears the active key from memory (entries remain but unreadable
// until unlocked again with the correct password)
func (pv *PasswordVault) Lock() {
	pv.mu.Lock()
	defer pv.mu.Unlock()
	pv.masterKey = nil
	pv.Unlocked = false
}

// IsUnlocked reports current lock state
func (pv *PasswordVault) IsUnlocked() bool {
	pv.mu.Lock()
	defer pv.mu.Unlock()
	return pv.Unlocked
}

// encrypt uses AES-GCM to encrypt plaintext with the current master key
func (pv *PasswordVault) encrypt(plaintext string) ([]byte, []byte, error) {
	block, err := aes.NewCipher(pv.masterKey)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	return ciphertext, nonce, nil
}

// decrypt reverses encrypt() using the current master key
func (pv *PasswordVault) decrypt(ciphertext []byte, nonce []byte) (string, error) {
	block, err := aes.NewCipher(pv.masterKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// Save adds a new encrypted password entry. Vault must be unlocked.
func (pv *PasswordVault) Save(domain string, username string, password string) (*PasswordEntry, error) {
	pv.mu.Lock()
	defer pv.mu.Unlock()

	if !pv.Unlocked {
		return nil, errors.New("vault is locked")
	}

	ciphertext, nonce, err := pv.encrypt(password)
	if err != nil {
		return nil, err
	}

	entry := &PasswordEntry{
		ID:        pv.NextID,
		Domain:    strings.ToLower(strings.TrimSpace(domain)),
		Username:  username,
		encrypted: ciphertext,
		nonce:     nonce,
		CreatedAt: time.Now(),
	}

	pv.Entries = append(pv.Entries, entry)
	pv.NextID++
	return entry, nil
}

// Reveal decrypts and returns the plaintext password for one entry.
// This is the ONLY function that ever exposes a decrypted password.
func (pv *PasswordVault) Reveal(id int) (string, error) {
	pv.mu.Lock()
	defer pv.mu.Unlock()

	if !pv.Unlocked {
		return "", errors.New("vault is locked")
	}

	for _, e := range pv.Entries {
		if e.ID == id {
			return pv.decrypt(e.encrypted, e.nonce)
		}
	}
	return "", errors.New("entry not found")
}

// Remove deletes a password entry permanently
func (pv *PasswordVault) Remove(id int) bool {
	pv.mu.Lock()
	defer pv.mu.Unlock()

	newList := []*PasswordEntry{}
	removed := false
	for _, e := range pv.Entries {
		if e.ID == id {
			removed = true
			continue
		}
		newList = append(newList, e)
	}
	pv.Entries = newList
	return removed
}

// GetMetadataOnly returns entries WITHOUT decrypted passwords —
// safe to send to the UI for listing (domain + username only)
type PasswordMetadata struct {
	ID       int
	Domain   string
	Username string
}

func (pv *PasswordVault) GetMetadataOnly() []PasswordMetadata {
	pv.mu.Lock()
	defer pv.mu.Unlock()

	result := []PasswordMetadata{}
	for _, e := range pv.Entries {
		result = append(result, PasswordMetadata{
			ID:       e.ID,
			Domain:   e.Domain,
			Username: e.Username,
		})
	}
	return result
}

// FindForDomain returns metadata for entries matching a domain (for autofill prompt)
func (pv *PasswordVault) FindForDomain(domain string) []PasswordMetadata {
	pv.mu.Lock()
	defer pv.mu.Unlock()

	domain = strings.ToLower(domain)
	result := []PasswordMetadata{}
	for _, e := range pv.Entries {
		if e.Domain == domain {
			result = append(result, PasswordMetadata{ID: e.ID, Domain: e.Domain, Username: e.Username})
		}
	}
	return result
}

// WipeAll destroys the entire vault (used by "Clear All Data" in settings)
func (pv *PasswordVault) WipeAll() {
	pv.mu.Lock()
	defer pv.mu.Unlock()
	pv.Entries = []*PasswordEntry{}
	pv.masterKey = nil
	pv.Unlocked = false
}

// GeneratePassword creates a strong random password
func GenerateStrongPassword(length int) string {
	if length < 8 {
		length = 16
	}
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	result := make([]byte, length)
	randomBytes := make([]byte, length)
	rand.Read(randomBytes)
	for i, b := range randomBytes {
		result[i] = charset[int(b)%len(charset)]
	}
	return string(result)
}