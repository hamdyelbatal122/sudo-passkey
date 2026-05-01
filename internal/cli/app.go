package cli

import (
	"errors"
	"flag"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/hamdyelbatal122/sudo-passkey/internal/config"
	"github.com/hamdyelbatal122/sudo-passkey/internal/gate"
)

func Run(args []string) int {
	if len(args) == 0 {
		printUsage()
		return 0
	}

	switch args[0] {
	case "help", "-h", "--help":
		printUsage()
		return 0
	case "version", "-v", "--version":
		fmt.Println(Version)
		return 0
	case "init":
		return runInit(args[1:])
	case "enroll":
		return runEnroll(args[1:])
	case "check":
		return runCheck(args[1:])
	case "run":
		return runCommand(args[1:])
	case "allow":
		return runAllow(args[1:])
	case "settings":
		return runSettings(args[1:])
	case "passkey":
		return runPasskey(args[1:])
	case "add-passkey":
		return runEnroll(args[1:])
	default:
		fmt.Printf("unknown command: %s\n\n", args[0])
		printUsage()
		return 2
	}
}

func runPasskey(args []string) int {
	if len(args) == 0 {
		fmt.Println("usage: passkey-sudo passkey <add|list|remove>")
		return 2
	}

	switch args[0] {
	case "add":
		return runEnroll(args[1:])
	case "list":
		cfg, err := config.LoadOrInitDefault()
		if err != nil {
			fmt.Printf("failed to load config: %v\n", err)
			return 1
		}
		if len(cfg.Credentials) == 0 {
			fmt.Println("no passkeys enrolled")
			return 0
		}
		for i, cred := range cfg.Credentials {
			fmt.Printf("%d) credential-id-bytes=%d\n", i+1, len(cred.ID))
		}
		return 0
	case "remove":
		if len(args) < 2 {
			fmt.Println("usage: passkey-sudo passkey remove <index>")
			return 2
		}
		idx, err := strconv.Atoi(args[1])
		if err != nil || idx < 1 {
			fmt.Println("index must be a positive number")
			return 2
		}
		cfg, err := config.LoadOrInitDefault()
		if err != nil {
			fmt.Printf("failed to load config: %v\n", err)
			return 1
		}
		if idx > len(cfg.Credentials) {
			fmt.Printf("index out of range. enrolled passkeys: %d\n", len(cfg.Credentials))
			return 2
		}
		cfg.Credentials = append(cfg.Credentials[:idx-1], cfg.Credentials[idx:]...)
		if err := config.SaveDefault(cfg); err != nil {
			fmt.Printf("failed to save config: %v\n", err)
			return 1
		}
		fmt.Println("passkey removed successfully")
		return 0
	default:
		fmt.Println("usage: passkey-sudo passkey <add|list|remove>")
		return 2
	}
}

func runAllow(args []string) int {
	if len(args) == 0 {
		fmt.Println("usage: passkey-sudo allow <list|add|remove>")
		return 2
	}

	cfg, err := config.LoadOrInitDefault()
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		return 1
	}

	switch args[0] {
	case "list":
		if len(cfg.AllowedCommands) == 0 {
			fmt.Println("allow list is empty. all commands are currently accepted")
			return 0
		}
		sorted := append([]string(nil), cfg.AllowedCommands...)
		sort.Strings(sorted)
		for i, cmd := range sorted {
			fmt.Printf("%d) %s\n", i+1, cmd)
		}
		return 0
	case "add":
		if len(args) < 2 {
			fmt.Println("usage: passkey-sudo allow add <command-or-path>")
			return 2
		}
		candidate := strings.TrimSpace(args[1])
		if candidate == "" {
			fmt.Println("command cannot be empty")
			return 2
		}
		if !strings.Contains(candidate, "/") {
			if resolved, err := exec.LookPath(candidate); err == nil {
				candidate = resolved
			}
		} else {
			candidate = filepath.Clean(candidate)
		}
		for _, existing := range cfg.AllowedCommands {
			if existing == candidate {
				fmt.Println("already exists in allow list")
				return 0
			}
		}
		cfg.AllowedCommands = append(cfg.AllowedCommands, candidate)
		if err := config.SaveDefault(cfg); err != nil {
			fmt.Printf("failed to save config: %v\n", err)
			return 1
		}
		fmt.Printf("added to allow list: %s\n", candidate)
		return 0
	case "remove":
		if len(args) < 2 {
			fmt.Println("usage: passkey-sudo allow remove <command-or-path>")
			return 2
		}
		target := strings.TrimSpace(args[1])
		if target == "" {
			fmt.Println("command cannot be empty")
			return 2
		}
		filtered := cfg.AllowedCommands[:0]
		removed := false
		for _, c := range cfg.AllowedCommands {
			if c == target {
				removed = true
				continue
			}
			filtered = append(filtered, c)
		}
		if !removed {
			fmt.Println("item not found in allow list")
			return 0
		}
		cfg.AllowedCommands = filtered
		if err := config.SaveDefault(cfg); err != nil {
			fmt.Printf("failed to save config: %v\n", err)
			return 1
		}
		fmt.Printf("removed from allow list: %s\n", target)
		return 0
	default:
		fmt.Println("usage: passkey-sudo allow <list|add|remove>")
		return 2
	}
}

