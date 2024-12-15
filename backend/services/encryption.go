package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

type EncryptionService struct {
	shamirService *ShamirService
}

func NewEncryptionService(shamirService *ShamirService) *EncryptionService {
	return &EncryptionService{
		shamirService: shamirService,
	}
}

func (s *EncryptionService) EncryptFile(data []byte, n, k int) (encrypted []byte, iv []byte, salt []byte, shares []KeyShare, err error) {
	// Generate encryption key
	key := make([]byte, 32) // 256-bit key
	if _, err := rand.Read(key); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to generate key: %w", err)
	}

	// Split key into shares using Shamir's Secret Sharing
	shares, err = s.shamirService.SplitKey(key, n, k)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to split key: %w", err)
	}

	// Generate salt
	salt = make([]byte, 32) // 256-bit salt
	if _, err := rand.Read(salt); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCMWithNonceSize(block, 16) // Force 16-byte nonce size
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate IV of fixed size (16 bytes)
	iv = make([]byte, 16)
	if _, err := rand.Read(iv); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	// Encrypt the data
	encrypted = gcm.Seal(nil, iv, data, nil)
	return encrypted, iv, salt, shares, nil
}

func (s *EncryptionService) DecryptFile(encrypted []byte, iv []byte, keyShares []KeyShare, k int) ([]byte, error) {
	// Reconstruct key from shares
	key, err := s.shamirService.RecombineKey(keyShares, k)
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct key: %w", err)
	}

	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCMWithNonceSize(block, 16) // Force 16-byte nonce size
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Check that IV length matches 16 bytes
	if len(iv) != 16 {
		return nil, fmt.Errorf("invalid IV length: got %d, expected 16", len(iv))
	}

	// Decrypt the data
	decrypted, err := gcm.Open(nil, iv, encrypted, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return decrypted, nil
}

// EncryptKeyFragment now works with []byte
func (s *EncryptionService) EncryptKeyFragment(fragment []byte, password []byte) ([]byte, error) {
	// Generate a random salt for PBKDF2
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive key using PBKDF2
	key := pbkdf2.Key(password, salt, 4096, 32, sha256.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the fragment
	ciphertext := gcm.Seal(nil, nonce, fragment, nil)

	// Combine salt + nonce + ciphertext
	result := make([]byte, len(salt)+len(nonce)+len(ciphertext))
	copy(result[0:32], salt)
	copy(result[32:48], nonce)
	copy(result[48:], ciphertext)

	return result, nil
}

// DecryptKeyFragment now works with []byte
func (s *EncryptionService) DecryptKeyFragment(encryptedFragment []byte, password []byte) ([]byte, error) {
	// Check minimum length (32 bytes salt + 16 bytes nonce + at least 1 byte data)
	if len(encryptedFragment) < 49 {
		return nil, fmt.Errorf("encrypted fragment too short")
	}

	// Extract salt and nonce
	salt := encryptedFragment[0:32]
	nonce := encryptedFragment[32:48]
	ciphertext := encryptedFragment[48:]

	// Derive key using PBKDF2 with the same parameters
	key := pbkdf2.Key(password, salt, 4096, 32, sha256.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt the fragment
	decrypted, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt fragment: %w", err)
	}

	return decrypted, nil
}
