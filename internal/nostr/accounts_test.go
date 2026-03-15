package nostr

import (
	"context"
	"testing"
	"time"
)

func TestClientFetchProfileUsesCache(t *testing.T) {
	client := NewClient([]string{"wss://nos.lol"}, time.Second, time.Hour)
	called := 0
	client.fetchProfile = func(_ context.Context, pubkey string, relays []string, timeout time.Duration) (*Profile, error) {
		called++
		return &Profile{PubKey: pubkey, DisplayName: "bruno"}, nil
	}

	first, err := client.FetchProfile(context.Background(), "pubkey-1")
	if err != nil {
		t.Fatalf("FetchProfile() error = %v", err)
	}
	second, err := client.FetchProfile(context.Background(), "pubkey-1")
	if err != nil {
		t.Fatalf("FetchProfile() error = %v", err)
	}

	if called != 1 {
		t.Fatalf("fetchProfile call count = %d, want 1", called)
	}
	if first != second {
		t.Fatal("expected cached profile pointer to be reused")
	}
}

func TestClientFetchProfileRefetchesAfterCacheExpiry(t *testing.T) {
	client := NewClient([]string{"wss://nos.lol"}, time.Second, 10*time.Millisecond)
	called := 0
	client.fetchProfile = func(_ context.Context, pubkey string, relays []string, timeout time.Duration) (*Profile, error) {
		called++
		return &Profile{PubKey: pubkey, DisplayName: time.Now().Format(time.RFC3339Nano)}, nil
	}

	first, err := client.FetchProfile(context.Background(), "pubkey-1")
	if err != nil {
		t.Fatalf("FetchProfile() error = %v", err)
	}

	time.Sleep(20 * time.Millisecond)

	second, err := client.FetchProfile(context.Background(), "pubkey-1")
	if err != nil {
		t.Fatalf("FetchProfile() error = %v", err)
	}

	if called != 2 {
		t.Fatalf("fetchProfile call count = %d, want 2", called)
	}
	if first == second {
		t.Fatal("expected expired cache to refetch a new profile")
	}
}
