package auth

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32) // 256-bit
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
