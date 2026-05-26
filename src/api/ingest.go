package api

import (
	"strings"

	"axial/models"

	"gorm.io/gorm"
)

// ingest writes a batch of remote-sourced records into the DB on behalf of
// the sync handlers (Messages, Bulletins, Users). Duplicate-key errors are
// swallowed on a per-item basis — those records were already synced from
// another peer. Once the batch is in, the database content hashes are
// recomputed exactly once so the next multicast broadcast advertises the
// new Full Hash.
//
// The function is the seam between transport (the HTTP handler) and the
// post-write side effects (insert + RefreshHashes). Tests can call it
// directly against an in-memory SQLite without spinning up an HTTP server,
// which is how the per-handler tests exercise the duplicate-suppression
// and RefreshHashes-on-completion behaviour.
func ingest[T any](db *gorm.DB, items []T) error {
	for i := range items {
		if err := db.Create(&items[i]).Error; err != nil {
			if isDuplicate(err) {
				continue
			}
			return err
		}
	}
	return models.RefreshHashes(db)
}

// isDuplicate matches either Postgres' SQLSTATE 23505 (via
// models.IsDuplicateError) or SQLite's "UNIQUE constraint failed" wording.
// Sync rounds are eventually-consistent — a duplicate ingest happens
// whenever the same content arrives from two peers — so absorption must
// work on both drivers, not just the production one.
func isDuplicate(err error) bool {
	if err == nil {
		return false
	}
	if models.IsDuplicateError(err) {
		return true
	}
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
