package synchronization

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"time"

	"axial/models"
)

// MemoryStore is a simple in-memory implementation of api.SyncStore suitable for tests.
// It is deterministic and ignores files/bulletins for hashing at the moment.
type MemoryStore struct {
	Messages []models.Message
	Users    []models.User
	Bulletins []models.Bulletin
}

func (m *MemoryStore) GenerateHashRanges(periods []models.Period) ([]models.HashedPeriod, error) {
	out := make([]models.HashedPeriod, 0, len(periods))
	for _, p := range periods {
		h := m.hashMessagesInPeriod(p)
		out = append(out, models.HashedPeriod{Period: p, Hash: h})
	}
	return out, nil
}

func (m *MemoryStore) CountMessagesByPeriod(period models.Period) int64 {
	msgs := m.messagesInPeriod(period)
	return int64(len(msgs))
}

func (m *MemoryStore) GetMessagesByPeriod(period models.Period) ([]models.Message, error) {
	return m.messagesInPeriod(period), nil
}

func (m *MemoryStore) GetDatabaseHashes() (models.HashSet, error) {
	// messages hash
	hm := sha256.New()
	ids := make([]string, 0, len(m.Messages))
	for _, msg := range m.Messages {
		ids = append(ids, msg.ID)
	}
	sort.Strings(ids)
	for _, id := range ids {
		hm.Write([]byte(id))
	}
	messagesHash := hex.EncodeToString(hm.Sum(nil))

	// users hash
	hu := sha256.New()
	fps := make([]string, 0, len(m.Users))
	for _, u := range m.Users {
		fps = append(fps, string(u.Fingerprint))
	}
	sort.Strings(fps)
	for _, fp := range fps {
		hu.Write([]byte(fp))
	}
	usersHash := hex.EncodeToString(hu.Sum(nil))

	// Combine
	full := sha256.New()
	full.Write([]byte("messages:" + messagesHash))
	full.Write([]byte("users:" + usersHash))
	full.Write([]byte("files:"))
	return models.HashSet{Messages: messagesHash, Users: usersHash, Files: "", Full: hex.EncodeToString(full.Sum(nil))}, nil
}

func (m *MemoryStore) GetUsersHashByFingerprintRange(start, end string) (string, error) {
	hu := sha256.New()
	fps := []string{}
	for _, u := range m.Users {
		fp := string(u.Fingerprint)
		if fp >= start && fp <= end {
			fps = append(fps, fp)
		}
	}
	sort.Strings(fps)
	for _, fp := range fps {
		hu.Write([]byte(fp))
	}
	return hex.EncodeToString(hu.Sum(nil)), nil
}

func (m *MemoryStore) UpsertMessage(msg models.Message) error {
	for _, existing := range m.Messages {
		if existing.ID == msg.ID {
			return nil
		}
	}
	m.Messages = append(m.Messages, msg)
	return nil
}

func (m *MemoryStore) UpsertBulletin(b models.Bulletin) error {
	for _, existing := range m.Bulletins {
		if existing.ID == b.ID {
			return nil
		}
	}
	m.Bulletins = append(m.Bulletins, b)
	return nil
}

// Helpers
func (m *MemoryStore) messagesInPeriod(p models.Period) []models.Message {
	start := models.RealizeStart(p.Start)
	end := models.RealizeEnd(p.End)
	out := []models.Message{}
	for _, msg := range m.Messages {
		if !msg.CreatedAt.Before(start) && msg.CreatedAt.Before(end) {
			out = append(out, msg)
		}
	}
	return out
}

func (m *MemoryStore) hashMessagesInPeriod(p models.Period) string {
	msgs := m.messagesInPeriod(p)
	ids := make([]string, 0, len(msgs))
	for _, msg := range msgs {
		ids = append(ids, msg.ID)
	}
	sort.Strings(ids)
	h := sha256.New()
	for _, id := range ids {
		h.Write([]byte(id))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Seed helpers for tests
func (m *MemoryStore) AddMessages(msgs []models.Message) { m.Messages = append(m.Messages, msgs...) }
func (m *MemoryStore) AddUsers(users []models.User)     { m.Users = append(m.Users, users...) }
func (m *MemoryStore) AddBulletins(b []models.Bulletin) { m.Bulletins = append(m.Bulletins, b...) }

// Utility to construct ranges like production
func StartingSyncRangesForTest(now time.Time) ([]models.Period, []models.StringRange) {
	// Mirror startingSyncRanges(DefaultClock) logic but deterministic using provided now
	earliestStartTime := models.RealizeStart(nil)
	latestEndTime := models.RealizeEnd(nil)

	periodSteps := []struct{ Years, Months, Days int }{{0, -1, 0}, {0, -6, 0}, {-2, 0, 0}}
	weekStart := now.AddDate(0, 0, -int((int(now.Weekday()+6)%7))) // Monday start
	weekStart = weekStart.Truncate(24 * time.Hour)
	var previousStart *time.Time
	if weekStart.Before(earliestStartTime) {
		previousStart = &earliestStartTime
	} else {
		previousStart = &weekStart
	}

	periods := []models.Period{{Start: previousStart, End: &latestEndTime}}
	for _, step := range periodSteps {
		start := previousStart.AddDate(step.Years, step.Months, step.Days)
		if start.Before(earliestStartTime) {
			periods = append(periods, models.Period{Start: &earliestStartTime, End: previousStart})
			break
		}
		periods = append(periods, models.Period{Start: &start, End: previousStart})
		previousStart = &start
	}
	periods = append(periods, models.Period{Start: &earliestStartTime, End: previousStart})

	var userRanges []models.StringRange
	for i := 0; i < 10; i++ { userRanges = append(userRanges, models.StringRange{Start: string(rune('0'+i)), End: string(rune('0'+i+1))}) }
	for i := 0; i < 25; i++ { userRanges = append(userRanges, models.StringRange{Start: string(rune('a'+i)), End: string(rune('a'+i+1))}) }
	return periods, userRanges
}
