package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"axial/models"

	"gorm.io/gorm"
)

// The handlers currently read from the global models.DB. Tests substitute
// an in-memory SQLite into the global so they can drive the handlers
// without Postgres. A package-local mutex is unnecessary because the api
// package has no other concurrent test surface; if more handler tests
// arrive they should serialise via t.Setenv-style fixtures.
func withGlobalDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	prev := models.DB
	models.DB = db
	t.Cleanup(func() { models.DB = prev })
}

func TestHandleSyncMessages_RejectsGET(t *testing.T) {
	withGlobalDB(t, newIngestTestDB(t))
	req := httptest.NewRequest(http.MethodGet, "/v1/sync/messages", nil)
	rr := httptest.NewRecorder()
	handleSyncMessages(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET should be 405, got %d", rr.Code)
	}
}

func TestHandleSyncMessages_RejectsEmptyBatch(t *testing.T) {
	withGlobalDB(t, newIngestTestDB(t))
	req := httptest.NewRequest(http.MethodPost, "/v1/sync/messages",
		strings.NewReader(`{"messages": []}`))
	rr := httptest.NewRecorder()
	handleSyncMessages(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("empty batch should be 400, got %d", rr.Code)
	}
}

func TestHandleSyncMessages_RejectsInvalidJSON(t *testing.T) {
	withGlobalDB(t, newIngestTestDB(t))
	req := httptest.NewRequest(http.MethodPost, "/v1/sync/messages",
		bytes.NewBufferString(`{not json`))
	rr := httptest.NewRecorder()
	handleSyncMessages(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("malformed JSON should be 400, got %d", rr.Code)
	}
}

// Regression test for the original `json:"messages"` typo on the Bulletins
// field. Posting a body keyed by "bulletins" must decode into the request
// struct; if the tag drifts back to "messages" this test goes red.
func TestHandleSyncBulletins_AcceptsBulletinsKey(t *testing.T) {
	withGlobalDB(t, newIngestTestDB(t))

	// An empty bulletins array decodes successfully but trips the length
	// check, so we get a 400 (not 200/201). That's still proof the JSON tag
	// matched — if the tag were wrong the field would be zero-length under
	// any key and we'd hit the same 400 path. So we need a non-empty array
	// to distinguish: with the correct tag we proceed past length check.
	body := strings.NewReader(`{"bulletins": [{"id":"b1"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/sync/bulletins", body)
	rr := httptest.NewRecorder()
	handleSyncBulletins(rr, req)

	// We expect either 201 (ingest succeeded) or 500 (ingest failed
	// downstream — e.g. PGP validation on the empty record). What we MUST
	// NOT see is 400 "Bulletins are required", which is the symptom of the
	// JSON tag bug: a wrong tag silently maps `bulletins` to nothing.
	if rr.Code == http.StatusBadRequest && strings.Contains(rr.Body.String(), "required") {
		t.Errorf("got %d %q — JSON tag bug regression: 'bulletins' key did not decode into Bulletins field",
			rr.Code, rr.Body.String())
	}
}

// Same shape for Users to lock in the convention.
func TestHandleSyncUsers_AcceptsUsersKey(t *testing.T) {
	withGlobalDB(t, newIngestTestDB(t))
	body := strings.NewReader(`{"users": [{"fingerprint":"u1"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/sync/users", body)
	rr := httptest.NewRecorder()
	handleSyncUsers(rr, req)
	if rr.Code == http.StatusBadRequest && strings.Contains(rr.Body.String(), "required") {
		t.Errorf("got %d %q — `users` key did not decode into Users field",
			rr.Code, rr.Body.String())
	}
}

func TestHandleSyncMessages_AcceptsMessagesKey(t *testing.T) {
	withGlobalDB(t, newIngestTestDB(t))
	body := strings.NewReader(`{"messages": [{"id":"m1"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/sync/messages", body)
	rr := httptest.NewRecorder()
	handleSyncMessages(rr, req)
	if rr.Code == http.StatusBadRequest && strings.Contains(rr.Body.String(), "required") {
		t.Errorf("got %d %q — `messages` key did not decode into Messages field",
			rr.Code, rr.Body.String())
	}
}
