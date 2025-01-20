package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
    "log"
	"golang.org/x/crypto/pbkdf2"
)

const (
	MasterKeySize     = 32 
	PBKDF2Iterations  = 100000
	KeyEncryptionSize = 32 
)

// GenerateMasterKey generates a new random master key
func GenerateMasterKey() ([]byte, error) {
	masterKey := make([]byte, MasterKeySize)
	if _, err := rand.Read(masterKey); err != nil {
		return nil, fmt.Errorf("failed to generate master key: %w", err)
	}
	return masterKey, nil
}

// DeriveKeyEncryptionKey derives a key encryption key from password using PBKDF2
func DeriveKeyEncryptionKey(password string, salt []byte) ([]byte, error) {
	if len(salt) != 32 {
		return nil, fmt.Errorf("invalid salt length: expected 32, got %d", len(salt))
	}

	kek := pbkdf2.Key([]byte(password), salt, PBKDF2Iterations, KeyEncryptionSize, sha256.New)
	return kek, nil
}

// EncryptMasterKey encrypts the master key using AES-GCM
func EncryptMasterKey(data []byte, key []byte, nonce []byte) ([]byte, error) {
    // Key should be exactly 32 bytes
    if len(key) != 32 {
        return nil, fmt.Errorf("invalid master key length: expected 32, got %d", len(key))
    }

    // Nonce should be exactly 12 bytes for GCM
    if len(nonce) != 12 {
        nonce = nonce[:12]
    }

    // Log input parameters
    log.Printf("EncryptMasterKey input - Data length: %d, Key length: %d, Nonce length: %d",
        len(data), len(key), len(nonce))
    log.Printf("Data hex: %x", data)
    log.Printf("Key hex: %x", key)
    log.Printf("Nonce hex: %x", nonce)

    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, fmt.Errorf("failed to create cipher: %w", err)
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, fmt.Errorf("failed to create GCM: %w", err)
    }

    // Encrypt data
    encryptedData := gcm.Seal(nil, nonce, data, nil)
    log.Printf("Encrypted result length: %d bytes, Value: %x", len(encryptedData), encryptedData)

    return encryptedData, nil
}

// DecryptMasterKey decrypts the master key using AES-GCM
func DecryptMasterKey(encryptedKey []byte, kek []byte, nonce []byte) ([]byte, error) {
	if len(kek) != KeyEncryptionSize {
		return nil, fmt.Errorf("invalid key encryption key length: expected %d, got %d", KeyEncryptionSize, len(kek))
	}
	if len(nonce) != 12 {
		return nil, fmt.Errorf("invalid nonce length: expected 12, got %d", len(nonce))
	}

	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt master key
	masterKey, err := gcm.Open(nil, nonce, encryptedKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt master key: %w", err)
	}

	return masterKey, nil
}
