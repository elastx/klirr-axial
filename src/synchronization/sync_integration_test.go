//go:build integration

package synchronization

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"axial/api"
	"axial/models"
	"axial/remote"
	"strings"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type inMemoryRequester struct {
    DB *gorm.DB
}

func (r inMemoryRequester) RequestSync(_ remote.API, req api.SyncRequest) (api.SyncResponse, error) {
    return api.ComputeSyncResponse(r.DB, req)
}

func newTestDB(t *testing.T) *gorm.DB {
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

type testKeys struct {
    senderKey     *crypto.Key
    recipientKey  *crypto.Key
    senderPubArm  string
    recipientPubArm string
    senderFP      string
    recipientFP   string
}

func genKeys(t *testing.T) testKeys {
    t.Helper()
    // Generate two ephemeral keys with empty passphrase
    senderKey, err := crypto.GenerateKey("sender", "sender@example.com", "", 0)
    if err != nil {
        t.Fatalf("generate sender key: %v", err)
    }
    recipientKey, err := crypto.GenerateKey("recipient", "recipient@example.com", "", 0)
    if err != nil {
        t.Fatalf("generate recipient key: %v", err)
    }

    senderPubArm, err := senderKey.GetArmoredPublicKey()
    if err != nil {
        t.Fatalf("armored sender pub: %v", err)
    }
    recipientPubArm, err := recipientKey.GetArmoredPublicKey()
    if err != nil {
        t.Fatalf("armored recipient pub: %v", err)
    }
    senderFP := senderKey.GetFingerprint()
    recipientFP := recipientKey.GetFingerprint()

    return testKeys{
        senderKey:       senderKey,
        recipientKey:    recipientKey,
        senderPubArm:    senderPubArm,
        recipientPubArm: recipientPubArm,
        senderFP:        senderFP,
        recipientFP:     recipientFP,
    }
}

func makeEncryptedSigned(t *testing.T, keys testKeys, plain string) models.Crypto {
    t.Helper()
    msg := crypto.NewPlainMessageFromString(plain)
    senderKR, err := crypto.NewKeyRing(keys.senderKey)
    if err != nil { t.Fatalf("sender ring: %v", err) }
    recipientKR, err := crypto.NewKeyRing(keys.recipientKey)
    if err != nil {
        t.Fatalf("recipient ring: %v", err)
    }
    // Encrypt and sign in one step using sender key ring
    enc, err := recipientKR.Encrypt(msg, senderKR)
    if err != nil {
        t.Fatalf("encrypt+sign: %v", err)
    }
    armored, err := enc.GetArmored()
    if err != nil {
        t.Fatalf("get armored: %v", err)
    }
    return models.Crypto(armored)
}

func insertUser(t *testing.T, db *gorm.DB, pubArm string) models.User {
    t.Helper()
    u := models.User{CreateUser: models.CreateUser{PublicKey: pubArm}}
    if err := db.Create(&u).Error; err != nil {
        t.Fatalf("create user: %v", err)
    }
    return u
}

func insertMessage(t *testing.T, db *gorm.DB, content models.Crypto) models.Message {
    t.Helper()
    m := models.Message{CreateMessage: models.CreateMessage{Content: content}}
    // Manually set ID and CreatedAt since hooks are skipped
    m.Base.BeforeCreate(m.Hash())
    if err := db.Session(&gorm.Session{SkipHooks: true}).Create(&m).Error; err != nil {
        t.Fatalf("create message: %v", err)
    }
    return m
}

func computeFullHash(t *testing.T, db *gorm.DB) models.HashSet {
    t.Helper()
    mh, err := models.GetMessagesHash(db, nil, nil)
    if err != nil { t.Fatalf("messages hash: %v", err) }
    uh, err := models.GetUsersHash(db)
    if err != nil { t.Fatalf("users hash: %v", err) }
    fh, err := models.GetFilesHash(db)
    if err != nil { t.Fatalf("files hash: %v", err) }

    full := models.HashSet{Messages: mh, Users: uh, Files: fh}
    // Combine deterministically same as RefreshHashes
    hasher := sha256.New()
    hasher.Write([]byte("messages:" + mh))
    hasher.Write([]byte("users:" + uh))
    hasher.Write([]byte("files:" + fh))
    sum := hasher.Sum(nil)
    full.Full = hex.EncodeToString(sum)
    return full
}

func TestFullSyncConverges(t *testing.T) {
    dbA := newTestDB(t)
    dbB := newTestDB(t)

    keys := genKeys(t)

    // Seed users on both nodes for sender/recipient
    insertUser(t, dbA, keys.senderPubArm)
    insertUser(t, dbA, keys.recipientPubArm)
    insertUser(t, dbB, keys.senderPubArm)
    insertUser(t, dbB, keys.recipientPubArm)

    // Create three messages, split across nodes
    m1 := makeEncryptedSigned(t, keys, "hello-"+randString(t))
    m2 := makeEncryptedSigned(t, keys, "world-"+randString(t))
    m3 := makeEncryptedSigned(t, keys, "sync-"+randString(t))

    insertMessage(t, dbA, m1)
    insertMessage(t, dbA, m2)
    insertMessage(t, dbB, m2)
    insertMessage(t, dbB, m3)

    // Prepare ranges
    periods, _ := startingSyncRanges()
    hashedA, err := models.GenerateHashRanges(dbA, periods)
    if err != nil { t.Fatalf("hash ranges A: %v", err) }
    hashedB, err := models.GenerateHashRanges(dbB, periods)
    if err != nil { t.Fatalf("hash ranges B: %v", err) }

    nodeA := remote.API{Address: "nodeA"}
    nodeB := remote.API{Address: "nodeB"}

    // Round 1: A syncs with B
    models.DB = dbA.Session(&gorm.Session{SkipHooks: true})
    missingForBFromA, err := SyncWithRequester(inMemoryRequester{DB: dbB}, nodeB, hashedA, nil)
    if err != nil { t.Fatalf("sync A->B: %v", err) }
    // Apply to B
    for _, m := range missingForBFromA {
        if err := dbB.Session(&gorm.Session{SkipHooks: true}).Create(&m).Error; err != nil && !isDuplicate(err) {
            t.Fatalf("apply A->B message: %v", err)
        }
    }

    // Round 2: B syncs with A
    models.DB = dbB.Session(&gorm.Session{SkipHooks: true})
    missingForAFromB, err := SyncWithRequester(inMemoryRequester{DB: dbA}, nodeA, hashedB, nil)
    if err != nil { t.Fatalf("sync B->A: %v", err) }
    // Apply to A
    for _, m := range missingForAFromB {
        if err := dbA.Session(&gorm.Session{SkipHooks: true}).Create(&m).Error; err != nil && !isDuplicate(err) {
            t.Fatalf("apply B->A message: %v", err)
        }
    }

    // Optional extra round to ensure convergence if splits occurred
    models.DB = dbA.Session(&gorm.Session{SkipHooks: true})
    hashedA2, _ := models.GenerateHashRanges(dbA, periods)
    models.DB = dbB.Session(&gorm.Session{SkipHooks: true})
    hashedB2, _ := models.GenerateHashRanges(dbB, periods)

    models.DB = dbA.Session(&gorm.Session{SkipHooks: true})
    moreForB, err := SyncWithRequester(inMemoryRequester{DB: dbB}, nodeB, hashedA2, nil)
    if err != nil { t.Fatalf("sync A->B (2): %v", err) }
    for _, m := range moreForB {
        if err := dbB.Session(&gorm.Session{SkipHooks: true}).Create(&m).Error; err != nil && !isDuplicate(err) {
            t.Fatalf("apply A->B (2): %v", err)
        }
    }

    models.DB = dbB.Session(&gorm.Session{SkipHooks: true})
    moreForA, err := SyncWithRequester(inMemoryRequester{DB: dbA}, nodeA, hashedB2, nil)
    if err != nil { t.Fatalf("sync B->A (2): %v", err) }
    for _, m := range moreForA {
        if err := dbA.Session(&gorm.Session{SkipHooks: true}).Create(&m).Error; err != nil && !isDuplicate(err) {
            t.Fatalf("apply B->A (2): %v", err)
        }
    }

    // Verify both sides have identical full hashes
    fullA := computeFullHash(t, dbA)
    fullB := computeFullHash(t, dbB)
    if fullA.Full != fullB.Full || fullA.Messages != fullB.Messages || fullA.Users != fullB.Users || fullA.Files != fullB.Files {
        t.Fatalf("databases did not converge after sync: A=%+v B=%+v", fullA, fullB)
    }
}

func randString(t *testing.T) string {
    t.Helper()
    buf := make([]byte, 8)
    if _, err := rand.Read(buf); err != nil { t.Fatalf("rand: %v", err) }
    return time.Now().UTC().Format(time.RFC3339Nano) + string(buf)
}

func isDuplicate(err error) bool {
    if models.IsDuplicateError(err) {
        return true
    }
    s := strings.ToLower(err.Error())
    return strings.Contains(s, "unique") || strings.Contains(s, "constraint failed") || strings.Contains(s, "duplicate")
}
