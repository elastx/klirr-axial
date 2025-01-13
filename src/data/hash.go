package data

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

func CalculateHash(data []DataBlock) string {
	hasher := sha256.New()
	for _, block := range data {
		jsonBlock, _ := json.Marshal(block)
		hasher.Write(jsonBlock)
	}
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
