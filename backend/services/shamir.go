package services

import (
	"encoding/hex"
	"fmt"
	"log"
	"path/filepath"
	"github.com/hashicorp/vault/shamir"
)

// KeyShare represents a Shamir's share with metadata
type KeyShare struct {
	Index            int     `json:"index"`            // The share's x‑coordinate (1..255)
	Value            string  `json:"value"`            // Hex-encoded 32‑byte secret portion
	HolderType       string  `json:"holder_type"`      // "server" or "user"
	NodeIndex        int     `json:"node_index"`       // Which node stores this share
	FragmentPath     string  `json:"fragment_path"`    // Path in node storage
	EncryptionNonce  []byte  `json:"encryption_nonce"` // For encrypting the share
	MasterKeyVersion *int    `json:"master_key_version,omitempty"`
	ServerKeyID      *string `json:"server_key_id,omitempty"`
}

type ShamirService struct {
	nodeCount int
}

func NewShamirService(nodeCount int) *ShamirService {
	return &ShamirService{
		nodeCount: nodeCount,
	}
}

// SplitKey splits a 32‑byte key into n shares with threshold k
func (s *ShamirService) SplitKey(key []byte, n, k int, fileID uint, serverKeyID string) ([]KeyShare, error) {
	if k > n {
		return nil, fmt.Errorf("threshold k cannot be greater than total shares n")
	}
	if k < 2 {
		return nil, fmt.Errorf("threshold must be at least 2")
	}
	if n > 255 {
		return nil, fmt.Errorf("maximum number of shares is 255")
	}

	log.Printf("Splitting key for file %d: n=%d, k=%d", fileID, n, k)

	rawShares, err := shamir.Split(key, n, k)
	if err != nil {
		return nil, fmt.Errorf("failed to split key: %w", err)
	}

	keyShares := make([]KeyShare, len(rawShares))
	for i, shareBytes := range rawShares {
		if len(shareBytes) < 2 {
			return nil, fmt.Errorf("unexpected share length: got %d, want >= 2", len(shareBytes))
		}

		xCoord := shareBytes[0]
		secretPart := shareBytes[1:] // should be exactly 32 bytes

		// Determine holder type and node assignment
		holderType := "user"
		var masterKeyVersion *int
		var sKeyID *string
		nodeIndex := i % s.nodeCount

		if i < k-1 { // First k-1 shares go to server
			holderType = "server"
			sKeyID = &serverKeyID
		} else {
			version := 1
			masterKeyVersion = &version
		}

		fragmentPath := filepath.Join(
            fmt.Sprintf("file_%d", fileID),
            fmt.Sprintf("fragment_%d", xCoord),
        )

		keyShares[i] = KeyShare{
			Index:            int(xCoord),
			Value:            hex.EncodeToString(secretPart),
			HolderType:       holderType,
			NodeIndex:        nodeIndex,
			FragmentPath:     fragmentPath,
			MasterKeyVersion: masterKeyVersion,
			ServerKeyID:      sKeyID,
		}

		log.Printf("Created share %d: Index=%d, Type=%s, Node=%d, Length=%d",
			i, xCoord, holderType, nodeIndex, len(secretPart))
	}

	// Test reconstruction before returning
	testShares := keyShares[:k]
	testResult, err := s.testReconstruction(testShares, key)
	if err != nil {
		return nil, fmt.Errorf("share validation failed: %w", err)
	}
	log.Printf("Share validation successful: %v", testResult)

	return keyShares, nil
}

