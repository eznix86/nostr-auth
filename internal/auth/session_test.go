package auth

import (
	"testing"
	"time"
)

func TestSessionSignerRoundTrip(t *testing.T) {
	signer := NewSessionSigner("secret", time.Hour)

	token, err := signer.Sign("pubkey123")
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	claims, err := signer.Verify(token)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if claims.PubKey != "pubkey123" {
		t.Fatalf("Verify() pubkey = %s, want %s", claims.PubKey, "pubkey123")
	}
}

func TestSessionSignerRejectsTamperedToken(t *testing.T) {
	signer := NewSessionSigner("secret", time.Hour)

	token, err := signer.Sign("pubkey123")
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	if _, err := signer.Verify(token + "x"); err != ErrInvalidSession {
		t.Fatalf("Verify() error = %v, want %v", err, ErrInvalidSession)
	}
}

func TestSessionSignerRejectsExpiredToken(t *testing.T) {
	signer := NewSessionSigner("secret", -time.Second)

	token, err := signer.Sign("pubkey123")
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	if _, err := signer.Verify(token); err != ErrInvalidSession {
		t.Fatalf("Verify() error = %v, want %v", err, ErrInvalidSession)
	}
}

func TestSessionSignerRejectsEmptyPubKey(t *testing.T) {
	signer := NewSessionSigner("secret", time.Hour)

	if _, err := signer.Sign(""); err != nil {
		t.Fatalf("Sign() with empty pubkey error = %v", err)
	}
}

func TestSessionSignerRejectsInvalidTokenFormat(t *testing.T) {
	signer := NewSessionSigner("secret", time.Hour)

	// Missing signature part
	if _, err := signer.Verify("payload"); err != ErrInvalidSession {
		t.Fatalf("Verify() error = %v, want %v", err, ErrInvalidSession)
	}

	// Too many parts
	if _, err := signer.Verify("payload.signature.extra"); err != ErrInvalidSession {
		t.Fatalf("Verify() error = %v, want %v", err, ErrInvalidSession)
	}
}

func TestSessionSignerRejectsWrongSecret(t *testing.T) {
	signer1 := NewSessionSigner("secret1", time.Hour)
	signer2 := NewSessionSigner("secret2", time.Hour)

	token, err := signer1.Sign("pubkey123")
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	// Verify with different secret should fail
	if _, err := signer2.Verify(token); err != ErrInvalidSession {
		t.Fatalf("Verify() error = %v, want %v", err, ErrInvalidSession)
	}
}
