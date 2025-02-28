package lib

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/google/uuid"
)

func GenerateShortURL() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func GenerateUUID() string {
	return uuid.New().String()
}
