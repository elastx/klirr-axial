package hashrange

import (
	"testing"
	"time"
)

func TestRealizeStart_DefaultsToReleaseYear(t *testing.T) {
	got := RealizeStart(nil)
	want := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("RealizeStart(nil) = %v, want %v", got, want)
	}
}

func TestRealizeStart_PassesThroughNonNil(t *testing.T) {
	want := time.Date(2030, 6, 15, 12, 0, 0, 0, time.UTC)
	got := RealizeStart(&want)
	if !got.Equal(want) {
		t.Errorf("RealizeStart(&t) = %v, want %v", got, want)
	}
}

func TestRealizeEnd_DefaultsToNow(t *testing.T) {
	before := time.Now()
	got := RealizeEnd(nil)
	after := time.Now()
	if got.Before(before) || got.After(after) {
		t.Errorf("RealizeEnd(nil) = %v, want within [%v, %v]", got, before, after)
	}
}

func TestRealizeEnd_PassesThroughNonNil(t *testing.T) {
	want := time.Date(2030, 6, 15, 12, 0, 0, 0, time.UTC)
	got := RealizeEnd(&want)
	if !got.Equal(want) {
		t.Errorf("RealizeEnd(&t) = %v, want %v", got, want)
	}
}
