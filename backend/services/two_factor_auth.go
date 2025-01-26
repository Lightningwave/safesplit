package services

import (
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	tokenLength       = 6
	tokenExpiry       = 10 * time.Minute
	maxAttempts       = 3
	requestsPerMinute = 5
)

type TwoFactorEmailSender interface {
	SendEmail(to, subject, body string) error
}

type RateLimiter struct {
	requests map[uint][]time.Time
	mu       sync.Mutex
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		requests: make(map[uint][]time.Time),
	}
}

func (r *RateLimiter) Allow(userID uint) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	window := now.Add(-time.Minute)

	// Clear old requests
	if times, exists := r.requests[userID]; exists {
		valid := make([]time.Time, 0)
		for _, t := range times {
			if t.After(window) {
				valid = append(valid, t)
			}
		}
		r.requests[userID] = valid
	}

	if len(r.requests[userID]) >= requestsPerMinute {
		return false
	}

	r.requests[userID] = append(r.requests[userID], now)
	return true
}

type TwoFactorToken struct {
	Token     string
	ExpiresAt time.Time
}

type TwoFactorAuthService struct {
	emailSender TwoFactorEmailSender
	tokens      map[uint]*TwoFactorToken
	attempts    map[uint]int
	rateLimiter *RateLimiter
	mu          sync.RWMutex
}

func NewTwoFactorAuthService(emailSender TwoFactorEmailSender) *TwoFactorAuthService {
	return &TwoFactorAuthService{
		emailSender: emailSender,
		tokens:      make(map[uint]*TwoFactorToken),
		attempts:    make(map[uint]int),
		rateLimiter: NewRateLimiter(),
	}
}

func generateToken() (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes)[:tokenLength], nil
}

func (s *TwoFactorAuthService) SendTwoFactorToken(userID uint, email string) error {
	if !s.rateLimiter.Allow(userID) {
		return errors.New("rate limit exceeded")
	}

	token, err := generateToken()
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	s.mu.Lock()
	s.tokens[userID] = &TwoFactorToken{
		Token:     token,
		ExpiresAt: time.Now().Add(tokenExpiry),
	}
	s.attempts[userID] = 0
	s.mu.Unlock()

	subject := "Two-Factor Authentication Code"
	body := fmt.Sprintf("Your authentication code is: %s\nThis code will expire in 10 minutes.", token)

	return s.emailSender.SendEmail(email, subject, body)
}

func (s *TwoFactorAuthService) VerifyToken(userID uint, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokenInfo, exists := s.tokens[userID]
	if !exists || time.Now().After(tokenInfo.ExpiresAt) {
		return errors.New("invalid or expired token")
	}

	s.attempts[userID]++
	if s.attempts[userID] > maxAttempts {
		delete(s.tokens, userID)
		delete(s.attempts, userID)
		return errors.New("max verification attempts exceeded")
	}

	if token != tokenInfo.Token {
		return errors.New("invalid token")
	}

	delete(s.tokens, userID)
	delete(s.attempts, userID)
	return nil
}

func (s *TwoFactorAuthService) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for userID, tokenInfo := range s.tokens {
		if now.After(tokenInfo.ExpiresAt) {
			delete(s.tokens, userID)
			delete(s.attempts, userID)
		}
	}
}
