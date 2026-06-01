package discovery

import (
	"net"
	"strings"
)

// Announcement is a parsed discovery packet broadcast by another Node:
// "NodeID|FullHash|APIAddr|LocalIP".
type Announcement struct {
	NodeID  string
	Hash    string
	APIAddr string
	LocalIP string
	Src     *net.UDPAddr
}

// ParseAnnouncement decodes a raw multicast packet. Returns (ann, true) when
// the packet matches the four-field axial format, (zero, false) otherwise.
// The function does no I/O and depends on nothing but its inputs, so the
// parsing logic can be unit-tested without sockets.
func ParseAnnouncement(packet string, src *net.UDPAddr) (Announcement, bool) {
	parts := strings.Split(packet, "|")
	if len(parts) != 4 {
		return Announcement{}, false
	}
	return Announcement{
		NodeID:  parts[0],
		Hash:    parts[1],
		APIAddr: parts[2],
		LocalIP: parts[3],
		Src:     src,
	}, true
}
