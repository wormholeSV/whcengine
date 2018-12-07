package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
)

func CyptoSha256(key string) string {
	h := sha256.New()
	h.Write([]byte(key))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func GenSalt(size int) string {
	rand.Seed(time.Now().UnixNano())
	data := make([]byte, size)
	rand.Read(data)
	return hex.EncodeToString(data)
}
