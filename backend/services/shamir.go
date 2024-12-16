package services

import (
	"encoding/base64"
	"fmt"

	"github.com/hashicorp/vault/shamir"
)

type ShamirService struct{}

type KeyShare struct {
	Index int    `json:"index"`
	Value string `json:"value"`
}

func NewShamirService() *ShamirService {
	return &ShamirService{}
}

// SplitKey splits a key into n shares with threshold k
func (s *ShamirService) SplitKey(key []byte, n, k int) ([]KeyShare, error) {
	if k > n {
		return nil, fmt.Errorf("threshold k cannot be greater than total shares n")
	}
	if k < 2 {
		return nil, fmt.Errorf("threshold must be at least 2")
	}
	if n > 255 {
		return nil, fmt.Errorf("maximum number of shares is 255")
	}

	// Use Hashicorp's Shamir implementation to split the secret
	shares, err := shamir.Split(key, n, k)
	if err != nil {
		return nil, fmt.Errorf("failed to split key: %w", err)
	}

	// Convert to our KeyShare format
	keyShares := make([]KeyShare, n)
	for i := 0; i < n; i++ {
		keyShares[i] = KeyShare{
			Index: i + 1,
			Value: base64.StdEncoding.EncodeToString(shares[i]),
		}
	}

	return keyShares, nil
}

// RecombineKey reconstructs the original key from k shares
func (s *ShamirService) RecombineKey(shares []KeyShare, k int) ([]byte, error) {
	if len(shares) < k {
		return nil, fmt.Errorf("insufficient shares: got %d, need %d", len(shares), k)
	}

	// Convert our KeyShare format to bytes for Hashicorp's implementation
	rawShares := make([][]byte, len(shares))
	for i, share := range shares {
		value, err := base64.StdEncoding.DecodeString(share.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid share value at index %d: %w", share.Index, err)
		}
		rawShares[i] = value
	}

	// Use Hashicorp's Shamir implementation to combine the shares
	secret, err := shamir.Combine(rawShares)
	if err != nil {
		return nil, fmt.Errorf("failed to combine shares: %w", err)
	}

	return secret, nil
}

// ValidateShares checks if shares are valid for reconstruction
func (s *ShamirService) ValidateShares(shares []KeyShare) error {
	if len(shares) == 0 {
		return fmt.Errorf("no shares provided")
	}

	seenIndices := make(map[int]bool)
	for _, share := range shares {
		if share.Index < 1 || share.Index > 255 {
			return fmt.Errorf("invalid share index: %d", share.Index)
		}
		if seenIndices[share.Index] {
			return fmt.Errorf("duplicate share index: %d", share.Index)
		}
		seenIndices[share.Index] = true

		// Validate share value
		_, err := base64.StdEncoding.DecodeString(share.Value)
		if err != nil {
			return fmt.Errorf("invalid share value at index %d: %w", share.Index, err)
		}
	}

	return nil
}