func runSettings(args []string) int {
	if len(args) == 0 {
		fmt.Println("usage: passkey-sudo settings <show|set>")
		return 2
	}

	cfg, err := config.LoadOrInitDefault()
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		return 1
	}

	switch args[0] {
	case "show":
		fmt.Printf("rp_id: %s\n", cfg.RPID)
		fmt.Printf("rp_origin: %s\n", cfg.RPOrigin)
		fmt.Printf("rp_display_name: %s\n", cfg.RPDisplayName)
		fmt.Printf("username: %s\n", cfg.Username)
		fmt.Printf("sudo_non_interactive: %v\n", cfg.SudoNonInteractive)
		fmt.Printf("open_browser_on_prompt: %v\n", cfg.OpenBrowserOnPrompt)
		fmt.Printf("config_path: %s\n", config.DefaultPath())
		return 0
	case "set":
		if len(args) < 3 {
			fmt.Println("usage: passkey-sudo settings set <key> <value>")
			return 2
		}
		key := strings.TrimSpace(args[1])
		value := strings.TrimSpace(args[2])
		switch key {
		case "rp-id":
			cfg.RPID = value
		case "rp-origin":
			cfg.RPOrigin = value
		case "rp-name":
			cfg.RPDisplayName = value
		case "username":
			cfg.Username = value
		case "sudo-non-interactive":
			b, err := strconv.ParseBool(value)
			if err != nil {
				fmt.Println("value must be true or false")
				return 2
			}
			cfg.SudoNonInteractive = b
		case "open-browser":
			b, err := strconv.ParseBool(value)
			if err != nil {
				fmt.Println("value must be true or false")
				return 2
			}
			cfg.OpenBrowserOnPrompt = b
		default:
			fmt.Println("supported keys: rp-id, rp-origin, rp-name, username, sudo-non-interactive, open-browser")
			return 2
		}

		if err := config.SaveDefault(cfg); err != nil {
			fmt.Printf("failed to save config: %v\n", err)
			return 1
		}
		fmt.Println("settings updated")
		return 0
	default:
		fmt.Println("usage: passkey-sudo settings <show|set>")
		return 2
	}
}

func runInit(args []string) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(new(strings.Builder))
	rpID := fs.String("rp-id", "localhost", "Relying Party ID")
	rpOrigin := fs.String("rp-origin", "http://localhost:14141", "Relying Party origin")
	rpName := fs.String("rp-name", "Passkey-Sudo", "Relying Party display name")
	username := fs.String("username", "local-admin", "Display username for passkey registration")
	if err := fs.Parse(args); err != nil {
		fmt.Println(err)
		return 2
	}

	cfg, err := config.Init(*rpID, *rpOrigin, *rpName, *username)
	if err != nil {
		fmt.Printf("failed to initialize config: %v\n", err)
		return 1
	}

	fmt.Printf("initialized config at %s\n", config.DefaultPath())
	fmt.Printf("next step: passkey-sudo enroll\n")
	if len(cfg.Credentials) > 0 {
		fmt.Println("existing credentials were preserved")
	}
	return 0
}

func runEnroll(args []string) int {
	fs := flag.NewFlagSet("enroll", flag.ContinueOnError)
	fs.SetOutput(new(strings.Builder))
	if err := fs.Parse(args); err != nil {
		fmt.Println(err)
		return 2
	}

	cfg, err := config.LoadOrInitDefault()
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		return 1
	}

	if err := gate.Enroll(cfg); err != nil {
		fmt.Printf("enrollment failed: %v\n", err)
		return 1
	}

	fmt.Println("passkey enrolled successfully")
	return 0
}

func runCheck(args []string) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	fs.SetOutput(new(strings.Builder))
	if err := fs.Parse(args); err != nil {
		fmt.Println(err)
		return 2
	}

	cfg, err := config.LoadOrInitDefault()
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		return 1
	}

	if err := gate.Authenticate(cfg); err != nil {
		fmt.Printf("authentication failed: %v\n", err)
		return 1
	}

	fmt.Println("authentication passed")
	return 0
}

func runCommand(args []string) int {
	if len(args) == 0 {
		fmt.Println("usage: passkey-sudo run -- <command> [args...]")
		return 2
	}

	if args[0] == "--" {
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Println("missing command")
		return 2
	}

	cfg, err := config.LoadOrInitDefault()
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		return 1
	}

	if err := gate.AssertCommandAllowed(cfg, args[0]); err != nil {
		fmt.Println(err)
		return 1
	}

	if err := gate.Authenticate(cfg); err != nil {
		fmt.Printf("authentication failed: %v\n", err)
		return 1
	}

	exitCode, err := gate.ExecSudo(cfg, args)
	if err != nil {
		var ee gate.ExitCoder
		if errors.As(err, &ee) {
			return ee.ExitCode()
		}
		fmt.Printf("failed to execute command: %v\n", err)
		return 1
	}

	return exitCode
}

func printUsage() {
	fmt.Println("Passkey-Sudo (The Security Gate)")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  passkey-sudo init [--rp-id localhost --rp-origin http://localhost:14141 --rp-name Passkey-Sudo --username local-admin]")
	fmt.Println("  passkey-sudo enroll")
	fmt.Println("  passkey-sudo add-passkey")
	fmt.Println("  passkey-sudo passkey <add|list|remove>")
	fmt.Println("  passkey-sudo allow <list|add|remove>")
	fmt.Println("  passkey-sudo settings <show|set>")
	fmt.Println("  passkey-sudo check")
	fmt.Println("  passkey-sudo run -- <command> [args...]")
	fmt.Println("  passkey-sudo version")
}
