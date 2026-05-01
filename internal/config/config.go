package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	webauthnlib "github.com/go-webauthn/webauthn/webauthn"
)

const (
	configDirName  = "passkey-sudo"
	configFileName = "config.json"
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
		return Load(DefaultPath())
	}
	return Init("localhost", "http://127.0.0.1:14141", "Passkey-Sudo", "local-admin")
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
	return Save(DefaultPath(), cfg)
}

func randomUserID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "passkey-sudo-user"
	}
	return hex.EncodeToString(buf)
}
