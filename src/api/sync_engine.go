package api

import (
	"axial/models"
	"math"
	"sort"
	"strings"
)

// SyncStore defines the minimal DB operations the sync engine needs.
// Implementations can wrap real databases or in-memory test stores.
type SyncStore interface {
    GenerateHashRanges(periods []models.Period) ([]models.HashedPeriod, error)
    CountMessagesByPeriod(period models.Period) int64
    GetMessagesByPeriod(period models.Period) ([]models.Message, error)
    GetDatabaseHashes() (models.HashSet, error)
    GetUsersHashByFingerprintRange(start, end string) (string, error)
    CountUsersByFingerprintRange(start, end string) (int64, error)
    GetUsersByFingerprintRange(start, end string) ([]models.User, error)
    UpsertMessage(msg models.Message) error
}

// ModelSyncStore is the default store backed by models.DB.
type ModelSyncStore struct{}

func (s *ModelSyncStore) GenerateHashRanges(periods []models.Period) ([]models.HashedPeriod, error) {
    return models.GenerateHashRanges(models.DB, periods)
}

func (s *ModelSyncStore) CountMessagesByPeriod(period models.Period) int64 {
    return models.CountMessagesByPeriod(models.DB, period)
}

func (s *ModelSyncStore) GetMessagesByPeriod(period models.Period) ([]models.Message, error) {
    return models.GetMessagesByPeriod(models.DB, period)
}

func (s *ModelSyncStore) GetDatabaseHashes() (models.HashSet, error) {
    return models.GetDatabaseHashes(models.DB)
}

func (s *ModelSyncStore) GetUsersHashByFingerprintRange(start, end string) (string, error) {
    return models.GetUsersHashByFingerprintRange(models.DB, start, end)
}

func (s *ModelSyncStore) CountUsersByFingerprintRange(start, end string) (int64, error) {
    return models.CountUsersByFingerprintRange(models.DB, start, end)
}

func (s *ModelSyncStore) GetUsersByFingerprintRange(start, end string) ([]models.User, error) {
    return models.GetUsersByFingerprintRange(models.DB, start, end)
}

func (s *ModelSyncStore) UpsertMessage(msg models.Message) error {
    if err := models.DB.Create(&msg).Error; err != nil {
        // Ignore duplicate key errors since those messages were already synced
        if strings.Contains(err.Error(), "duplicate key") {
            return nil
        }
        return err
    }
    return nil
}

// FindMismatchingRanges returns our hashed periods that mismatch the hashes
// provided by the remote for the same start/end windows.
func FindMismatchingRanges(store SyncStore, incoming []models.HashedPeriod) ([]models.HashedPeriod, error) {
    periods := make([]models.Period, 0, len(incoming))
    for _, hp := range incoming {
        periods = append(periods, models.Period{Start: hp.Start, End: hp.End})
    }

    ourRanges, err := store.GenerateHashRanges(periods)
    if err != nil {
        return nil, err
    }

    mismatching := []models.HashedPeriod{}
    for _, our := range ourRanges {
        ourStart := models.RealizeStart(our.Start)
        ourEnd := models.RealizeEnd(our.End)
        for _, theirs := range incoming {
            theirStart := models.RealizeStart(theirs.Start)
            theirEnd := models.RealizeEnd(theirs.End)
            if ourStart == theirStart && ourEnd == theirEnd && our.Hash != theirs.Hash {
                mismatching = append(mismatching, our)
            }
        }
    }
    return mismatching, nil
}

// BuildSyncResponse constructs a SyncResponse deterministically based on
// mismatching ranges and a batch size limit. It performs no I/O beyond store calls.
func BuildSyncResponse(store SyncStore, req SyncRequest, maxBatchSize int) (SyncResponse, error) {
    resp := SyncResponse{}

    // Database full-hash snapshot
    hashes, err := store.GetDatabaseHashes()
    if err != nil {
        return resp, err
    }
    resp.Hashes = hashes

    // Find mismatches against incoming ranges
    mismatchingRanges, err := FindMismatchingRanges(store, req.Ranges)
    if err != nil {
        return resp, err
    }

    // Rank ranges by message count ascending
    type ranked struct {
        idx   int
        count int64
    }
    ranks := make([]ranked, 0, len(mismatchingRanges))
    for i, r := range mismatchingRanges {
        period := models.Period{Start: r.Start, End: r.End}
        ranks = append(ranks, ranked{idx: i, count: store.CountMessagesByPeriod(period)})
    }
    sort.Slice(ranks, func(i, j int) bool { return ranks[i].count < ranks[j].count })

    totalPlain := int64(0)
    for _, rk := range ranks {
        mr := mismatchingRanges[rk.idx]
        period := models.Period{Start: mr.Start, End: mr.End}
        if totalPlain+rk.count < int64(maxBatchSize) {
            msgs, err := store.GetMessagesByPeriod(period)
            if err != nil {
                return resp, err
            }
            resp.Messages = append(resp.Messages, models.MessagesPeriod{Period: mr.Period, Messages: msgs})
            totalPlain += rk.count
            continue
        }

        // Split oversized ranges into hashed subranges.
        // Strategy:
        // - If the range is "very large" (>= 10x maxBatch), split aggressively so that
        //   each subrange is roughly <= maxBatch to enable plain message transfer soon.
        // - Otherwise, keep a single hashed range to avoid over-sharding in modest mismatches.
        splits := 1
        if rk.count >= int64(maxBatchSize)*10 {
            splits = int(math.Ceil(float64(rk.count) / float64(maxBatchSize)))
            if splits < 2 { // ensure at least 2 parts when we decide to split
                splits = 2
            }
        }
        subPeriods := models.SplitTimeRange(period, splits)
        hashedSubs, err := store.GenerateHashRanges(subPeriods)
        if err != nil {
            return resp, err
        }
        for _, hp := range hashedSubs {
            resp.Ranges = append(resp.Ranges, models.HashedPeriod{Period: hp.Period, Hash: hp.Hash})
        }
    }

    // Add user range mismatches, if any
    if err := addUserRangeMismatches(store, req, &resp, maxBatchSize); err != nil {
        return resp, err
    }

    return resp, nil
}

// Compare user hashed ranges and include mismatches in response.
func addUserRangeMismatches(store SyncStore, req SyncRequest, resp *SyncResponse, maxBatchSize int) error {
    for _, userRange := range req.Users {
        ourHash, err := store.GetUsersHashByFingerprintRange(userRange.Start, userRange.End)
        if err != nil {
            return err
        }
        if ourHash != userRange.Hash {
            resp.UserRangeHashes = append(resp.UserRangeHashes, models.HashedUsersRange{
                StringRange: models.StringRange{Start: userRange.Start, End: userRange.End},
                Hash:        ourHash,
            })

            // If the mismatch is small enough, include actual users to allow immediate convergence.
            // Note: No signature filtering; all users are eligible for sync.
            if count, err := store.CountUsersByFingerprintRange(userRange.Start, userRange.End); err == nil {
                if count > 0 && count <= int64(maxBatchSize) {
                    if users, err := store.GetUsersByFingerprintRange(userRange.Start, userRange.End); err == nil {
                        resp.Users = append(resp.Users, models.UsersRange{
                            StringRange: models.StringRange{Start: userRange.Start, End: userRange.End},
                            Users:       users,
                        })
                    }
                }
            }
        }
    }
    return nil
}
