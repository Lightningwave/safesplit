package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
)

type EncryptionService struct{}

func NewEncryptionService() *EncryptionService {
	return &EncryptionService{}
}

func (s *EncryptionService) EncryptFile(data []byte) (encrypted []byte, iv []byte, salt []byte, err error) {
	// Generate encryption parameters
	key := make([]byte, 32)
	iv = make([]byte, aes.BlockSize)
	salt = make([]byte, 32)

	if _, err := rand.Read(key); err != nil {
		return nil, nil, nil, err
	}
	if _, err := rand.Read(iv); err != nil {
		return nil, nil, nil, err
	}
	if _, err := rand.Read(salt); err != nil {
		return nil, nil, nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, nil, err
	}

	encrypted = gcm.Seal(nil, iv, data, nil)
	return encrypted, iv, salt, nil
}
