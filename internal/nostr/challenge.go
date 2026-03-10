package nostr

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

const (
	ChallengeCookieName = "nostr_challenge_id"
)

var (
	ErrChallengeNotFound = errors.New("challenge not found")
	ErrChallengeExpired  = errors.New("challenge expired")
	ErrChallengeMismatch = errors.New("challenge mismatch")
)

type Challenge struct {
	Value     string
	ExpiresAt time.Time
}

type ChallengeStore struct {
	mu      sync.Mutex
	ttl     time.Duration
	entries map[string]Challenge
}

func NewChallengeStore(ttl time.Duration) *ChallengeStore {
	return &ChallengeStore{
		ttl:     ttl,
		entries: make(map[string]Challenge),
	}
}

func (s *ChallengeStore) Issue(sessionID string) (Challenge, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	s.cleanupExpired(now)

	value, err := randomHex(16)
	if err != nil {
		return Challenge{}, err
	}

	challenge := Challenge{
		Value:     value,
		ExpiresAt: now.Add(s.ttl),
	}

	s.entries[sessionID] = challenge

	return challenge, nil
}

func (s *ChallengeStore) Current(sessionID string) (Challenge, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.current(sessionID, time.Now())
}

func (s *ChallengeStore) Consume(sessionID, challenge string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	stored, err := s.current(sessionID, time.Now())
	if err != nil {
		return err
	}

	if stored.Value != challenge {
		return ErrChallengeMismatch
	}

	delete(s.entries, sessionID)

	return nil
}

func (s *ChallengeStore) current(sessionID string, now time.Time) (Challenge, error) {
	challenge, ok := s.entries[sessionID]
	if !ok {
		return Challenge{}, ErrChallengeNotFound
	}

	if now.After(challenge.ExpiresAt) {
		delete(s.entries, sessionID)
		return Challenge{}, ErrChallengeExpired
	}

	return challenge, nil
}

func NewSessionID() (string, error) {
	return randomHex(16)
}

func randomHex(size int) (string, error) {
	buf := make([]byte, size)

	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}

func (s *ChallengeStore) cleanupExpired(now time.Time) {
	for sessionID, challenge := range s.entries {
		if now.After(challenge.ExpiresAt) {
			delete(s.entries, sessionID)
		}
	}
}
