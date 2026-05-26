package hashrange

import (
	"testing"
	"time"
)

func hp(start, end time.Time, hash string) HashedPeriod {
	return HashedPeriod{
		Period: Period{Start: &start, End: &end},
		Hash:   hash,
	}
}

func TestMismatchedPeriods_ReturnsOnlyDifferingBoundsMatch(t *testing.T) {
	t1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	t3 := time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC)

	ours := []HashedPeriod{
		hp(t1, t2, "AAA"),
		hp(t2, t3, "BBB"),
	}
	theirs := []HashedPeriod{
		hp(t1, t2, "AAA"), // same window, same hash — not mismatched
		hp(t2, t3, "ZZZ"), // same window, different hash — mismatched
	}

	got := MismatchedPeriods(ours, theirs)
	if len(got) != 1 {
		t.Fatalf("expected 1 mismatch, got %d: %+v", len(got), got)
	}
	if got[0].Hash != "ZZZ" {
		t.Errorf("returned hash should be theirs (ZZZ), got %q", got[0].Hash)
	}
}

func TestMismatchedPeriods_IgnoresNonOverlappingBounds(t *testing.T) {
	t1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	t3 := time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC)
	t4 := time.Date(2025, 1, 4, 0, 0, 0, 0, time.UTC)

	ours := []HashedPeriod{hp(t1, t2, "AAA")}
	theirs := []HashedPeriod{hp(t3, t4, "BBB")} // different window entirely

	got := MismatchedPeriods(ours, theirs)
	if len(got) != 0 {
		t.Errorf("expected no mismatch (windows don't overlap), got %+v", got)
	}
}

func TestMismatchedPeriods_FullyMatchingSetReturnsEmpty(t *testing.T) {
	t1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)

	ours := []HashedPeriod{hp(t1, t2, "AAA")}
	theirs := []HashedPeriod{hp(t1, t2, "AAA")}

	got := MismatchedPeriods(ours, theirs)
	if len(got) != 0 {
		t.Errorf("matching set should produce no mismatches, got %+v", got)
	}
}
