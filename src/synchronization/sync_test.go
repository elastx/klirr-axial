package synchronization

import (
	"axial/api"
	"axial/models"
	"fmt"
	"testing"
	"time"
)

type NodeState struct {
	Messages  []models.Message
	Bulletins []models.Bulletin
	Users     []models.User
}

var emptyNodeState = NodeState{
	Messages:  []models.Message{},
	Bulletins: []models.Bulletin{},
	Users:     []models.User{},
}

var timezone = time.FixedZone("UTC+2", 2*60*60)

var user1 = models.Fingerprint("user1fingerprint")
var user2 = models.Fingerprint("user2fingerprint")
var user3 = models.Fingerprint("user3fingerprint")

var messageContent1 = models.Crypto("EncryptedMessageContent1")
var messageContent2 = models.Crypto("EncryptedMessageContent2")

var simpleNodeState = NodeState{
	Messages: []models.Message{
		{
			Base: models.Base{
				ID:        "msg1",
				CreatedAt: time.Date(2025, 12, 31, 13, 37, 0, 0, timezone),
			},
			Sender:     user1,
			Recipients: models.Fingerprints{user2, user3},
			CreateMessage: models.CreateMessage{
				Content: messageContent1,
			},
		},
	},
	Bulletins: []models.Bulletin{
		{
			Base: models.Base{
				ID:        "bulletin1",
				CreatedAt: time.Date(2025, 12, 31, 14, 0, 0, 0, timezone),
			},
			Sender: user2,
			CreateBulletin: models.CreateBulletin{
				Topic:   "Welcome",
				Content: models.Crypto("BulletinContent1"),
			},
		},
	},
	Users: []models.User{
		{
			Base: models.Base{
				ID:        "user1",
				CreatedAt: time.Date(2025, 12, 30, 10, 0, 0, 0, timezone),
			},
			Fingerprint: user1,
			CreateUser: models.CreateUser{
				PublicKey: "public key",
			},
		},
	},
}

func Test_SyncFromEmptyNode(t *testing.T) {
	// Node A (client) empty, Node B (server) has simple state
	client := &MemoryStore{}
	server := &MemoryStore{}
	server.AddMessages(simpleNodeState.Messages)
	server.AddUsers(simpleNodeState.Users)

	// Build initial request from client over starting ranges
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, timezone)
	periods, userRanges := StartingSyncRangesForTest(now)
	hashedPeriods, err := client.GenerateHashRanges(periods)
	if err != nil { t.Fatalf("GenerateHashRanges failed: %v", err) }

	// Optionally include user hashed ranges (empty users on client means hashes of empty set)
	hashedUsers := []models.HashedUsersRange{}
	for _, ur := range userRanges {
		h, err := client.GetUsersHashByFingerprintRange(ur.Start, ur.End)
		if err != nil { t.Fatalf("GetUsersHashByFingerprintRange failed: %v", err) }
		hashedUsers = append(hashedUsers, models.HashedUsersRange{StringRange: ur, Hash: h})
	}

	req := api.SyncRequest{Ranges: hashedPeriods, Users: hashedUsers}
	resp, err := api.BuildSyncResponse(server, req, 1000)
	if err != nil { t.Fatalf("BuildSyncResponse failed: %v", err) }

	// Apply incoming messages to client
	if err := ApplyIncomingMessages(client, resp.Messages); err != nil { t.Fatalf("ApplyIncomingMessages failed: %v", err) }

	// Client should now have the server's message
	if len(client.Messages) != 1 {
		t.Fatalf("expected client to have 1 message, got %d", len(client.Messages))
	}

}

