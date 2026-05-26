package discovery

import (
	"net"
	"testing"
)

func TestParseAnnouncement_HappyPath(t *testing.T) {
	src := &net.UDPAddr{IP: net.ParseIP("192.168.1.42"), Port: 45678}
	ann, ok := ParseAnnouncement("node-a|abc123|:8080|192.168.1.99", src)
	if !ok {
		t.Fatal("expected parse to succeed for a 4-field axial packet")
	}
	if ann.NodeID != "node-a" {
		t.Errorf("NodeID: got %q want %q", ann.NodeID, "node-a")
	}
	if ann.Hash != "abc123" {
		t.Errorf("Hash: got %q want %q", ann.Hash, "abc123")
	}
	if ann.APIAddr != ":8080" {
		t.Errorf("APIAddr: got %q want %q", ann.APIAddr, ":8080")
	}
	if ann.LocalIP != "192.168.1.99" {
		t.Errorf("LocalIP: got %q want %q", ann.LocalIP, "192.168.1.99")
	}
	if ann.Src != src {
		t.Errorf("Src: got %v want %v", ann.Src, src)
	}
}

func TestParseAnnouncement_RejectsWrongFieldCount(t *testing.T) {
	cases := []string{
		"",
		"only-one",
		"a|b",
		"a|b|c",
		"a|b|c|d|e",
	}
	for _, c := range cases {
		if _, ok := ParseAnnouncement(c, nil); ok {
			t.Errorf("ParseAnnouncement(%q) should have returned ok=false", c)
		}
	}
}

func TestParseAnnouncement_PreservesEmptyFields(t *testing.T) {
	// If a future bug ships a packet with empty fields, the parser should
	// still recognise its shape and let the handler decide what to do with
	// the empties — silently dropping is worse.
	ann, ok := ParseAnnouncement("||abc|", nil)
	if !ok {
		t.Fatal("expected parse to succeed for 4-field packet with empty fields")
	}
	if ann.NodeID != "" || ann.Hash != "" || ann.APIAddr != "abc" || ann.LocalIP != "" {
		t.Errorf("unexpected fields: %+v", ann)
	}
}
