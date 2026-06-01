package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"axial/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newTestServer spins up an in-memory SQLite DB and binds an api.Server to
// it. The smoke test below proves the DB seam — handlers can be exercised
// without Postgres, network sockets, or a multicast listener.
func newTestServer(t *testing.T) *Server {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite memory DB: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Message{}, &models.Bulletin{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return NewServer(db)
}

func TestServer_HandleGetUsers_EmptyDB(t *testing.T) {
	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/users", nil)
	rr := httptest.NewRecorder()
	s.handleGetUsers(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%q", rr.Code, rr.Body.String())
	}
	var got []models.User
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode body: %v (body=%q)", err, rr.Body.String())
	}
	if len(got) != 0 {
		t.Errorf("empty DB should return zero users, got %d", len(got))
	}
}

func TestServer_HandlePing_RefreshesHashesOnEmptyDB(t *testing.T) {
	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/ping", nil)
	rr := httptest.NewRecorder()
	s.handlePing(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%q", rr.Code, rr.Body.String())
	}
	var got PingResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode body: %v (body=%q)", err, rr.Body.String())
	}
	if got.Hashes.Full == "" {
		t.Error("expected non-empty Full hash even for empty DB (sha256 of empty input)")
	}
}