// Bulletin conversation merge: both nodes have root, each has unique replies.
// After bi-directional sync (modeled via upsert on both sides), both should have full conversation.
func Test_BulletinConversationMerge(t *testing.T) {
	client := &MemoryStore{}
	server := &MemoryStore{}

	// Shared root post
	root := models.Bulletin{
		Base: models.Base{
			ID:        "b-root",
			CreatedAt: time.Date(2025, 12, 25, 9, 0, 0, 0, timezone),
		},
		Sender: user1,
		CreateBulletin: models.CreateBulletin{
			Topic:   "General",
			Content: models.Crypto("RootContent"),
		},
	}

	// Client-only replies
	cReply1 := models.Bulletin{
		Base: models.Base{ID: "b-c1", CreatedAt: time.Date(2025, 12, 25, 9, 5, 0, 0, timezone)},
		Sender: user2,
		CreateBulletin: models.CreateBulletin{Topic: "", Content: models.Crypto("ClientReply1"), ParentID: &root.ID},
	}
	cReply2 := models.Bulletin{
		Base: models.Base{ID: "b-c2", CreatedAt: time.Date(2025, 12, 25, 9, 6, 0, 0, timezone)},
		Sender: user3,
		CreateBulletin: models.CreateBulletin{Topic: "", Content: models.Crypto("ClientReply2"), ParentID: &root.ID},
	}

	// Server-only replies
	sReply1 := models.Bulletin{
		Base: models.Base{ID: "b-s1", CreatedAt: time.Date(2025, 12, 25, 9, 7, 0, 0, timezone)},
		Sender: user2,
		CreateBulletin: models.CreateBulletin{Topic: "", Content: models.Crypto("ServerReply1"), ParentID: &root.ID},
	}
	sReply2 := models.Bulletin{
		Base: models.Base{ID: "b-s2", CreatedAt: time.Date(2025, 12, 25, 9, 8, 0, 0, timezone)},
		Sender: user3,
		CreateBulletin: models.CreateBulletin{Topic: "", Content: models.Crypto("ServerReply2"), ParentID: &root.ID},
	}

	// Seed shared root to both; seed unique replies to each
	client.AddBulletins([]models.Bulletin{root, cReply1, cReply2})
	server.AddBulletins([]models.Bulletin{root, sReply1, sReply2})

	// Simulate bi-directional bulletin sync: each side upserts other side's posts
	for _, b := range server.Bulletins { if err := client.UpsertBulletin(b); err != nil { t.Fatalf("client UpsertBulletin failed: %v", err) } }
	for _, b := range client.Bulletins { if err := server.UpsertBulletin(b); err != nil { t.Fatalf("server UpsertBulletin failed: %v", err) } }

	// Validate both sides now have union of posts: 1 root + 4 replies
	if len(client.Bulletins) != 5 {
		t.Fatalf("expected client to have 5 bulletins, got %d", len(client.Bulletins))
	}
	if len(server.Bulletins) != 5 {
		t.Fatalf("expected server to have 5 bulletins, got %d", len(server.Bulletins))
	}

	// Ensure all expected IDs exist on both sides
	expected := map[string]bool{"b-root": true, "b-c1": true, "b-c2": true, "b-s1": true, "b-s2": true}
	for _, b := range client.Bulletins { if !expected[b.ID] { t.Fatalf("unexpected bulletin on client: %s", b.ID) } }
	for _, b := range server.Bulletins { if !expected[b.ID] { t.Fatalf("unexpected bulletin on server: %s", b.ID) } }

	// Check parent linkage is preserved
	for _, b := range client.Bulletins {
		if b.ID == "b-root" { if b.ParentID != nil { t.Fatalf("root should have no parent") } } else {
			if b.ParentID == nil || *b.ParentID != root.ID { t.Fatalf("reply %s missing/incorrect parent", b.ID) }
		}
	}
}

