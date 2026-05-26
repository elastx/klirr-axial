package synchronization

import "strings"

// reconcile compares the items the remote Node sent us against the items we
// have locally for the same Hash Range, inserting remote-only items and
// returning local-only items so the caller can push them to the remote in
// the next phase of the sync round.
//
// The function is parameterised over the resource type so Messages,
// Bulletins, and Users all share the same algorithm. Identity is supplied
// by `equal` — Messages and Bulletins use ID equality; Users additionally
// match across the synthetic fingerprint groups the test harness produces.
//
// `insertSkip` is consulted when `insert` fails: returning true swallows
// the error (used to absorb duplicate-key races during concurrent sync).
func reconcile[T any](
	loader func() ([]T, error),
	remoteItems []T,
	equal func(a, b T) bool,
	insert func(T) error,
	insertSkip func(error) bool,
) (missingInRemote []T, err error) {
	ourItems, err := loader()
	if err != nil {
		return nil, err
	}

	for _, item := range remoteItems {
		if containsEq(ourItems, item, equal) {
			continue
		}
		if err := insert(item); err != nil {
			if insertSkip == nil || !insertSkip(err) {
				return nil, err
			}
		}
	}

	for _, ourItem := range ourItems {
		if !containsEq(remoteItems, ourItem, equal) {
			missingInRemote = append(missingInRemote, ourItem)
		}
	}

	return missingInRemote, nil
}

func containsEq[T any](haystack []T, needle T, equal func(a, b T) bool) bool {
	for _, h := range haystack {
		if equal(h, needle) {
			return true
		}
	}
	return false
}

// isDuplicateKeyErr is the legacy string-based duplicate-key check used by
// the sync ingestion loops. The PostgreSQL-specific check in
// models.IsDuplicateError is more precise but does not match SQLite's
// "UNIQUE constraint failed" wording the tests rely on.
func isDuplicateKeyErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate key")
}
