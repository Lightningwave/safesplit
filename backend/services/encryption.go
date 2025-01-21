package services

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/twofish"
)

type EncryptionType string

const (
	StandardEncryption EncryptionType = "standard" // AES-256-GCM
	ChaCha20           EncryptionType = "chacha20" // ChaCha20-Poly1305
	Twofish            EncryptionType = "twofish"  // Twofish
)

type EncryptionService struct {
	shamirService *ShamirService
}

func NewEncryptionService(shamirService *ShamirService) *EncryptionService {
	return &EncryptionService{
		shamirService: shamirService,
	}
}

func (s *EncryptionService) EncryptFileWithType(
	data []byte,
	n, k int,
	fileID uint,
	serverKeyID string,
	encType EncryptionType,
) (encrypted []byte, iv []byte, salt []byte, shares []KeyShare, err error) {
	log.Printf("Starting file encryption with type=%s, n=%d, k=%d, fileID=%d", encType, n, k, fileID)

	// Generate encryption key (32 bytes for all types)
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to generate key: %w", err)
	}
	log.Printf("Generated encryption key: %x (length=%d)", key, len(key))

	// Split key into shares and store them
	shares, err = s.shamirService.SplitKey(key, n, k, fileID, serverKeyID)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to split and store key: %w", err)
	}
	log.Printf("Split key into %d shares (threshold: %d)", len(shares), k)

	// Generate IV/nonce with appropriate size
	var nonceSize int
	switch encType {
	case ChaCha20:
		nonceSize = chacha20poly1305.NonceSizeX // 24 bytes
	case Twofish:
		nonceSize = 12 // GCM requires 12 bytes for Twofish
	case StandardEncryption:
		nonceSize = 16 // AES-GCM can use 16 bytes
	}

	iv = make([]byte, nonceSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to generate IV: %w", err)
	}
	log.Printf("Generated IV: %x (length=%d)", iv, len(iv))

	// Prepend the original data size to the data
	sizeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(sizeBytes, uint64(len(data)))
	dataWithSize := append(sizeBytes, data...)

	// Encrypt based on type
	switch encType {
	case StandardEncryption:
		encrypted, err = s.encryptAES(dataWithSize, key, iv)
	case ChaCha20:
		encrypted, err = s.encryptChaCha20(dataWithSize, key, iv)
	case Twofish:
		encrypted, err = s.encryptTwofish(dataWithSize, key, iv)
	}

	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("encryption failed: %w", err)
	}

	log.Printf("Data encrypted - Original: %d bytes, Encrypted: %d bytes",
		len(dataWithSize), len(encrypted))

	// Generate salt
	salt = make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	log.Printf("Generated salt: %x", salt)

	// Test reconstruction
	if err := s.testReconstruction(shares[:k], k, key); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("key reconstruction test failed: %w", err)
	}

	return encrypted, iv, salt, shares, nil
}

func (s *EncryptionService) encryptAES(data, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Use consistent nonce size of 16 bytes
	gcm, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return gcm.Seal(nil, iv, data, nil), nil
}

func (s *EncryptionService) encryptChaCha20(data, key, nonce []byte) ([]byte, error) {
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create ChaCha20-Poly1305: %w", err)
	}

	return aead.Seal(nil, nonce, data, nil), nil
}

func (s *EncryptionService) encryptTwofish(data, key, iv []byte) ([]byte, error) {
	block, err := twofish.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create Twofish cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block) // GCM will use default 12-byte nonce
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	if len(iv) != gcm.NonceSize() {
		return nil, fmt.Errorf("invalid nonce size for Twofish-GCM: got %d, want %d",
			len(iv), gcm.NonceSize())
	}

	return gcm.Seal(nil, iv, data, nil), nil
}

// DecryptFileWithType handles decryption for all encryption types
func (s *EncryptionService) DecryptFileWithType(
	encrypted []byte,
	iv []byte,
	keyShares []KeyShare,
	k int,
	salt []byte,
	encType EncryptionType,
) ([]byte, error) {
	log.Printf("\nStarting file decryption with type %s:", encType)
	log.Printf("Input parameters:")
	log.Printf("- Encrypted data length: %d bytes", len(encrypted))
	log.Printf("- IV: %x (length=%d)", iv, len(iv))
	log.Printf("- Salt: %x (length=%d)", salt, len(salt))
	log.Printf("- Shares provided: %d, Threshold: %d", len(keyShares), k)

	// Reconstruct key from shares
	key, err := s.shamirService.RecombineKey(keyShares, k)
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct key: %w", err)
	}
	log.Printf("Key reconstruction successful")

	// Decrypt based on type
	var decrypted []byte
	switch encType {
	case StandardEncryption:
		decrypted, err = s.decryptAES(encrypted, key, iv)
	case ChaCha20:
		decrypted, err = s.decryptChaCha20(encrypted, key, iv)
	case Twofish:
		decrypted, err = s.decryptTwofish(encrypted, key, iv)
	default:
		return nil, fmt.Errorf("unsupported encryption type: %s", encType)
	}

	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
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

