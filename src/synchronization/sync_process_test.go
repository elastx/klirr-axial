package synchronization

import (
	"crypto/rand"
	"strings"
	"testing"
	"time"

	"axial/api"
	"axial/models"
	"axial/remote"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// fakeRequester uses the target DB to compute the real sync response
// via api.ComputeSyncResponse, avoiding HTTP for unit tests.
type fakeRequester struct {
    DB *gorm.DB
}

// We can't import *gorm.DB type alias directly here without extra deps, so the
// test skeleton below demonstrates structure without full DB wiring. The real
// implementation should hold a *gorm.DB and call api.ComputeSyncResponse.

func (f fakeRequester) RequestSync(node remote.API, req api.SyncRequest) (api.SyncResponse, error) {
    return api.ComputeSyncResponse(f.DB, req)
}

func newTestDBUnit(t *testing.T) *gorm.DB {
    t.Helper()
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        t.Fatalf("failed to open sqlite memory DB: %v", err)
    }
    if err := db.AutoMigrate(&models.User{}, &models.Message{}, &models.Bulletin{}, &models.File{}); err != nil {
        t.Fatalf("failed to migrate: %v", err)
    }
    return db
}

func insertMessageRawUnit(t *testing.T, db *gorm.DB, content models.Crypto) models.Message {
    t.Helper()
    m := models.Message{CreateMessage: models.CreateMessage{Content: content}}
    m.Base.ID = m.Hash()
    m.Base.BeforeCreate(nil)
    if err := db.Session(&gorm.Session{SkipHooks: true}).Create(&m).Error; err != nil {
        t.Fatalf("create message: %v", err)
    }
    return m
}

func randStringUnit(t *testing.T) string {
    t.Helper()
    buf := make([]byte, 8)
    if _, err := rand.Read(buf); err != nil { t.Fatalf("rand: %v", err) }
    return time.Now().UTC().Format(time.RFC3339Nano) + string(buf)
}

func isDuplicateUnit(err error) bool {
    if models.IsDuplicateError(err) {
        return true
    }
    s := err.Error()
    return containsAnyUnit(s, []string{"unique", "constraint failed", "duplicate"})
}

func containsAnyUnit(s string, needles []string) bool {
    for _, n := range needles {
        if strings.Contains(strings.ToLower(s), strings.ToLower(n)) {
            return true
        }
    }
    return false
}

func TestMismatchedPeriods(t *testing.T) {
    now := time.Now().UTC()
    earlier := now.Add(-24 * time.Hour)

    our := []models.HashedPeriod{
        {Period: models.Period{Start: &earlier, End: &now}, Hash: "aaa"},
    }
    theirs := []models.HashedPeriod{
        {Period: models.Period{Start: &earlier, End: &now}, Hash: "bbb"},
    }

    out := mismatchedPeriods(our, theirs)
    if len(out) != 1 {
        t.Fatalf("expected 1 mismatched period, got %d", len(out))
    }
    if out[0].Hash != "bbb" {
        t.Fatalf("expected remote hash 'bbb', got '%s'", out[0].Hash)
    }
}

func TestSyncExchangeSkeleton(t *testing.T) {
    // Minimal working exchange: two DBs, split messages, run SyncWithRequester both ways.
    dbA := newTestDBUnit(t)
    dbB := newTestDBUnit(t)

    // Seed messages (synthetic content, skip hooks)
    insertMessageRawUnit(t, dbA, models.Crypto("m1-"+randStringUnit(t)))
    m2 := insertMessageRawUnit(t, dbA, models.Crypto("m2-"+randStringUnit(t)))
    insertMessageRawUnit(t, dbB, models.Crypto(string(m2.Content))) // share m2 on B
    insertMessageRawUnit(t, dbB, models.Crypto("m3-"+randStringUnit(t)))

    periods, _ := startingSyncRanges()
    hashedA, err := models.GetMessagesHashRanges(dbA, periods)
    if err != nil { t.Fatalf("hash ranges A: %v", err) }
    hashedB, err := models.GetMessagesHashRanges(dbB, periods)
    if err != nil { t.Fatalf("hash ranges B: %v", err) }

    nodeA := remote.API{Address: "nodeA"}
    nodeB := remote.API{Address: "nodeB"}

    // Round 1: A pulls from B and computes messages to send to B
    models.DB = dbA.Session(&gorm.Session{SkipHooks: true})
    missingForBFromA, err := SyncWithRequester(fakeRequester{DB: dbB}, nodeB, hashedA, nil)
    if err != nil { t.Fatalf("sync A->B: %v", err) }
    // Apply remote messages to A
    for _, m := range missingForBFromA {
        if err := dbB.Session(&gorm.Session{SkipHooks: true}).Create(&m).Error; err != nil && !isDuplicateUnit(err) {
            t.Fatalf("apply A->B message: %v", err)
        }
    }

    // Round 2: B pulls from A and applies
    models.DB = dbB.Session(&gorm.Session{SkipHooks: true})
    missingForAFromB, err := SyncWithRequester(fakeRequester{DB: dbA}, nodeA, hashedB, nil)
    if err != nil { t.Fatalf("sync B->A: %v", err) }
    for _, m := range missingForAFromB {
        if err := dbA.Session(&gorm.Session{SkipHooks: true}).Create(&m).Error; err != nil && !isDuplicateUnit(err) {
            t.Fatalf("apply B->A message: %v", err)
        }
    }
}
