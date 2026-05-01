package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	webauthnlib "github.com/go-webauthn/webauthn/webauthn"
)

const (
	configDirName  = "passkey-sudo"
	configFileName = "config.json"
	defaultRPID    = "localhost"
	defaultOrigin  = "http://localhost:14141"
)

type Config struct {
	Version             int                      `json:"version"`
	RPID                string                   `json:"rp_id"`
	RPOrigin            string                   `json:"rp_origin"`
	RPDisplayName       string                   `json:"rp_display_name"`
	Username            string                   `json:"username"`
	UserID              string                   `json:"user_id"`
	Credentials         []webauthnlib.Credential `json:"credentials"`
	AllowedCommands     []string                 `json:"allowed_commands"`
	SudoNonInteractive  bool                     `json:"sudo_non_interactive"`
	OpenBrowserOnPrompt bool                     `json:"open_browser_on_prompt"`
}

func DefaultPath() string {
	base, err := os.UserConfigDir()
	if err != nil {
		base = "."
	}
	return filepath.Join(base, configDirName, configFileName)
}

func LoadOrInitDefault() (*Config, error) {
	if _, err := os.Stat(DefaultPath()); err == nil {
		cfg, err := Load(DefaultPath())
		if err != nil {
			return nil, err
		}
		changed, err := normalizeWebAuthnDomain(cfg)
		if err != nil {
			return nil, err
		}
		if changed {
			if err := SaveDefault(cfg); err != nil {
				return nil, err
			}
		}
		return cfg, nil
	}
	return Init(defaultRPID, defaultOrigin, "Passkey-Sudo", "local-admin")
}

func Init(rpID, rpOrigin, rpName, username string) (*Config, error) {
	cfg := &Config{
		Version:             1,
		RPID:                rpID,
		RPOrigin:            rpOrigin,
		RPDisplayName:       rpName,
		Username:            username,
		UserID:              randomUserID(),
		AllowedCommands:     []string{},
		SudoNonInteractive:  true,
		OpenBrowserOnPrompt: true,
	}

	if _, err := os.Stat(DefaultPath()); err == nil {
		existing, err := Load(DefaultPath())
		if err == nil {
			cfg.Credentials = existing.Credentials
			cfg.AllowedCommands = existing.AllowedCommands
		}
	}

	if _, err := normalizeWebAuthnDomain(cfg); err != nil {
		return nil, err
	}

	if err := Save(DefaultPath(), cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

func Save(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return err
	}
	return nil
}

func SaveDefault(cfg *Config) error {
	if _, err := normalizeWebAuthnDomain(cfg); err != nil {
		return err
	}
	return Save(DefaultPath(), cfg)
}

func normalizeWebAuthnDomain(cfg *Config) (bool, error) {
	changed := false

	if strings.TrimSpace(cfg.RPOrigin) == "" {
		cfg.RPOrigin = defaultOrigin
		changed = true
	}

	u, err := url.Parse(cfg.RPOrigin)
	if err != nil {
		return changed, fmt.Errorf("invalid rp-origin %q: %w", cfg.RPOrigin, err)
	}
	if u.Scheme == "" {
		u.Scheme = "http"
		changed = true
	}
	if u.Host == "" {
		u.Host = defaultRPID + ":14141"
		changed = true
	}

	host := u.Hostname()
	if host == "" {
		host = defaultRPID
	}

	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		host = defaultRPID
		changed = true
	}

	if strings.TrimSpace(cfg.RPID) == "" || cfg.RPID != host {
		cfg.RPID = host
		changed = true
	}

	if u.Hostname() != host {
		port := u.Port()
		if port != "" {
			u.Host = net.JoinHostPort(host, port)
		} else {
			u.Host = host
		}
		changed = true
	}

	if u.Path == "" {
		u.Path = "/"
	}

	normalized := strings.TrimRight(u.String(), "/")
	if normalized != cfg.RPOrigin {
		cfg.RPOrigin = normalized
		changed = true
	}

	return changed, nil
}

func randomUserID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "passkey-sudo-user"
	}
	return hex.EncodeToString(buf)
}
