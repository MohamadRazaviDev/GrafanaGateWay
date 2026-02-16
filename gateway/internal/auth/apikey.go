package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/MohamadRazaviDev/Grafana-Gateway/gateway/internal/config"
)

// Identity represents an authenticated user.
type Identity struct {
	User  string
	Team  string
	Roles []string
}

// APIKeyValidator validates API keys against configured hashes.
type APIKeyValidator struct {
	keys []config.APIKey
}

// NewAPIKeyValidator creates a validator from config.
func NewAPIKeyValidator(keys []config.APIKey) *APIKeyValidator {
	return &APIKeyValidator{keys: keys}
}

// Validate checks a raw API key against stored SHA-256 hashes.
// Uses constant-time comparison to prevent timing attacks.
func (v *APIKeyValidator) Validate(rawKey string) (*Identity, error) {
	rawKey = strings.TrimSpace(rawKey)
	if rawKey == "" {
		return nil, fmt.Errorf("empty API key")
	}

	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	for _, k := range v.keys {
		storedHash := strings.ToLower(strings.TrimSpace(k.Hash))
		if subtle.ConstantTimeCompare([]byte(keyHash), []byte(storedHash)) == 1 {
			return &Identity{
				User:  k.User,
				Team:  k.Team,
				Roles: k.Roles,
			}, nil
		}
	}

	return nil, fmt.Errorf("invalid API key")
}

// HashKey computes the SHA-256 hash of a raw API key.
// Use this to generate hashes for config files.
func HashKey(rawKey string) string {
	hash := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(hash[:])
}
