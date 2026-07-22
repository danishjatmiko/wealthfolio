package service

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters per OWASP's Password Storage Cheat Sheet's first
// recommended configuration (m=19 MiB, t=2, p=1) — the current best-practice
// default for password hashing, ahead of bcrypt/PBKDF2 in OWASP's ordering
// because it's tunable for both memory and compute cost, resisting GPU/ASIC
// cracking far better than either.
const (
	argon2Time    = 2
	argon2MemoryK = 19 * 1024 // KiB
	argon2Threads = 1
	argon2KeyLen  = 32
	argon2SaltLen = 16
)

// ErrInvalidHashFormat is returned by VerifyPassword when the stored hash
// isn't a well-formed PHC-style Argon2id string.
var ErrInvalidHashFormat = errors.New("invalid password hash format")

// HashPassword returns a PHC-formatted Argon2id hash string
// ($argon2id$v=19$m=...,t=...,p=...$salt$hash) safe to store directly in
// the database — the salt and parameters travel with the hash, so nothing
// else needs to be recorded to verify it later.
func HashPassword(password string) (string, error) {
	salt := make([]byte, argon2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, argon2Time, argon2MemoryK, argon2Threads, argon2KeyLen)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argon2MemoryK, argon2Time, argon2Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

// VerifyPassword checks a plaintext password against a PHC-formatted
// Argon2id hash produced by HashPassword, re-deriving the hash with the
// embedded salt/parameters and comparing in constant time so timing can't
// leak how close a guess was.
func VerifyPassword(password, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, ErrInvalidHashFormat
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, ErrInvalidHashFormat
	}

	var memoryK, timeCost uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memoryK, &timeCost, &threads); err != nil {
		return false, ErrInvalidHashFormat
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, ErrInvalidHashFormat
	}
	wantHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, ErrInvalidHashFormat
	}

	gotHash := argon2.IDKey([]byte(password), salt, timeCost, memoryK, threads, uint32(len(wantHash)))
	return subtle.ConstantTimeCompare(gotHash, wantHash) == 1, nil
}
