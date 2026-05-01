package cli

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/hamdy/passkey-sudo/internal/config"
	"github.com/hamdy/passkey-sudo/internal/gate"
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
	default:
		fmt.Printf("unknown command: %s\n\n", args[0])
		printUsage()
		return 2
	}
}

func runInit(args []string) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(new(strings.Builder))
	rpID := fs.String("rp-id", "localhost", "Relying Party ID")
	rpOrigin := fs.String("rp-origin", "http://127.0.0.1:14141", "Relying Party origin")
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
	fmt.Println("  passkey-sudo init [--rp-id localhost --rp-origin http://127.0.0.1:14141 --rp-name Passkey-Sudo --username local-admin]")
	fmt.Println("  passkey-sudo enroll")
	fmt.Println("  passkey-sudo check")
	fmt.Println("  passkey-sudo run -- <command> [args...]")
	fmt.Println("  passkey-sudo version")
}