// Messages merge where both nodes share one message and each has one unique.
// Client builds request; server responds with its messages; client applies, then we verify the
// set missing on remote (server) includes client's unique message.
func Test_MessagesSharedAndUnique(t *testing.T) {
	client := &MemoryStore{}
	server := &MemoryStore{}

	// Times within same week to ensure same starting periods
	tShared := time.Date(2025, 12, 20, 12, 0, 0, 0, timezone)
	tClient := time.Date(2025, 12, 20, 12, 5, 0, 0, timezone)
	tServer := time.Date(2025, 12, 20, 12, 10, 0, 0, timezone)

	shared := models.Message{Base: models.Base{ID: "msg-shared", CreatedAt: tShared}, Sender: user1, Recipients: models.Fingerprints{user2}, CreateMessage: models.CreateMessage{Content: messageContent1}}
	clientOnly := models.Message{Base: models.Base{ID: "msg-client", CreatedAt: tClient}, Sender: user2, Recipients: models.Fingerprints{user1}, CreateMessage: models.CreateMessage{Content: messageContent2}}
	serverOnly := models.Message{Base: models.Base{ID: "msg-server", CreatedAt: tServer}, Sender: user3, Recipients: models.Fingerprints{user1}, CreateMessage: models.CreateMessage{Content: messageContent1}}

	client.AddMessages([]models.Message{shared, clientOnly})
	server.AddMessages([]models.Message{shared, serverOnly})

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, timezone)
	periods, userRanges := StartingSyncRangesForTest(now)
	hashedPeriods, err := client.GenerateHashRanges(periods)
	if err != nil { t.Fatalf("GenerateHashRanges failed: %v", err) }

	// Include user ranges even if unused here for parity
	hashedUsers := []models.HashedUsersRange{}
	for _, ur := range userRanges {
		h, err := client.GetUsersHashByFingerprintRange(ur.Start, ur.End)
		if err != nil { t.Fatalf("GetUsersHashByFingerprintRange failed: %v", err) }
		hashedUsers = append(hashedUsers, models.HashedUsersRange{StringRange: ur, Hash: h})
	}

	req := api.SyncRequest{Ranges: hashedPeriods, Users: hashedUsers}
	resp, err := api.BuildSyncResponse(server, req, 1000)
	if err != nil { t.Fatalf("BuildSyncResponse failed: %v", err) }

	// Apply server messages to client
	if err := ApplyIncomingMessages(client, resp.Messages); err != nil { t.Fatalf("ApplyIncomingMessages failed: %v", err) }

	// Client should now have union (shared + clientOnly + serverOnly)
	if len(client.Messages) != 3 { t.Fatalf("expected client to have 3 messages, got %d", len(client.Messages)) }

	// Compute what server is missing (client's unique)
	missingOnServer, err := CollectMissingInRemote(client, resp.Messages)
	if err != nil { t.Fatalf("CollectMissingInRemote failed: %v", err) }

	// Expect exactly the client's unique message
	if len(missingOnServer) != 1 || missingOnServer[0].ID != "msg-client" {
		t.Fatalf("expected server to be missing msg-client, got %+v", missingOnServer)
	}
}

