package auth

import (
	"testing"

	"github.com/MohamadRazaviDev/GrafanaGateWay/gateway/internal/config"
)

func TestHashKey(t *testing.T) {
	hash := HashKey("my-secret-key")
	if hash == "" {
		t.Error("expected non-empty hash")
	}
	if hash == "my-secret-key" {
		t.Error("hash should not equal raw key")
	}
	// Same input → same output
	if HashKey("my-secret-key") != hash {
		t.Error("expected deterministic hash")
	}
}

func TestValidate(t *testing.T) {
	rawKey := "test-api-key-12345"
	hash := HashKey(rawKey)

	validator := NewAPIKeyValidator([]config.APIKey{
		{Hash: hash, User: "alice", Team: "sre", Roles: []string{"viewer"}},
	})

	identity, err := validator.Validate(rawKey)
	if err != nil {
		t.Fatal(err)
	}
	if identity.User != "alice" {
		t.Errorf("expected alice, got %s", identity.User)
	}
	if identity.Team != "sre" {
		t.Errorf("expected sre, got %s", identity.Team)
	}
}

func TestValidateInvalidKey(t *testing.T) {
	validator := NewAPIKeyValidator([]config.APIKey{
		{Hash: HashKey("real-key"), User: "alice", Team: "sre"},
	})

	_, err := validator.Validate("wrong-key")
	if err == nil {
		t.Error("expected error for invalid key")
	}
}

func TestValidateEmptyKey(t *testing.T) {
	validator := NewAPIKeyValidator([]config.APIKey{
		{Hash: HashKey("real-key"), User: "alice", Team: "sre"},
	})

	_, err := validator.Validate("")
	if err == nil {
		t.Error("expected error for empty key")
	}
}

func TestValidateMultipleKeys(t *testing.T) {
	validator := NewAPIKeyValidator([]config.APIKey{
		{Hash: HashKey("key-1"), User: "alice", Team: "sre"},
		{Hash: HashKey("key-2"), User: "bob", Team: "dev"},
	})

	id, err := validator.Validate("key-2")
	if err != nil {
		t.Fatal(err)
	}
	if id.User != "bob" {
		t.Errorf("expected bob, got %s", id.User)
	}
}
