package services

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/smtp"
	"sync"
)

type SMTPConfig struct {
	Host      string
	Port      int
	Username  string
	Password  string
	FromName  string
	FromEmail string
}

type SMTPEmailService struct {
	config SMTPConfig
	auth   smtp.Auth
	client *smtp.Client
	mu     sync.Mutex
}

func NewSMTPEmailService(config SMTPConfig) (*SMTPEmailService, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)

	return &SMTPEmailService{
		config: config,
		auth:   auth,
	}, nil
}

func validateConfig(config SMTPConfig) error {
	if config.Host == "" {
		return errors.New("SMTP host is required")
	}
	if config.Port == 0 {
		return errors.New("SMTP port is required")
	}
	if config.Username == "" {
		return errors.New("SMTP username is required")
	}
	if config.Password == "" {
		return errors.New("SMTP password is required")
	}
	if config.FromEmail == "" {
		return errors.New("sender email is required")
	}
	return nil
}

func (s *SMTPEmailService) connect() error {
    s.mu.Lock()
    defer s.mu.Unlock()

    tlsConfig := &tls.Config{
        ServerName: s.config.Host,
        MinVersion: tls.VersionTLS12,
    }

    conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", s.config.Host, s.config.Port), tlsConfig)
    if err != nil {
        return fmt.Errorf("failed to establish TLS connection: %w", err)
    }

    client, err := smtp.NewClient(conn, s.config.Host)
    if err != nil {
        conn.Close()
        return fmt.Errorf("failed to create SMTP client: %w", err)
    }

    // Authenticate
    if err := client.Auth(s.auth); err != nil {
        client.Close()
        return fmt.Errorf("authentication failed: %w", err)
    }

    s.client = client
    return nil
}
func (s *SMTPEmailService) SendEmail(to, subject, body string) error {
	if err := s.connect(); err != nil {
		return err
	}

	msg := []byte(fmt.Sprintf("From: %s <%s>\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/plain; charset=utf-8\r\n"+
		"\r\n"+
		"%s", s.config.FromName, s.config.FromEmail, to, subject, body))

	if err := s.client.Mail(s.config.FromEmail); err != nil {
		s.reconnect()
		return fmt.Errorf("MAIL FROM command failed: %w", err)
	}

	if err := s.client.Rcpt(to); err != nil {
		s.reconnect()
		return fmt.Errorf("RCPT TO command failed: %w", err)
	}

	w, err := s.client.Data()
	if err != nil {
		s.reconnect()
		return fmt.Errorf("DATA command failed: %w", err)
	}

	if _, err := w.Write(msg); err != nil {
		s.reconnect()
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err := w.Close(); err != nil {
		s.reconnect()
		return fmt.Errorf("failed to close message writer: %w", err)
	}

	return nil
}

func (s *SMTPEmailService) reconnect() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client != nil {
		s.client.Close()
		s.client = nil
	}
}

func (s *SMTPEmailService) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client != nil {
		if err := s.client.Quit(); err != nil {
			return fmt.Errorf("failed to close SMTP connection: %w", err)
		}
		s.client = nil
	}
	return nil
}