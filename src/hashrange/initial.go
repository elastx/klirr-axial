package hashrange

import (
	"fmt"
	"time"
)

// InitialPeriods returns the coarse-grained time windows a Node uses to
// bootstrap a sync round. The layout is recent-week + 1-month + 6-month +
// 2-year + everything-older, descending in granularity so that recent data
// reconciles fast and stale data is checked at coarse resolution.
func InitialPeriods() []Period {
	earliestStartTime := RealizeStart(nil)
	latestEndTime := RealizeEnd(nil)

	periodSteps := []struct {
		Years  int
		Months int
		Days   int
	}{
		{0, -1, 0},
		{0, -6, 0},
		{-2, 0, 0},
	}

	var previousStart *time.Time
	weekStart := weekStartOf(time.Now())
	if weekStart.Before(earliestStartTime) {
		previousStart = &earliestStartTime
	} else {
		previousStart = &weekStart
	}

	periods := []Period{
		{Start: previousStart, End: &latestEndTime},
	}

	for _, step := range periodSteps {
		start := previousStart.AddDate(step.Years, step.Months, step.Days)
		if start.Before(earliestStartTime) {
			periods = append(periods, Period{
				Start: &earliestStartTime,
				End:   previousStart,
			})
			break
		}
		periods = append(periods, Period{
			Start: &start,
			End:   previousStart,
		})
		previousStart = &start
	}

	periods = append(periods, Period{
		Start: &earliestStartTime,
		End:   previousStart,
	})

	return periods
}

// InitialFingerprintRanges returns the lexicographic partitions a Node uses
// to bootstrap user sync: one window per digit (0-9) and one per
// lower-case letter (a-z). Fingerprints are hex strings, so 0-9 and a-f
// carry actual users; g-z exist to keep the partition uniform without
// special-casing.
func InitialFingerprintRanges() []StringRange {
	var ranges []StringRange
	for i := 0; i < 10; i++ {
		ranges = append(ranges, StringRange{
			Start: fmt.Sprintf("%c", '0'+i),
			End:   fmt.Sprintf("%c", '0'+i+1),
		})
	}
	for i := 0; i < 25; i++ {
		ranges = append(ranges, StringRange{
			Start: fmt.Sprintf("%c", 'a'+i),
			End:   fmt.Sprintf("%c", 'a'+i+1),
		})
	}
	return ranges
}

func weekStartOf(now time.Time) time.Time {
	weekday := now.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	return now.AddDate(0, 0, -int(weekday-time.Monday)).Truncate(24 * time.Hour)
}
