// Package hashrange owns the Hash Range domain concept: the bounded windows
// (time or fingerprint) over which a Node computes content hashes during a
// sync round, plus the math used to drill those windows down to the smallest
// mismatching slice.
//
// ARCHITECTURE.md treats Hash Range as a first-class glossary term; this
// package is its home. DB-touching code that *produces* hashes for a range
// lives in the models package — hashrange itself depends on nothing but the
// standard library so its math is unit-testable in isolation.
package hashrange

import "time"

// Period is a half-open time window with optional bounds. A nil Start means
// "from the dawn of the dataset" (see RealizeStart); a nil End means "up to
// now" (see RealizeEnd).
type Period struct {
	Start *time.Time `json:"start,omitempty"`
	End   *time.Time `json:"end,omitempty"`
}

// HashedPeriod is a Period together with the content hash a Node computed
// for that window. Exchanged during a sync round so peers can decide which
// windows differ.
type HashedPeriod struct {
	Period
	Hash string `json:"hash"`
}

// StringRange is a half-open lexicographic window — used for partitioning
// the Fingerprint key-space during user sync.
type StringRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// HashedUsersRange is a StringRange together with the hash of the user
// fingerprints inside it.
type HashedUsersRange struct {
	StringRange
	Hash string `json:"hash"`
}

// HashedFilesRange mirrors HashedUsersRange for file sync (placeholder for
// the FUTURE file-attachment work in ARCHITECTURE.md).
type HashedFilesRange struct {
	StringRange
	Hash string `json:"hash"`
}

// RealizeStart resolves a possibly-nil Period.Start to a concrete time.
// nil means "from the release year (2025-01-01 UTC)" — the earliest data
// that can exist on the network.
func RealizeStart(start *time.Time) time.Time {
	if start == nil {
		return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	return *start
}

// RealizeEnd resolves a possibly-nil Period.End to a concrete time. nil
// means "up to now."
func RealizeEnd(end *time.Time) time.Time {
	if end == nil {
		return time.Now()
	}
	return *end
}
