package services

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
)

type ShamirService struct {
	prime *big.Int
}

type KeyShare struct {
	Index int    `json:"index"`
	Value string `json:"value"`
}

func NewShamirService() *ShamirService {
	// Large prime for finite field arithmetic (256-bit)
	prime, _ := new(big.Int).SetString("115792089237316195423570985008687907853269984665640564039457584007913129639747", 10)
	return &ShamirService{prime: prime}
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

	secret := new(big.Int).SetBytes(key)

	// Generate random coefficients for the polynomial
	coefficients := make([]*big.Int, k-1)
	for i := range coefficients {
		coeff, err := rand.Int(rand.Reader, s.prime)
		if err != nil {
			return nil, fmt.Errorf("failed to generate coefficient: %w", err)
		}
		coefficients[i] = coeff
	}

	// Generate shares
	shares := make([]KeyShare, n)
	for i := 0; i < n; i++ {
		x := big.NewInt(int64(i + 1))
		y := s.evaluatePolynomial(secret, coefficients, x)

		shares[i] = KeyShare{
			Index: i + 1,
			Value: base64.StdEncoding.EncodeToString(y.Bytes()),
		}
	}

	return shares, nil
}

// RecombineKey reconstructs the original key from k shares
func (s *ShamirService) RecombineKey(shares []KeyShare, k int) ([]byte, error) {
	if len(shares) < k {
		return nil, fmt.Errorf("insufficient shares: got %d, need %d", len(shares), k)
	}

	// Convert shares from base64 to big.Int
	points := make(map[int64]*big.Int)
	for i := 0; i < k; i++ {
		value, err := base64.StdEncoding.DecodeString(shares[i].Value)
		if err != nil {
			return nil, fmt.Errorf("invalid share value at index %d: %w", i, err)
		}
		points[int64(shares[i].Index)] = new(big.Int).SetBytes(value)
	}

	// Perform Lagrange interpolation
	secret := new(big.Int).SetInt64(0)
	for i := range points {
		numerator := big.NewInt(1)
		denominator := big.NewInt(1)

		for j := range points {
			if i == j {
				continue
			}
			zero := big.NewInt(0)
			zero.Sub(zero, big.NewInt(j))
			numerator.Mul(numerator, zero)
			temp := big.NewInt(i)
			temp.Sub(temp, big.NewInt(j))
			denominator.Mul(denominator, temp)
		}

		denominator.ModInverse(denominator, s.prime)
		numerator.Mul(numerator, denominator)
		numerator.Mul(numerator, points[i])
		numerator.Mod(numerator, s.prime)

		secret.Add(secret, numerator)
		secret.Mod(secret, s.prime)
	}

	return secret.Bytes(), nil
}

// Helper: Evaluates polynomial at point x
func (s *ShamirService) evaluatePolynomial(secret *big.Int, coefficients []*big.Int, x *big.Int) *big.Int {
	result := new(big.Int).Set(secret)
	xPow := new(big.Int).Set(x)

	for _, coeff := range coefficients {
		term := new(big.Int).Mul(coeff, xPow)
		result.Add(result, term)
		result.Mod(result, s.prime)
		xPow.Mul(xPow, x)
		xPow.Mod(xPow, s.prime)
	}

	return result
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
