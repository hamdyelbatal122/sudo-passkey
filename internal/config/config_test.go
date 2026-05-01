package config

import "testing"

func TestRandomUserIDNotEmpty(t *testing.T) {
	id := randomUserID()
	if id == "" {
		t.Fatal("expected randomUserID to return non-empty value")
	}
}

func TestNormalizeWebAuthnDomain_LoopbackOrigin(t *testing.T) {
	cfg := &Config{
		RPID:     "localhost",
		RPOrigin: "http://127.0.0.1:14141",
	}

	changed, err := normalizeWebAuthnDomain(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Fatal("expected config to be normalized")
	}
	if cfg.RPID != "localhost" {
		t.Fatalf("expected rp_id localhost, got %q", cfg.RPID)
	}
	if cfg.RPOrigin != "http://localhost:14141" {
		t.Fatalf("expected rp_origin http://localhost:14141, got %q", cfg.RPOrigin)
	}
}

func TestNormalizeWebAuthnDomain_AlignMismatchedHost(t *testing.T) {
	cfg := &Config{
		RPID:     "localhost",
		RPOrigin: "http://example.com:14141",
	}

	_, err := normalizeWebAuthnDomain(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.RPID != "example.com" {
		t.Fatalf("expected rp_id example.com, got %q", cfg.RPID)
	}
	if cfg.RPOrigin != "http://example.com:14141" {
		t.Fatalf("expected rp_origin http://example.com:14141, got %q", cfg.RPOrigin)
	}
}
