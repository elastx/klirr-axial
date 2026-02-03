package models

import (
	"database/sql/driver"
	"fmt"
)

type Fingerprint string

// Implement driver.Valuer so GORM knows how to store Fingerprint
func (f Fingerprint) Value() (driver.Value, error) {
	return string(f), nil
}

// Implement sql.Scanner so GORM knows how to read Fingerprint
func (f *Fingerprint) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		*f = Fingerprint(v)
		return nil
	case []byte:
		*f = Fingerprint(string(v))
		return nil
	case nil:
		*f = ""
		return nil
	default:
		return fmt.Errorf("unsupported Scan type for Fingerprint: %T", value)
	}
}
