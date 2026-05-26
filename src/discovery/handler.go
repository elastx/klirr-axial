package discovery

import (
	"fmt"

	"axial/models"
	"axial/remote"
	"axial/synchronization"
)

// Handler reacts to a parsed Announcement. The multicast listener calls
// Handle once per well-formed discovery packet — exactly one adapter in
// production (SyncOnMismatchHandler) and one in tests, so the interface
// earns its keep as a real seam between the socket layer and the
// decision-to-sync logic.
type Handler interface {
	Handle(Announcement)
}

// SyncOnMismatchHandler triggers a sync round when the remote Node's
// announced FullHash differs from ours. The three function-valued fields
// are injection points: production wires them to models/synchronization;
// tests substitute stubs to exercise the decision logic without a database
// or HTTP client.
type SyncOnMismatchHandler struct {
	Hashes    func() models.HashSet
	Busy      func() bool
	StartSync func(node remote.API, hash string) error
}

// NewSyncOnMismatchHandler wires the production dependencies.
func NewSyncOnMismatchHandler() *SyncOnMismatchHandler {
	return &SyncOnMismatchHandler{
		Hashes:    models.GetHashes,
		Busy:      models.IsSyncing,
		StartSync: synchronization.StartSync,
	}
}

func (h *SyncOnMismatchHandler) Handle(ann Announcement) {
	if h.Busy() {
		fmt.Printf("Ignoring ping from %s because we're already syncing\n", ann.Src)
		return
	}

	ourHash := h.Hashes().Full
	if ann.Hash == ourHash {
		fmt.Printf("Matching hash from %s\n", ann.Src)
		return
	}

	fmt.Printf("Mismatching hash from %s: %s != %s\n", ann.Src, ann.Hash, ourHash)
	remoteNode := remote.API{
		Address: fmt.Sprintf("%s%s", ann.Src.IP, ann.APIAddr),
	}
	if err := h.StartSync(remoteNode, ann.Hash); err != nil {
		fmt.Printf("Failed to start sync: %v\n", err)
		return
	}
	fmt.Printf("Synchronized with %s\n", remoteNode.Address)
}
