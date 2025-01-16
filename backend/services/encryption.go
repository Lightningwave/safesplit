package services

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"

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
	log.Printf("Starting file encryption with n=%d, k=%d", n, k)

	// Generate encryption key (32 bytes for AES-256)
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to generate key: %w", err)
	}
	log.Printf("Generated encryption key: %x (length=%d)", key, len(key))

	// Split key into shares
	shares, err = s.shamirService.SplitKey(key, n, k)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to split key: %w", err)
	}
	log.Printf("Split key into %d shares (threshold: %d)", len(shares), k)

	// Debug: Log each share
	for i, share := range shares {
		log.Printf("Share %d: Index=%d, Value=%s (length=%d)",
			i, share.Index, share.Value, len(share.Value))
	}

	// Test reconstruction with minimum shares
	testShares := shares[:k]
	log.Printf("Testing reconstruction with first %d shares...", k)
	reconstructedKey, err := s.shamirService.RecombineKey(testShares, k)
	if err != nil {
		log.Printf("Test reconstruction failed: %v", err)
	} else {
		log.Printf("Test reconstruction result:")
		log.Printf("  Original key    : %x", key)
		log.Printf("  Reconstructed key: %x", reconstructedKey)
		log.Printf("  Keys match: %v", bytes.Equal(key, reconstructedKey))
	}

	// Generate IV (16 bytes for GCM)
	iv = make([]byte, 16)
	if _, err := rand.Read(iv); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to generate IV: %w", err)
	}
	log.Printf("Generated IV: %x", iv)

	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create cipher: %w", err)
	}
	log.Printf("Created AES cipher block. Block size: %d bytes", block.BlockSize())

	// Create GCM with 16-byte nonce size
	gcm, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}
	log.Printf("Created GCM with 16-byte nonce size")

	// Prepend the original data size to the data
	sizeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(sizeBytes, uint64(len(data)))
	dataWithSize := append(sizeBytes, data...)
	log.Printf("Added data size prefix: original=%d bytes, with prefix=%d bytes",
		len(data), len(dataWithSize))

	// Encrypt data
	encrypted = gcm.Seal(nil, iv, dataWithSize, nil)
	log.Printf("Data encrypted - Original: %d bytes, Encrypted: %d bytes",
		len(dataWithSize), len(encrypted))

	// Generate salt
	salt = make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	log.Printf("Generated salt: %x", salt)

	return encrypted, iv, salt, shares, nil
}

func (s *EncryptionService) DecryptFile(encrypted []byte, iv []byte, keyShares []KeyShare, k int, salt []byte) ([]byte, error) {
	log.Printf("\nStarting file decryption:")
	log.Printf("Input parameters:")
	log.Printf("- Encrypted data length: %d bytes", len(encrypted))
	log.Printf("- IV: %x (length=%d)", iv, len(iv))
	log.Printf("- Salt: %x (length=%d)", salt, len(salt))
	log.Printf("- Shares provided: %d, Threshold: %d", len(keyShares), k)

	// Log all shares
	for i, share := range keyShares {
		log.Printf("Share %d: Index=%d, Value=%s (length=%d)",
			i, share.Index, share.Value, len(share.Value))
		if rawBytes, err := hex.DecodeString(share.Value); err == nil {
			log.Printf("  Decoded bytes: %x (length=%d)", rawBytes, len(rawBytes))
		}
	}

	// Reconstruct key from shares
	log.Printf("\nReconstructing key from shares...")
	key, err := s.shamirService.RecombineKey(keyShares, k)
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct key: %w", err)
	}
	log.Printf("Key reconstruction successful:")
	log.Printf("- Raw key: %x", key)
	log.Printf("- Length: %d bytes", len(key))

	// Ensure key is exactly 32 bytes
	if len(key) != 32 {
		paddedKey := make([]byte, 32)
		copy(paddedKey, key)
		key = paddedKey
		log.Printf("Key padded to 32 bytes: %x", key)
	}

	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}
	log.Printf("Created AES cipher block. Block size: %d bytes", block.BlockSize())

	// Create GCM
	gcm, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}
	log.Printf("Created GCM with 16-byte nonce size")

	// Decrypt data
	log.Printf("\nAttempting decryption...")
	decrypted, err := gcm.Open(nil, iv, encrypted, nil)
	if err != nil {
		log.Printf("Decryption failed:")
		log.Printf("- Error: %v", err)
		log.Printf("- Key used: %x", key)
		log.Printf("- IV used: %x", iv)
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	// Extract original size and data
	if len(decrypted) < 8 {
		return nil, fmt.Errorf("decrypted data too short")
	}
	originalSize := binary.LittleEndian.Uint64(decrypted[:8])
	if uint64(len(decrypted)-8) < originalSize {
		return nil, fmt.Errorf("decrypted data shorter than original size")
	}

	data := decrypted[8 : 8+originalSize]
	log.Printf("Decryption successful:")
	log.Printf("- Original size: %d bytes", originalSize)
	log.Printf("- Decrypted data size: %d bytes", len(data))

	return data, nil
}

func (s *EncryptionService) validateKeyShares(shares []KeyShare, k int) error {
	if len(shares) < k {
		return fmt.Errorf("insufficient shares generated: got %d, need %d", len(shares), k)
	}

	seenIndices := make(map[int]bool)
	for i, share := range shares {
		if share.Index <= 0 || share.Index > 255 {
			return fmt.Errorf("invalid share index at position %d: %d", i, share.Index)
		}
		if seenIndices[share.Index] {
			return fmt.Errorf("duplicate share index: %d", share.Index)
		}
		seenIndices[share.Index] = true

		if share.Value == "" {
			return fmt.Errorf("empty share value at position %d", i)
		}
	}

	return nil
}

func (s *EncryptionService) validateEncryptionOutputs(encrypted []byte, iv []byte, shares []KeyShare, k int) error {
	if len(encrypted) == 0 {
		return fmt.Errorf("encrypted data is empty")
	}

	if len(iv) != 16 {
		return fmt.Errorf("invalid IV length: got %d, want 16", len(iv))
	}

	return s.validateKeyShares(shares, k)
}

func (s *EncryptionService) validateDecryptionInputs(encrypted []byte, iv []byte, salt []byte, shares []KeyShare, k int) error {
	if len(encrypted) == 0 {
		return fmt.Errorf("encrypted data is empty")
	}

	if len(iv) != 16 {
		return fmt.Errorf("invalid IV length: got %d, want 16", len(iv))
	}

	if len(salt) != 32 {
		return fmt.Errorf("invalid salt length: got %d, want 32", len(salt))
	}

	return s.validateKeyShares(shares, k)
}

// EncryptKeyFragment encrypts a fragment with a password
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

// DecryptKeyFragment decrypts a fragment with a password
func (s *EncryptionService) DecryptKeyFragment(encryptedFragment []byte, password []byte) ([]byte, error) {
	// Check minimum length (32 bytes salt + 16 bytes nonce + at least 1 byte data)
	if len(encryptedFragment) < 49 {
		return nil, fmt.Errorf("encrypted fragment too short")
	}

	// Extract salt and nonce
	salt := encryptedFragment[0:32]
	nonce := encryptedFragment[32:48]
	ciphertext := encryptedFragment[48:]

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

	// Decrypt the fragment
	decrypted, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt fragment: %w", err)
	}

	return decrypted, nil
}
