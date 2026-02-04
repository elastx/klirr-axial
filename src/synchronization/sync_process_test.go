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
	if err := db.AutoMigrate(&models.User{}, &models.Message{}, &models.Bulletin{}); err != nil {
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

func insertBulletinRawUnit(t *testing.T, db *gorm.DB, topic string, content models.Crypto, parentId string) models.Bulletin {
	t.Helper()
	b := models.Bulletin{CreateBulletin: models.CreateBulletin{Topic: topic, Content: content, ParentID: &parentId}}
	b.Base.ID = b.Hash()
	b.Base.BeforeCreate(nil)
	if err := db.Session(&gorm.Session{SkipHooks: true}).Create(&b).Error; err != nil {
		t.Fatalf("create bulletin: %v", err)
	}
	return b
}

func insertUserRawUnit(t *testing.T, db *gorm.DB, fingerprint string) models.User {
	t.Helper()
	u := models.User{Fingerprint: fingerprint}
	u.Base.ID = u.Hash()
	u.Base.BeforeCreate(nil)
	if err := db.Session(&gorm.Session{SkipHooks: true}).Create(&u).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return u
}

func randStringUnit(t *testing.T) string {
	t.Helper()
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("rand: %v", err)
	}
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

	out := mismatchedMessagesPeriods(our, theirs)
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

	// Seed bulletins (synthetic content, skip hooks)
	insertBulletinRawUnit(t, dbA, "topic1", models.Crypto("b1-"+randStringUnit(t)), "")
	b2 := insertBulletinRawUnit(t, dbA, "topic2", models.Crypto("b2-"+randStringUnit(t)), "")
	insertBulletinRawUnit(t, dbB, "topic2", models.Crypto(string(b2.Content)), b2.Base.ID)
	insertBulletinRawUnit(t, dbB, "topic3", models.Crypto("b3-"+randStringUnit(t)), "")

	// Seed user profiles (synthetic fingerprints, skip hooks)
	insertUserRawUnit(t, dbA, "FP_A1_"+randStringUnit(t))
	insertUserRawUnit(t, dbA, "FP_A2_"+randStringUnit(t))
	insertUserRawUnit(t, dbB, "FP_A2_"+randStringUnit(t)) // share FP_A2 on B
	insertUserRawUnit(t, dbB, "FP_B3_"+randStringUnit(t))

	periods, _ := startingSyncRanges()
	hashedMessagesA, err := models.GetMessagesHashRanges(dbA, periods)
	if err != nil {
		t.Fatalf("hash messages ranges A: %v", err)
	}

	hashedBulletinsA, err := models.GetBulletinsHashRanges(dbA, periods)
	if err != nil {
		t.Fatalf("hash bulletins ranges A: %v", err)
	}

	hashedUsersA, err := models.GetUsersHashRanges(dbA, []models.StringRange{{Start: "", End: "zzzz"}})
	if err != nil {
		t.Fatalf("hash users ranges A: %v", err)
	}

	hashedMessagesB, err := models.GetMessagesHashRanges(dbB, periods)
	if err != nil {
		t.Fatalf("hash messages ranges B: %v", err)
	}

	hashedBulletinsB, err := models.GetBulletinsHashRanges(dbB, periods)
	if err != nil {
		t.Fatalf("hash bulletins ranges B: %v", err)
	}

	hashedUsersB, err := models.GetUsersHashRanges(dbB, []models.StringRange{{Start: "", End: "zzzz"}})
	if err != nil {
		t.Fatalf("hash users ranges B: %v", err)
	}

	nodeA := remote.API{Address: "nodeA"}
	nodeB := remote.API{Address: "nodeB"}

	// Round 1: A pulls from B and computes messages to send to B
	models.DB = dbA.Session(&gorm.Session{SkipHooks: true})
	missingMessagesForBFromA, missingBulletinsForBFromA, missingUsersForBFromA, err := SyncWithRequester(fakeRequester{DB: dbB}, nodeB, hashedMessagesA, hashedBulletinsA, hashedUsersA)
	if err != nil {
		t.Fatalf("sync A->B: %v", err)
	}
	// Apply remote messages to A
	for _, m := range missingMessagesForBFromA {
		if err := dbB.Session(&gorm.Session{SkipHooks: true}).Create(&m).Error; err != nil && !isDuplicateUnit(err) {
			t.Fatalf("apply A->B message: %v", err)
		}
	}
	for _, b := range missingBulletinsForBFromA {
		if err := dbB.Session(&gorm.Session{SkipHooks: true}).Create(&b).Error; err != nil && !isDuplicateUnit(err) {
			t.Fatalf("apply A->B bulletin: %v", err)
		}
	}
	for _, u := range missingUsersForBFromA {
		if err := dbB.Session(&gorm.Session{SkipHooks: true}).Create(&u).Error; err != nil && !isDuplicateUnit(err) {
			t.Fatalf("apply A->B user: %v", err)
		}
	}

	// Round 2: B pulls from A and applies
	models.DB = dbB.Session(&gorm.Session{SkipHooks: true})
	missingMessagesForAFromB, missingBulletinsForAFromB, missingUsersForAFromB, err := SyncWithRequester(fakeRequester{DB: dbA}, nodeA, hashedMessagesB, hashedBulletinsB, hashedUsersB)
	if err != nil {
		t.Fatalf("sync B->A: %v", err)
	}
	for _, m := range missingMessagesForAFromB {
		if err := dbA.Session(&gorm.Session{SkipHooks: true}).Create(&m).Error; err != nil && !isDuplicateUnit(err) {
			t.Fatalf("apply B->A message: %v", err)
		}
	}
	for _, b := range missingBulletinsForAFromB {
		if err := dbA.Session(&gorm.Session{SkipHooks: true}).Create(&b).Error; err != nil && !isDuplicateUnit(err) {
			t.Fatalf("apply B->A bulletin: %v", err)
		}
	}
	for _, u := range missingUsersForAFromB {
		if err := dbA.Session(&gorm.Session{SkipHooks: true}).Create(&u).Error; err != nil && !isDuplicateUnit(err) {
			t.Fatalf("apply B->A user: %v", err)
		}
	}

	// both DBs should have all 3 messages
	var countA int64
	if err := dbA.Model(&models.Message{}).Count(&countA).Error; err != nil {
		t.Fatalf("count messages A: %v", err)
	}
	if countA != 3 {
		t.Fatalf("expected 3 messages in A, got %d", countA)
	}

	var countB int64
	if err := dbB.Model(&models.Message{}).Count(&countB).Error; err != nil {
		t.Fatalf("count messages B: %v", err)
	}
	if countB != 3 {
		t.Fatalf("expected 3 messages in B, got %d", countB)
	}

	// both DBs should have all 3 bulletins
	var bcountA int64
	if err := dbA.Model(&models.Bulletin{}).Count(&bcountA).Error; err != nil {
		t.Fatalf("count bulletins A: %v", err)
	}
	if bcountA != 3 {
		t.Fatalf("expected 3 bulletins in A, got %d", bcountA)
	}

	var bcountB int64
	if err := dbB.Model(&models.Bulletin{}).Count(&bcountB).Error; err != nil {
		t.Fatalf("count bulletins B: %v", err)
	}
	if bcountB != 3 {
		t.Fatalf("expected 3 bulletins in B, got %d", bcountB)
	}

	// both DBs should have all 3 users
	var ucountA int64
	if err := dbA.Model(&models.User{}).Count(&ucountA).Error; err != nil {
		t.Fatalf("count users A: %v", err)
	}
	if ucountA != 3 {
		t.Fatalf("expected 3 users in A, got %d", ucountA)
	}

	var ucountB int64
	if err := dbB.Model(&models.User{}).Count(&ucountB).Error; err != nil {
		t.Fatalf("count users B: %v", err)
	}
	if ucountB != 3 {
		t.Fatalf("expected 3 users in B, got %d", ucountB)
	}
}
