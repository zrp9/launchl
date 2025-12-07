// Package hash has util functions for hashing
package hash

import (
	"crypto"
	"encoding/hex"
)

func GenerateHashLink(key string) string {
	hasher := crypto.SHA512_256.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}
