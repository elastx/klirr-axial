package hashrange

import (
	"testing"
)

func TestInitialPeriods_ProducesAtLeastOneWindow(t *testing.T) {
	periods := InitialPeriods()
	if len(periods) == 0 {
		t.Fatal("expected at least one initial period")
	}
}

func TestInitialPeriods_AllWindowsHaveResolvedBounds(t *testing.T) {
	for i, p := range InitialPeriods() {
		if p.Start == nil {
			t.Errorf("period %d has nil Start", i)
		}
		if p.End == nil {
			t.Errorf("period %d has nil End", i)
		}
		if p.Start != nil && p.End != nil && !p.End.After(*p.Start) {
			t.Errorf("period %d has non-positive duration: %v..%v", i, *p.Start, *p.End)
		}
	}
}

func TestInitialFingerprintRanges_CoversDigitsAndLowercase(t *testing.T) {
	ranges := InitialFingerprintRanges()
	if len(ranges) != 35 { // 10 digits + 25 letters
		t.Errorf("expected 35 ranges (10 digits + 25 letters), got %d", len(ranges))
	}
	if ranges[0].Start != "0" {
		t.Errorf("first range should start at '0', got %q", ranges[0].Start)
	}
	if ranges[9].Start != "9" {
		t.Errorf("tenth range should start at '9', got %q", ranges[9].Start)
	}
	if ranges[10].Start != "a" {
		t.Errorf("eleventh range should start at 'a', got %q", ranges[10].Start)
	}
}
