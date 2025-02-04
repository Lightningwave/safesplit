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

	log.Printf("Deriving KEK - Salt: %x", salt)
	log.Printf("Password length: %d bytes", len(password))

	kek := pbkdf2.Key([]byte(password), salt, PBKDF2Iterations, KeyEncryptionSize, sha256.New)
	log.Printf("Derived KEK: %x", kek)

	return kek, nil
}

// EncryptMasterKey encrypts using AES-GCM with 16-byte nonce
func EncryptMasterKey(data []byte, key []byte, nonce []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("invalid key length: expected 32, got %d", len(key))
	}
	if len(nonce) != 16 {
		return nil, fmt.Errorf("invalid nonce length: expected 16, got %d", len(nonce))
	}

	log.Printf("EncryptMasterKey - Data length: %d", len(data))
	log.Printf("Key: %x", key)
	log.Printf("Nonce: %x", nonce)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	encrypted := gcm.Seal(nil, nonce, data, nil)
	log.Printf("Encrypted result - Length: %d, Value: %x", len(encrypted), encrypted)
	return encrypted, nil
}

// DecryptMasterKey decrypts using AES-GCM with 16-byte nonce
func DecryptMasterKey(encryptedKey []byte, key []byte, nonce []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("invalid key length: expected 32, got %d", len(key))
	}
	if len(nonce) != 16 {
		return nil, fmt.Errorf("invalid nonce length: expected 16, got %d", len(nonce))
	}

	// Take first 48 bytes for decryption
	encryptedKey = encryptedKey[:48]

	log.Printf("Original Encrypted Master Key Length: %d", len(encryptedKey))
	log.Printf("Original Encrypted Master Key: %x", encryptedKey)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	log.Printf("DecryptMasterKey Debug:")
	log.Printf("- Encrypted Key Len: %d", len(encryptedKey))
	log.Printf("- KEK: %x", key)
	log.Printf("- Nonce: %x", nonce)
	log.Printf("- GCM NonceSize: %d", gcm.NonceSize())

	decrypted, err := gcm.Open(nil, nonce, encryptedKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return decrypted, nil
}
