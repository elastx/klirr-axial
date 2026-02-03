package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Fingerprints []Fingerprint

// Value stores the slice as JSON in the DB
func (fs Fingerprints) Value() (driver.Value, error) {
	// convert to []string for clean JSON
	arr := make([]string, len(fs))
	for i, f := range fs {
		arr[i] = string(f)
	}
	b, err := json.Marshal(arr)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan loads the slice from JSON stored in the DB
func (fs *Fingerprints) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		return fs.fromJSON([]byte(v))
	case []byte:
		return fs.fromJSON(v)
	case nil:
		*fs = nil
		return nil
	default:
		return fmt.Errorf("unsupported Scan type for Fingerprints: %T", value)
	}
}

func (fs *Fingerprints) fromJSON(b []byte) error {
	var arr []string
	if len(b) == 0 {
		*fs = nil
		return nil
	}
	if err := json.Unmarshal(b, &arr); err != nil {
		return err
	}
	out := make([]Fingerprint, len(arr))
	for i, s := range arr {
		out[i] = Fingerprint(s)
	}
	*fs = out
	return nil
}
