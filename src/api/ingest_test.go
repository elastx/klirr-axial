package api

import (
	"testing"

	"axial/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newIngestTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Message{}, &models.Bulletin{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestIngest_EmptyBatchStillRefreshesHashes(t *testing.T) {
	db := newIngestTestDB(t)
	if err := ingest[models.User](db, nil); err != nil {
		t.Fatalf("ingest empty: %v", err)
	}
	// RefreshHashes populates the in-process hash cache; reading it back
	// proves the side effect ran.
	hs := models.GetHashes()
	if hs.Full == "" {
		t.Error("expected Full hash to be set after RefreshHashes, got empty")
	}
}

func TestIngest_DuplicateKeysAreSwallowed(t *testing.T) {
	db := newIngestTestDB(t)

	// Two users with the same fingerprint trip the unique-index. Skipping
	// hooks lets us bypass the BeforeCreate PGP validation so the test
	// stays focused on the duplicate-suppression branch.
	u := models.User{Fingerprint: "fp-dup-test"}
	u.Base.ID = "fp-dup-test"
	raw := db.Session(&gorm.Session{SkipHooks: true})
	if err := raw.Create(&u).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	// Now ingest the same record again — should not error.
	if err := ingest[models.User](raw, []models.User{u}); err != nil {
		t.Errorf("duplicate-key should be swallowed, got %v", err)
	}
}

func TestIngest_NonDuplicateErrorsPropagate(t *testing.T) {
	// Closing the DB session and pointing ingest at it forces a non-duplicate
	// error to surface; the caller (the handler) must see it as a 500.
	db := newIngestTestDB(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql.DB: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	err = ingest[models.User](db, []models.User{{Fingerprint: "fp"}})
	if err == nil {
		t.Error("expected error from ingest into closed DB, got nil")
	}
}
