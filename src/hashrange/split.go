package hashrange

import "time"

// SplitTimeRange divides a Period into n equal sub-Periods. Used during the
// drill phase of a sync round: when a window's hash mismatches and the
// data inside it is too large to ship whole, the window is split and each
// piece gets re-hashed in the next round.
//
// The final piece is anchored to the original End to absorb any rounding
// drift, so the union of returned Periods exactly covers the input.
func SplitTimeRange(period Period, n int) []Period {
	start := RealizeStart(period.Start)
	end := RealizeEnd(period.End)

	duration := end.Sub(start)
	partDuration := duration / time.Duration(n)

	ranges := make([]Period, n)
	for i := 0; i < n; i++ {
		partStart := start.Add(partDuration * time.Duration(i))
		partEnd := partStart.Add(partDuration)
		if i == n-1 {
			partEnd = end
		}

		ranges[i] = Period{
			Start: &partStart,
			End:   &partEnd,
		}
	}

	return ranges
}
