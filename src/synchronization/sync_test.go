package synchronization

import (
	"axial/api"
	"axial/models"
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
