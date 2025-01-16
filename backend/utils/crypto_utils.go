package utils

import (
    "crypto/rand"
    "fmt"
)

const (
    SaltSize  = 32 // 256 bits for salt
    NonceSize = 12 // 96 bits for GCM nonce
)

// GenerateSalt creates a new random salt
func GenerateSalt() ([]byte, error) {
    salt := make([]byte, SaltSize)
    if _, err := rand.Read(salt); err != nil {
        return nil, fmt.Errorf("failed to generate salt: %w", err)
    }
    return salt, nil
}

// GenerateNonce creates a new random nonce for AES-GCM
func GenerateNonce() ([]byte, error) {
    nonce := make([]byte, NonceSize)
    if _, err := rand.Read(nonce); err != nil {
        return nil, fmt.Errorf("failed to generate nonce: %w", err)
    }
    return nonce, nil
}