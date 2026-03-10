package nostr

import (
	"testing"
	"time"
)

func TestChallengeStoreIssueAndConsume(t *testing.T) {
	store := NewChallengeStore(time.Minute)

	challenge, err := store.Issue("session-1")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	if challenge.Value == "" {
		t.Fatal("Issue() returned empty challenge")
	}

	if err := store.Consume("session-1", challenge.Value); err != nil {
		t.Fatalf("Consume() error = %v", err)
	}

	if _, err := store.Current("session-1"); err != ErrChallengeNotFound {
		t.Fatalf("Current() error = %v, want %v", err, ErrChallengeNotFound)
	}
}

func TestChallengeStoreCurrent(t *testing.T) {
	store := NewChallengeStore(time.Minute)

	// Issue a challenge
	challenge, err := store.Issue("session-2")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	// Current should return the same challenge
	current, err := store.Current("session-2")
	if err != nil {
		t.Fatalf("Current() error = %v", err)
	}

	if current.Value != challenge.Value {
		t.Fatalf("Current() value = %s, want %s", current.Value, challenge.Value)
	}
}

func TestChallengeStoreConsumeWrongValue(t *testing.T) {
	store := NewChallengeStore(time.Minute)

	_, err := store.Issue("session-3")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	// Consume with wrong value should fail
	if err := store.Consume("session-3", "wrong-value"); err != ErrChallengeMismatch {
		t.Fatalf("Consume() error = %v, want %v", err, ErrChallengeMismatch)
	}
}

func TestChallengeStoreConsumeNonExistent(t *testing.T) {
	store := NewChallengeStore(time.Minute)

	// Consume non-existent session
	if err := store.Consume("non-existent", "any-value"); err != ErrChallengeNotFound {
		t.Fatalf("Consume() error = %v, want %v", err, ErrChallengeNotFound)
	}
}

func TestChallengeStoreExpiration(t *testing.T) {
	store := NewChallengeStore(time.Millisecond)

	_, err := store.Issue("session-4")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Current should return expired error
	if _, err := store.Current("session-4"); err != ErrChallengeExpired {
		t.Fatalf("Current() error = %v, want %v", err, ErrChallengeExpired)
	}

	// Consume returns challenge not found because expired challenges are deleted
	if err := store.Consume("session-4", "any-value"); err != ErrChallengeNotFound {
		t.Fatalf("Consume() error = %v, want %v", err, ErrChallengeNotFound)
	}
}

func TestNewSessionID(t *testing.T) {
	id1, err := NewSessionID()
	if err != nil {
		t.Fatalf("NewSessionID() error = %v", err)
	}

	if len(id1) != 32 { // 16 bytes = 32 hex chars
		t.Fatalf("NewSessionID() length = %d, want 32", len(id1))
	}

	// Should generate unique IDs
	id2, err := NewSessionID()
	if err != nil {
		t.Fatalf("NewSessionID() error = %v", err)
	}

	if id1 == id2 {
		t.Fatal("NewSessionID() should generate unique IDs")
	}
}
