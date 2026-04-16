package crypto

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateState() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}
