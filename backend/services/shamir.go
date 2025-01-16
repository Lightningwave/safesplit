package services

import (
    "encoding/hex"
    "fmt"

    "github.com/hashicorp/vault/shamir"
)

// KeyShare holds the separate index (int) plus the secret portion (32 bytes in hex).
type KeyShare struct {
    Index int    `json:"index"` // The share’s x‑coordinate (1..255)
    Value string `json:"value"` // Hex-encoded 32‑byte secret portion
}

type ShamirService struct{}

func NewShamirService() *ShamirService {
    return &ShamirService{}
}

// SplitKey splits a 32‑byte key into n shares with threshold k, but stores only
// the 32‑byte secret portion in Value. The 1‑byte index from Shamir is moved
// into KeyShare.Index.
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

    // HashiCorp's Shamir returns an array of 33‑byte shares:
    //   [ 1 byte x‑coordinate | 32 bytes secret ]
    rawShares, err := shamir.Split(key, n, k)
    if err != nil {
        return nil, fmt.Errorf("failed to split key: %w", err)
    }

    keyShares := make([]KeyShare, len(rawShares))
    for i, shareBytes := range rawShares {
        if len(shareBytes) < 2 {
            return nil, fmt.Errorf("unexpected share length: got %d, want >= 2", len(shareBytes))
        }

        // First byte is the x‑coordinate (1..255), next 32 bytes is the secret
        xCoord := shareBytes[0]
        secretPart := shareBytes[1:] // should be exactly 32 bytes

        keyShares[i] = KeyShare{
            Index: int(xCoord),
            Value: hex.EncodeToString(secretPart), // store 32 bytes in hex
        }
    }

    return keyShares, nil
}

// RecombineKey reconstructs the original 32‑byte key using the stored Index
// and 32‑byte Value. We prepend the 1‑byte index to the Value before calling
// shamir.Combine.
func (s *ShamirService) RecombineKey(shares []KeyShare, k int) ([]byte, error) {
    if len(shares) < k {
        return nil, fmt.Errorf("insufficient shares: got %d, need %d", len(shares), k)
    }

    // Build the raw 33‑byte slices by prepending the index byte
    rawShares := make([][]byte, 0, len(shares))
    for _, share := range shares {
        secretBytes, err := hex.DecodeString(share.Value)
        if err != nil {
            return nil, fmt.Errorf("invalid share value at index %d: %w", share.Index, err)
        }

        if len(secretBytes) != 32 {
            return nil, fmt.Errorf("share at index %d has invalid length: got %d, want 32",
                share.Index, len(secretBytes))
        }

        // Prepend the x‑coordinate as the first byte
        xByte := byte(share.Index)
        share33 := append([]byte{xByte}, secretBytes...)
        rawShares = append(rawShares, share33)
    }

    // Now combine them
    combinedKey, err := shamir.Combine(rawShares)
    if err != nil {
        return nil, fmt.Errorf("failed to combine shares: %w", err)
    }

    // Should now be the original 32‑byte key
    if len(combinedKey) != 32 {
        return nil, fmt.Errorf("unexpected recombined key size: got %d, want 32", len(combinedKey))
    }

    return combinedKey, nil
}

// ValidateShares checks if shares are syntactically valid (no duplicates, index in range, etc.)
// Note: This does not fully verify that they combine successfully, just does basic checks.
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

        // Check that Value is valid hex
        secretBytes, err := hex.DecodeString(share.Value)
        if err != nil {
            return fmt.Errorf("invalid share value at index %d: %w", share.Index, err)
        }
        if len(secretBytes) != 32 {
            return fmt.Errorf("invalid share length at index %d: got %d, want 32",
                share.Index, len(secretBytes))
        }
    }

    return nil
}
