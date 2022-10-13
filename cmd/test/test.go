package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/rand"
	"time"
)

// Test out Tesla OAuth
func main() {
	codeVerifyier := String(86)
	fmt.Printf("code_verifier: %v\n", codeVerifyier)
	h := sha256.New()
	h.Write([]byte(codeVerifyier))
	b := h.Sum(nil)
	fmt.Printf("hash: %v\n", base64.StdEncoding.EncodeToString(b))

}

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func String(length int) string {
	return StringWithCharset(length, charset)
}

// NewSHA256 ...
func NewSHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}
