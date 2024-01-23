package helpers

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashBytes(input []byte) string {
	sum := sha256.Sum256(input)
	return hex.EncodeToString(sum[:])
}
