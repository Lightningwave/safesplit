package utils

import (
	"crypto/rand"
	"fmt"
	"log"
)

const (
	SaltSize  = 32 // 256 bits for salt
	NonceSize = 16 // 128 bits for custom GCM nonce
)

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	log.Printf("Generated salt: %d bytes", len(salt))
	return salt, nil
}

func GenerateNonce() ([]byte, error) {
	nonce := make([]byte, NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	log.Printf("Generated nonce: %d bytes, Value: %x", len(nonce), nonce)
	return nonce, nil
}
