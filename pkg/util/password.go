package util

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

var (
	argonTime    uint32 = 1
	argonMemory  uint32 = 64 * 1024
	argonThreads uint8  = 2
	argonKeyLen  uint32 = 32
	saltLen      uint32 = 16
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

func VerifyPassword(password, encodedHash string) (bool, error) {
	parts := bytes.Split([]byte(encodedHash), []byte("$"))
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid encoded hash format")
	}

	var memory, time uint32
	var threads uint8
	if _, err := fmt.Sscanf(string(parts[3]), "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return false, fmt.Errorf("invalid argon2 parameters: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(string(parts[4]))
	if err != nil {
		return false, fmt.Errorf("invalid salt base64: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(string(parts[5]))
	if err != nil {
		return false, fmt.Errorf("invalid hash base64: %w", err)
	}

	computedHash := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(expectedHash)))

	return bytes.Equal(computedHash, expectedHash), nil
}