// Helper functions for different decryption types
func (s *EncryptionService) decryptAES(encrypted, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Use consistent nonce size of 16 bytes
	gcm, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return gcm.Open(nil, iv, encrypted, nil)
}

func (s *EncryptionService) decryptChaCha20(encrypted, key, nonce []byte) ([]byte, error) {
	aead, err := chacha20poly1305.NewX(key) // Now consistent with encryption method
	if err != nil {
		return nil, fmt.Errorf("failed to create ChaCha20-Poly1305: %w", err)
	}

	return aead.Open(nil, nonce, encrypted, nil)
}

func (s *EncryptionService) decryptTwofish(encrypted, key, iv []byte) ([]byte, error) {
	block, err := twofish.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create Twofish cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	if len(iv) != gcm.NonceSize() {
		return nil, fmt.Errorf("invalid nonce size for Twofish-GCM: got %d, want %d",
			len(iv), gcm.NonceSize())
	}

	return gcm.Open(nil, iv, encrypted, nil)
}

// Maintain backward compatibility
func (s *EncryptionService) EncryptFile(data []byte, n, k int, fileID uint, serverKeyID string) (encrypted []byte, iv []byte, salt []byte, shares []KeyShare, err error) {
	return s.EncryptFileWithType(data, n, k, fileID, serverKeyID, StandardEncryption)
}

func (s *EncryptionService) DecryptFile(encrypted []byte, iv []byte, keyShares []KeyShare, k int, salt []byte) ([]byte, error) {
	return s.DecryptFileWithType(encrypted, iv, keyShares, k, salt, StandardEncryption)
}

// Test key reconstruction
func (s *EncryptionService) testReconstruction(shares []KeyShare, k int, originalKey []byte) error {
	log.Printf("Testing reconstruction with %d shares...", k)
	reconstructedKey, err := s.shamirService.RecombineKey(shares, k)
	if err != nil {
		return fmt.Errorf("reconstruction failed: %w", err)
	}

	if !bytes.Equal(originalKey, reconstructedKey) {
		return fmt.Errorf("reconstructed key does not match original")
	}

	log.Printf("Key reconstruction test passed")
	return nil
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

func (s *EncryptionService) validateEncryptionOutputs(encrypted []byte, iv []byte, shares []KeyShare, k int, encType EncryptionType) error {
	if len(encrypted) == 0 {
		return fmt.Errorf("encrypted data is empty")
	}

	// Get expected IV size based on encryption type
	var expectedIVSize int
	switch encType {
	case ChaCha20:
		expectedIVSize = chacha20poly1305.NonceSizeX // 24 bytes
	case Twofish:
		expectedIVSize = 12 // GCM requires 12 bytes for Twofish
	case StandardEncryption:
		expectedIVSize = 16 // AES-GCM
	default:
		return fmt.Errorf("unsupported encryption type: %s", encType)
	}

	if len(iv) != expectedIVSize {
		return fmt.Errorf("invalid IV length for %s: got %d, want %d", encType, len(iv), expectedIVSize)
	}

	return s.validateKeyShares(shares, k)
}

func (s *EncryptionService) validateDecryptionInputs(encrypted []byte, iv []byte, salt []byte, shares []KeyShare, k int, encType EncryptionType) error {
	if len(encrypted) == 0 {
		return fmt.Errorf("encrypted data is empty")
	}

	// Get expected IV size based on encryption type
	var expectedIVSize int
	switch encType {
	case ChaCha20:
		expectedIVSize = chacha20poly1305.NonceSizeX // 24 bytes
	case Twofish:
		expectedIVSize = 12 // GCM requires 12 bytes for Twofish
	case StandardEncryption:
		expectedIVSize = 16 // AES-GCM
	default:
		return fmt.Errorf("unsupported encryption type: %s", encType)
	}

	if len(iv) != expectedIVSize {
		return fmt.Errorf("invalid IV length for %s: got %d, want %d", encType, len(iv), expectedIVSize)
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