func (s *ShamirService) RecombineKey(shares []KeyShare, k int) ([]byte, error) {
	log.Printf("\nStarting key reconstruction from %d shares (threshold: %d)", len(shares), k)
	if len(shares) < k {
		log.Printf("Error: Insufficient shares provided: got %d, need %d", len(shares), k)
		return nil, fmt.Errorf("insufficient shares: got %d, need %d", len(shares), k)
	}

	// Build raw shares
	rawShares := make([][]byte, 0, len(shares))
	for i, share := range shares {
		log.Printf("\nProcessing share %d:", i)
		log.Printf("- Index: %d", share.Index)
		log.Printf("- Node: %d", share.NodeIndex)
		log.Printf("- Path: %s", share.FragmentPath)

		// Try to decode the hex value first
		decodedValue, err := hex.DecodeString(share.Value)
		if err != nil {
			log.Printf("Warning: Failed to decode hex value for share %d: %v", i, err)
			continue
		}
		log.Printf("- Decoded hex value length: %d bytes", len(decodedValue))

		if len(decodedValue) != 32 {
			log.Printf("Warning: Invalid decoded share length: got %d, want 32", len(decodedValue))
			continue
		}

		// Prepend x-coordinate to create Shamir share format
		share33 := append([]byte{byte(share.Index)}, decodedValue...)
		rawShares = append(rawShares, share33)

		log.Printf("- Successfully added share to reconstruction set (33 bytes with x-coordinate)")
	}

	log.Printf("\nCollected %d valid shares for reconstruction", len(rawShares))
	if len(rawShares) < k {
		log.Printf("Error: Insufficient valid shares after processing: got %d, need %d", len(rawShares), k)
		return nil, fmt.Errorf("insufficient valid shares after processing: got %d, need %d",
			len(rawShares), k)
	}

	// Log shares before reconstruction
	for i, raw := range rawShares {
		log.Printf("Share %d: x-coordinate=%d, data length=%d bytes", i, raw[0], len(raw)-1)
	}

	log.Printf("\nAttempting to combine shares using Shamir algorithm...")
	combinedKey, err := shamir.Combine(rawShares)
	if err != nil {
		log.Printf("Error during share combination: %v", err)
		return nil, fmt.Errorf("failed to combine shares: %w", err)
	}

	log.Printf("Share combination successful")
	log.Printf("Combined key length: %d bytes", len(combinedKey))
	log.Printf("Combined key: %x", combinedKey)

	if len(combinedKey) != 32 {
		log.Printf("Error: Unexpected recombined key size: got %d, want 32", len(combinedKey))
		return nil, fmt.Errorf("unexpected recombined key size: got %d, want 32", len(combinedKey))
	}

	return combinedKey, nil
}

// testReconstruction verifies that shares can reconstruct the original key
func (s *ShamirService) testReconstruction(shares []KeyShare, originalKey []byte) (bool, error) {
	log.Printf("\nTesting reconstruction with %d shares...", len(shares))

	reconstructed, err := s.RecombineKey(shares, len(shares))
	if err != nil {
		log.Printf("Test reconstruction failed: %v", err)
		return false, err
	}

	log.Printf("Test reconstruction result:")
	log.Printf("  Original key    : %x", originalKey)
	log.Printf("  Reconstructed key: %x", reconstructed)

	// Compare reconstructed key with original
	if len(reconstructed) != len(originalKey) {
		log.Printf("Length mismatch: got %d, want %d", len(reconstructed), len(originalKey))
		return false, nil
	}

	matches := true
	for i := range reconstructed {
		if reconstructed[i] != originalKey[i] {
			matches = false
			log.Printf("Mismatch at byte %d: got %02x, want %02x",
				i, reconstructed[i], originalKey[i])
		}
	}

	log.Printf("Keys match: %v", matches)
	return matches, nil
}

// ValidateShares checks if shares are valid
func (s *ShamirService) ValidateShares(shares []KeyShare) error {
	log.Printf("\nValidating %d shares...", len(shares))
	if len(shares) == 0 {
		return fmt.Errorf("no shares provided")
	}

	seenIndices := make(map[int]bool)
	for i, share := range shares {
		log.Printf("\nValidating share %d:", i)
		log.Printf("- Index: %d", share.Index)
		log.Printf("- Node: %d", share.NodeIndex)
		log.Printf("- Path: %s", share.FragmentPath)

		// Validate index
		if share.Index < 1 || share.Index > 255 {
			log.Printf("Error: Invalid share index: %d", share.Index)
			return fmt.Errorf("invalid share index: %d", share.Index)
		}
		if seenIndices[share.Index] {
			log.Printf("Error: Duplicate share index: %d", share.Index)
			return fmt.Errorf("duplicate share index: %d", share.Index)
		}
		seenIndices[share.Index] = true

		// Validate node assignment
		if share.NodeIndex >= s.nodeCount {
			log.Printf("Error: Invalid node index: %d", share.NodeIndex)
			return fmt.Errorf("invalid node index: %d", share.NodeIndex)
		}

		// Try to decode the hex value
		if _, err := hex.DecodeString(share.Value); err != nil {
			log.Printf("Error: Invalid hex value: %v", err)
			return fmt.Errorf("invalid hex value in share %d: %v", i, err)
		}
	}

	log.Printf("\nAll shares validated successfully")
	return nil
}
