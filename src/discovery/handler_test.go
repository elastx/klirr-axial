package discovery

import (
	"errors"
	"net"
	"testing"

	"axial/models"
	"axial/remote"
)

func TestSyncOnMismatchHandler_TriggersOnMismatch(t *testing.T) {
	var captured remote.API
	var capturedHash string
	calls := 0

	h := &SyncOnMismatchHandler{
		Hashes: func() models.HashSet { return models.HashSet{Full: "ours"} },
		Busy:   func() bool { return false },
		StartSync: func(node remote.API, hash string) error {
			calls++
			captured = node
			capturedHash = hash
			return nil
		},
	}

	src := &net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 45678}
	h.Handle(Announcement{
		NodeID:  "remote-node",
		Hash:    "theirs",
		APIAddr: ":8080",
		Src:     src,
	})

	if calls != 1 {
		t.Fatalf("expected exactly one StartSync call, got %d", calls)
	}
	if captured.Address != "10.0.0.1:8080" {
		t.Errorf("Address: got %q want %q", captured.Address, "10.0.0.1:8080")
	}
	if capturedHash != "theirs" {
		t.Errorf("hash: got %q want %q", capturedHash, "theirs")
	}
}

func TestSyncOnMismatchHandler_SkipsWhenBusy(t *testing.T) {
	calls := 0
	h := &SyncOnMismatchHandler{
		Hashes:    func() models.HashSet { return models.HashSet{Full: "ours"} },
		Busy:      func() bool { return true },
		StartSync: func(remote.API, string) error { calls++; return nil },
	}
	h.Handle(Announcement{Hash: "theirs", Src: &net.UDPAddr{}})
	if calls != 0 {
		t.Errorf("StartSync should not be called when already syncing; got %d calls", calls)
	}
}

func TestSyncOnMismatchHandler_SkipsWhenHashMatches(t *testing.T) {
	calls := 0
	h := &SyncOnMismatchHandler{
		Hashes:    func() models.HashSet { return models.HashSet{Full: "same"} },
		Busy:      func() bool { return false },
		StartSync: func(remote.API, string) error { calls++; return nil },
	}
	h.Handle(Announcement{Hash: "same", Src: &net.UDPAddr{}})
	if calls != 0 {
		t.Errorf("StartSync should not be called when hashes match; got %d calls", calls)
	}
}

func TestSyncOnMismatchHandler_StartSyncErrorIsNotFatal(t *testing.T) {
	// A failed sync round must not panic — the handler will be called again
	// on the next broadcast and should remain usable.
	h := &SyncOnMismatchHandler{
		Hashes:    func() models.HashSet { return models.HashSet{Full: "ours"} },
		Busy:      func() bool { return false },
		StartSync: func(remote.API, string) error { return errors.New("boom") },
	}
	h.Handle(Announcement{Hash: "theirs", Src: &net.UDPAddr{IP: net.ParseIP("1.2.3.4")}})
}
