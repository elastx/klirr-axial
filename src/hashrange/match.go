package hashrange

// MismatchedPeriods returns the entries from `theirs` whose Period bounds
// match an entry in `ours` but whose Hash differs. Used by the sync round
// orchestrator to decide which time windows still need drilling after a
// hash exchange.
//
// The comparison is order-independent and quadratic in the input size;
// the inputs are tiny (a handful of periods per round) so this is fine.
func MismatchedPeriods(ours, theirs []HashedPeriod) []HashedPeriod {
	out := []HashedPeriod{}
	for _, our := range ours {
		ourStart := RealizeStart(our.Start)
		ourEnd := RealizeEnd(our.End)
		for _, their := range theirs {
			theirStart := RealizeStart(their.Start)
			theirEnd := RealizeEnd(their.End)
			if ourStart.Equal(theirStart) && ourEnd.Equal(theirEnd) && our.Hash != their.Hash {
				out = append(out, their)
			}
		}
	}
	return out
}