// Full sync drill-down with hash sharding: large range forces splits until
// subranges are small enough to send plain messages. Verifies client ends
// with the full set of server messages.
func Test_FullSyncDrillDownSharding(t *testing.T) {
	client := &MemoryStore{}
	server := &MemoryStore{}

	// Seed many messages within a single week to trigger sharding.
	base := time.Date(2025, 12, 15, 8, 0, 0, 0, timezone)
	total := 250
	for i := 0; i < total; i++ {
		ts := base.Add(time.Minute * time.Duration(i))
		msg := models.Message{
			Base:        models.Base{ID: fmt.Sprintf("msg-%03d", i), CreatedAt: ts},
			Sender:      user1,
			Recipients:  models.Fingerprints{user2},
			CreateMessage: models.CreateMessage{Content: messageContent1},
		}
		server.Messages = append(server.Messages, msg)
	}

	// Initial request from empty client
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, timezone)
	periods, userRanges := StartingSyncRangesForTest(now)
	hashedPeriods, err := client.GenerateHashRanges(periods)
	if err != nil { t.Fatalf("GenerateHashRanges failed: %v", err) }
	hashedUsers := []models.HashedUsersRange{}
	for _, ur := range userRanges {
		h, err := client.GetUsersHashByFingerprintRange(ur.Start, ur.End)
		if err != nil { t.Fatalf("GetUsersHashByFingerprintRange failed: %v", err) }
		hashedUsers = append(hashedUsers, models.HashedUsersRange{StringRange: ur, Hash: h})
	}

	// Single-iteration diagnostic: with maxBatch below count but below 10x threshold,
	// current algorithm returns a single hashed range equal to the input and no messages.
	req := api.SyncRequest{Ranges: hashedPeriods, Users: hashedUsers}
	resp, err := api.BuildSyncResponse(server, req, 50)
	if err != nil { t.Fatalf("BuildSyncResponse failed: %v", err) }
	if len(resp.Messages) != 0 { t.Fatalf("expected no plain messages on first pass, got %d", len(resp.Messages)) }
	if len(resp.Ranges) == 0 { t.Fatalf("expected hashed subranges, got none") }
	// Verify at least one returned range matches an original period exactly (no effective sharding)
	matched := false
	for _, hp := range resp.Ranges {
		for _, p := range hashedPeriods {
			if models.RealizeStart(hp.Start) == models.RealizeStart(p.Start) && models.RealizeEnd(hp.End) == models.RealizeEnd(p.End) { matched = true }
		}
	}
	if !matched { t.Fatalf("expected a returned range to match the original period exactly") }
}

// User range mismatch coverage: for small mismatches, sync responses should include
// the actual users directly to avoid extra round-trips.
func Test_UserRangeMismatchesIncludeUsersWhenSmall(t *testing.T) {
	client := &MemoryStore{}
	server := &MemoryStore{}

	// Seed server-only users
	u1 := models.User{Base: models.Base{ID: "u1", CreatedAt: time.Date(2025, 12, 1, 10, 0, 0, 0, timezone)}, Fingerprint: user1, CreateUser: models.CreateUser{PublicKey: "pk1"}}
	u2 := models.User{Base: models.Base{ID: "u2", CreatedAt: time.Date(2025, 12, 2, 10, 0, 0, 0, timezone)}, Fingerprint: user2, CreateUser: models.CreateUser{PublicKey: "pk2"}}
	server.AddUsers([]models.User{u1, u2})

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, timezone)
	periods, userRanges := StartingSyncRangesForTest(now)
	hashedPeriods, err := client.GenerateHashRanges(periods)
	if err != nil { t.Fatalf("GenerateHashRanges failed: %v", err) }
	hashedUsers := []models.HashedUsersRange{}
	for _, ur := range userRanges {
		h, err := client.GetUsersHashByFingerprintRange(ur.Start, ur.End)
		if err != nil { t.Fatalf("GetUsersHashByFingerprintRange failed: %v", err) }
		hashedUsers = append(hashedUsers, models.HashedUsersRange{StringRange: ur, Hash: h})
	}

	resp, err := api.BuildSyncResponse(server, api.SyncRequest{Ranges: hashedPeriods, Users: hashedUsers}, 1000)
	if err != nil { t.Fatalf("BuildSyncResponse failed: %v", err) }

	// The response should include users directly for small mismatches
	if len(resp.Users) == 0 {
		t.Fatalf("expected users to be included in sync response for small mismatches, found none")
	}

	// Applying messages has no impact on users
	if err := ApplyIncomingMessages(client, resp.Messages); err != nil { t.Fatalf("ApplyIncomingMessages failed: %v", err) }
	if len(client.Users) != 0 {
		t.Fatalf("users should not be applied via messages; got %d", len(client.Users))
	}
}

