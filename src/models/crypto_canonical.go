package models

// IsCanonicalFingerprint validates the canonical encryption KeyID format.
// Canonical format is 16-hex characters, lowercase, no whitespace.
func IsCanonicalFingerprint(s string) bool {
    if len(s) != 16 { return false }
    for _, c := range s {
        switch {
        case c >= '0' && c <= '9':
        case c >= 'a' && c <= 'f':
        default:
            return false
        }
    }
    return true
}
