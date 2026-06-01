package hashrange

import (
	"testing"
	"time"
)

func TestSplitTimeRange_ProducesNSubPeriods(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC) // 10 days
	period := Period{Start: &start, End: &end}

	parts := SplitTimeRange(period, 10)
	if len(parts) != 10 {
		t.Fatalf("expected 10 parts, got %d", len(parts))
	}
}

func TestSplitTimeRange_CoversInputExactly(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC)
	period := Period{Start: &start, End: &end}

	parts := SplitTimeRange(period, 7) // n that doesn't divide evenly
	if !parts[0].Start.Equal(start) {
		t.Errorf("first part starts at %v, want %v", *parts[0].Start, start)
	}
	if !parts[len(parts)-1].End.Equal(end) {
		t.Errorf("last part ends at %v, want %v (rounding-drift absorber)", *parts[len(parts)-1].End, end)
	}
	for i := 1; i < len(parts); i++ {
		if !parts[i].Start.Equal(*parts[i-1].End) {
			t.Errorf("part %d Start %v != part %d End %v (gap or overlap)",
				i, *parts[i].Start, i-1, *parts[i-1].End)
		}
	}
}

func TestSplitTimeRange_NEqualsOneReturnsInput(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC)
	period := Period{Start: &start, End: &end}

	parts := SplitTimeRange(period, 1)
	if len(parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(parts))
	}
	if !parts[0].Start.Equal(start) || !parts[0].End.Equal(end) {
		t.Errorf("single-part split should reproduce input; got %v..%v", *parts[0].Start, *parts[0].End)
	}
}

func TestSplitTimeRange_RealizesNilBounds(t *testing.T) {
	// nil bounds should be resolved (release year for Start, now for End) and
	// produce well-formed sub-Periods rather than nil-pointer panics.
	parts := SplitTimeRange(Period{}, 3)
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(parts))
	}
	for i, p := range parts {
		if p.Start == nil || p.End == nil {
			t.Errorf("part %d has nil bound: %+v", i, p)
		}
	}
}
