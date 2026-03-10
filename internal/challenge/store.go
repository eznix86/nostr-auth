package challenge

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

const SessionCookieName = "nostr_challenge_id"

var (
	ErrNotFound = errors.New("challenge not found")
	ErrExpired  = errors.New("challenge expired")
	ErrMismatch = errors.New("challenge mismatch")
)

type Value struct {
	Token     string
	ExpiresAt time.Time
}

type Store struct {
	mu      sync.Mutex
	ttl     time.Duration
	entries map[string]Value
}

func NewStore(ttl time.Duration) *Store {
	return &Store{
		ttl:     ttl,
		entries: make(map[string]Value),
	}
}

func (s *Store) Issue(sessionID string) (Value, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	s.cleanupExpired(now)

	token, err := randomHex(16)
	if err != nil {
		return Value{}, err
	}

	challenge := Value{
		Token:     token,
		ExpiresAt: now.Add(s.ttl),
	}

	s.entries[sessionID] = challenge

	return challenge, nil
}

func (s *Store) Current(sessionID string) (Value, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.current(sessionID, time.Now())
}

func (s *Store) Consume(sessionID, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	stored, err := s.current(sessionID, time.Now())
	if err != nil {
		return err
	}

	if stored.Token != token {
		return ErrMismatch
	}

	delete(s.entries, sessionID)
	return nil
}

func NewSessionID() (string, error) {
	return randomHex(16)
}

func (s *Store) current(sessionID string, now time.Time) (Value, error) {
	challenge, ok := s.entries[sessionID]
	if !ok {
		return Value{}, ErrNotFound
	}

	if now.After(challenge.ExpiresAt) {
		delete(s.entries, sessionID)
		return Value{}, ErrExpired
	}

	return challenge, nil
}

func (s *Store) cleanupExpired(now time.Time) {
	for sessionID, challenge := range s.entries {
		if now.After(challenge.ExpiresAt) {
			delete(s.entries, sessionID)
		}
	}
}

func randomHex(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}
