package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// argon2id parameters — OWASP 2024 recommended minimum.
const (
	argonTime    uint32 = 2
	argonMemory  uint32 = 64 * 1024 // 64 MB
	argonThreads uint8  = 1
	argonKeyLen  uint32 = 32
	argonSaltLen        = 16
)

// ErrInvalidHashFormat is returned when a stored hash cannot be parsed.
var ErrInvalidHashFormat = errors.New("auth: invalid password hash format")

// HashPassword derives an argon2id hash from plain and returns the encoded string.
// Format: $argon2id$v=19$m=<m>,t=<t>,p=<p>$<base64-salt>$<base64-hash>
func HashPassword(plain string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("auth: generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(plain), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argonMemory,
		argonTime,
		argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
	return encoded, nil
}

// VerifyPassword reports whether plain matches the stored argon2id hash.
// Uses constant-time comparison to prevent timing attacks.
func VerifyPassword(hash, plain string) (bool, error) {
	salt, storedHash, err := decodeHash(hash)
	if err != nil {
		return false, err
	}

	// Re-derive with same params (parsed from the encoded string).
	var m, t uint32
	var p uint8
	// We always store our fixed params; parse for future flexibility.
	_, err = fmt.Sscanf(
		extractParams(hash),
		"m=%d,t=%d,p=%d",
		&m, &t, &p,
	)
	if err != nil {
		return false, fmt.Errorf("auth: parse argon params: %w", err)
	}

	candidate := argon2.IDKey([]byte(plain), salt, t, m, p, uint32(len(storedHash)))
	if subtle.ConstantTimeCompare(candidate, storedHash) != 1 {
		return false, nil
	}
	return true, nil
}

// decodeHash parses the $argon2id$... encoded string and returns (salt, hash, err).
func decodeHash(encoded string) ([]byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	// parts: ["", "argon2id", "v=19", "m=...,t=...,p=...", "<salt>", "<hash>"]
	if len(parts) != 6 || parts[1] != "argon2id" {
		return nil, nil, ErrInvalidHashFormat
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, fmt.Errorf("auth: decode salt: %w", err)
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, fmt.Errorf("auth: decode hash: %w", err)
	}

	return salt, hash, nil
}

// extractParams returns the "m=...,t=...,p=..." segment from the encoded string.
func extractParams(encoded string) string {
	parts := strings.Split(encoded, "$")
	if len(parts) < 5 {
		return ""
	}
	return parts[3]
}
