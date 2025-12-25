package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/codeneuss/lampcontrol/internal/domain"
	"golang.org/x/crypto/pbkdf2"
)

// TwitchStorage handles persistent storage of Twitch configuration
type TwitchStorage struct {
	filePath string
	encKey   []byte
	mu       sync.RWMutex
	config   *domain.TwitchConfig
}

// NewTwitchStorage creates a new Twitch storage instance
func NewTwitchStorage() (*TwitchStorage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".lampcontrol")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	filePath := filepath.Join(configDir, "twitch_config.json")

	// Generate encryption key from machine-specific data
	encKey := generateEncryptionKey()

	storage := &TwitchStorage{
		filePath: filePath,
		encKey:   encKey,
		config:   domain.NewTwitchConfig(),
	}

	// Load existing config
	if err := storage.load(); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	return storage, nil
}

// Get returns the current Twitch configuration
func (s *TwitchStorage) Get() *domain.TwitchConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.config
}

// Save saves the Twitch configuration
func (s *TwitchStorage) Save(config *domain.TwitchConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = config
	return s.persist()
}

// encrypt encrypts sensitive data
func (s *TwitchStorage) encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher(s.encKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts sensitive data
func (s *TwitchStorage) decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, cipherBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, cipherBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// persist saves config to file
func (s *TwitchStorage) persist() error {
	// Create a copy for encryption
	encConfig := *s.config

	// Encrypt sensitive fields
	var err error
	encConfig.AccessToken, err = s.encrypt(s.config.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to encrypt access token: %w", err)
	}

	encConfig.RefreshToken, err = s.encrypt(s.config.RefreshToken)
	if err != nil {
		return fmt.Errorf("failed to encrypt refresh token: %w", err)
	}

	data, err := json.MarshalIndent(encConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(s.filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// load loads config from file
func (s *TwitchStorage) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	var encConfig domain.TwitchConfig
	if err := json.Unmarshal(data, &encConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Decrypt sensitive fields
	encConfig.AccessToken, err = s.decrypt(encConfig.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to decrypt access token: %w", err)
	}

	encConfig.RefreshToken, err = s.decrypt(encConfig.RefreshToken)
	if err != nil {
		return fmt.Errorf("failed to decrypt refresh token: %w", err)
	}

	s.config = &encConfig
	return nil
}

// generateEncryptionKey generates a machine-specific encryption key
func generateEncryptionKey() []byte {
	// Use hostname as salt for machine-specific key
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "lampcontrol-default"
	}

	// Derive key using PBKDF2
	return pbkdf2.Key([]byte(hostname), []byte("lampcontrol-twitch"), 100000, 32, sha256.New)
}
