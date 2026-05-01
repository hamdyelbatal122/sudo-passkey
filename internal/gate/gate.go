package gate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/hamdy/passkey-sudo/internal/config"
	"github.com/hamdy/passkey-sudo/internal/webauthnserver"
)

type ExitCoder interface {
	error
	ExitCode() int
}

func Enroll(cfg *config.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	return webauthnserver.Run(ctx, cfg, webauthnserver.ModeRegister)
}

func Authenticate(cfg *config.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	return webauthnserver.Run(ctx, cfg, webauthnserver.ModeLogin)
}

func AssertCommandAllowed(cfg *config.Config, command string) error {
	if len(cfg.AllowedCommands) == 0 {
		return nil
	}
	resolved, err := exec.LookPath(command)
	if err != nil {
		resolved = command
	}
	resolved = filepath.Clean(resolved)

	for _, allowed := range cfg.AllowedCommands {
		if strings.TrimSpace(allowed) == "" {
			continue
		}
		if resolved == filepath.Clean(allowed) || command == allowed {
			return nil
		}
	}
	return fmt.Errorf("command %q is not in allowed list; update %s", command, config.DefaultPath())
}

func ExecSudo(cfg *config.Config, command []string) (int, error) {
	args := []string{}
	if cfg.SudoNonInteractive {
		args = append(args, "-n")
	}
	args = append(args, "--")
	args = append(args, command...)

	cmd := exec.Command("sudo", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err == nil {
		return 0, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode(), exitErr
	}
	return 1, err
}
