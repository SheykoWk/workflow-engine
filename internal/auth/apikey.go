package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	apiKeyPrefix   = "wf_"
	bcryptCost     = bcrypt.DefaultCost
	prefixRandLen  = 9  // base64url → 12 chars
	secretRandLen  = 32 // base64url → 43 chars
)

// GenerateAPIKey returns the full key (wf_<prefix>.<secret>), the DB lookup prefix
// (without wf_), and an error. Uses crypto/rand and base64url without padding.
func GenerateAPIKey() (fullKey string, dbPrefix string, err error) {
	p := make([]byte, prefixRandLen)
	if _, err := rand.Read(p); err != nil {
		return "", "", fmt.Errorf("generate prefix: %w", err)
	}
	s := make([]byte, secretRandLen)
	if _, err := rand.Read(s); err != nil {
		return "", "", fmt.Errorf("generate secret: %w", err)
	}
	dbPrefix = base64.RawURLEncoding.EncodeToString(p)
	secret := base64.RawURLEncoding.EncodeToString(s)
	fullKey = apiKeyPrefix + dbPrefix + "." + secret
	return fullKey, dbPrefix, nil
}

// HashAPIKey bcrypt-hashes the complete raw API key (what the client sends in Authorization).
func HashAPIKey(rawKey string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(rawKey), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash api key: %w", err)
	}
	return string(b), nil
}

// CompareAPIKey checks rawKey against a bcrypt hash.
func CompareAPIKey(keyHash, rawKey string) bool {
	return bcrypt.CompareHashAndPassword([]byte(keyHash), []byte(rawKey)) == nil
}

// ParseBearerAPIKey extracts the raw token from "Bearer <token>" or a bare "wf_..." key.
// Swagger UI sends apiKey headers without the Bearer prefix, so both formats are accepted.
func ParseBearerAPIKey(authHeader string) (token string, ok bool) {
	const bearer = "Bearer "
	if strings.EqualFold(authHeader[:min(len(authHeader), len(bearer))], bearer) {
		t := strings.TrimSpace(authHeader[len(bearer):])
		if t == "" {
			return "", false
		}
		return t, true
	}
	if strings.HasPrefix(authHeader, apiKeyPrefix) {
		return authHeader, true
	}
	return "", false
}

// SplitWFAPIKey parses wf_<prefix>.<secret> and returns prefix (for DB lookup) and full raw key for bcrypt.
func SplitWFAPIKey(raw string) (dbPrefix, fullKey string, ok bool) {
	if !strings.HasPrefix(raw, apiKeyPrefix) {
		return "", "", false
	}
	rest := raw[len(apiKeyPrefix):]
	dot := strings.IndexByte(rest, '.')
	if dot <= 0 || dot == len(rest)-1 {
		return "", "", false
	}
	dbPrefix = rest[:dot]
	if len(dbPrefix) < 8 || len(dbPrefix) > 64 { // sanity bounds
		return "", "", false
	}
	return dbPrefix, raw, true
}