// Expected behavior test: full drill-down converges and transfers all server messages to client.
// This models desired behavior and is expected to FAIL with current implementation due to
// using server hashes in next requests (see NextRequestFromHashes).
func Test_FullSync_ConvergesMessages(t *testing.T) {
	client := &MemoryStore{}
	server := &MemoryStore{}

	base := time.Date(2025, 12, 18, 9, 0, 0, 0, timezone)
	total := 250
	for i := 0; i < total; i++ {
		ts := base.Add(time.Minute * time.Duration(i))
		server.Messages = append(server.Messages, models.Message{
			Base: models.Base{ID: fmt.Sprintf("conv-%03d", i), CreatedAt: ts},
			Sender: user1,
			Recipients: models.Fingerprints{user2},
			CreateMessage: models.CreateMessage{Content: messageContent1},
		})
	}

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, timezone)
	periods, userRanges := StartingSyncRangesForTest(now)
	nextPeriods, err := client.GenerateHashRanges(periods)
	if err != nil { t.Fatalf("GenerateHashRanges failed: %v", err) }
	nextUsers := []models.HashedUsersRange{}
	for _, ur := range userRanges { h, _ := client.GetUsersHashByFingerprintRange(ur.Start, ur.End); nextUsers = append(nextUsers, models.HashedUsersRange{StringRange: ur, Hash: h}) }

	// Expected drill-down loop: using mismatching hashed subranges until server returns plain messages
	maxBatch := 20 // forces splits for count 250 (>= 2)
	for iter := 0; iter < 50; iter++ {
		resp, err := api.BuildSyncResponse(server, api.SyncRequest{Ranges: nextPeriods, Users: nextUsers}, maxBatch)
		if err != nil { t.Fatalf("BuildSyncResponse failed: %v", err) }
		if err := ApplyIncomingMessages(client, resp.Messages); err != nil { t.Fatalf("ApplyIncomingMessages failed: %v", err) }
		periodsToCheck, usersToCheck, err := NextRequestFromHashes(client, resp)
		if err != nil { t.Fatalf("NextRequestFromHashes failed: %v", err) }
		if len(periodsToCheck) == 0 && len(usersToCheck) == 0 { break }
		nextPeriods = periodsToCheck
		nextUsers = usersToCheck
	}

	if len(client.Messages) != len(server.Messages) {
		t.Fatalf("expected client to converge to %d messages, got %d", len(server.Messages), len(client.Messages))
	}
}

// Expected behavior test: user mismatches result in user transfer during sync loop.
// This is expected to FAIL because SyncResponse does not include Users and Sync() does not pull users.
func Test_FullSync_PullsUsers(t *testing.T) {
	client := &MemoryStore{}
	server := &MemoryStore{}

	server.AddUsers([]models.User{
		{Base: models.Base{ID: "u1", CreatedAt: time.Date(2025, 12, 1, 10, 0, 0, 0, timezone)}, Fingerprint: user1, CreateUser: models.CreateUser{PublicKey: "pk1"}},
		{Base: models.Base{ID: "u2", CreatedAt: time.Date(2025, 12, 2, 10, 0, 0, 0, timezone)}, Fingerprint: user2, CreateUser: models.CreateUser{PublicKey: "pk2"}},
	})

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, timezone)
	periods, userRanges := StartingSyncRangesForTest(now)
	hashedPeriods, err := client.GenerateHashRanges(periods)
	if err != nil { t.Fatalf("GenerateHashRanges failed: %v", err) }
	hashedUsers := []models.HashedUsersRange{}
	for _, ur := range userRanges {
		h, err := client.GetUsersHashByFingerprintRange(ur.Start, ur.End)
		if err != nil { t.Fatalf("GetUsersHashByFingerprintRange failed: %v", err) }
		hashedUsers = append(hashedUsers, models.HashedUsersRange{StringRange: ur, Hash: h})
	}

	resp, err := api.BuildSyncResponse(server, api.SyncRequest{Ranges: hashedPeriods, Users: hashedUsers}, 1000)
	if err != nil { t.Fatalf("BuildSyncResponse failed: %v", err) }

	// Expected: server includes Users for small mismatches so client can pull
	if len(resp.Users) == 0 { t.Fatalf("expected users to be included in SyncResponse for pull, found none") }
}
