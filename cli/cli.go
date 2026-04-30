package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"lazymosh/config"
	"lazymosh/log"
)

// Config holds all CLI flags and parsed options.
type Config struct {
	Help       bool
	Version    bool
	Verbosity  string // "error","warn","info","debug"
	ConfigPath string // --file / -f flag
}

// StringFlag implements flag.Value for dual -f/--file registration.
type StringFlag struct {
	Value *string
}

func (f StringFlag) String() string { return *f.Value }
func (f StringFlag) Set(s string) error {
	*f.Value = s
	return nil
}

// Parse reads os.Args, sets log level, applies config path override.
// Returns whether to exit early (help/version printed).
func Parse() (cfg Config, exit bool) {
	fs := flag.NewFlagSet("lazymosh", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {}

	fs.BoolVar(&cfg.Help, "h", false, "show help")
	fs.BoolVar(&cfg.Help, "help", false, "show help")
	fs.BoolVar(&cfg.Version, "version", false, "show version")

	fs.StringVar(&cfg.Verbosity, "verbosity", "info",
		"log level: error, warn, info, debug")
	fs.StringVar(&cfg.Verbosity, "V", "info",
		"log level: short (same as --verbosity)")

	// Register both -f and --file as the same string var
	fs.Var(StringFlag{Value: &cfg.ConfigPath}, "file", "path to servers.json")
	fs.Var(StringFlag{Value: &cfg.ConfigPath}, "f", "path to servers.json (short)")

	if err := fs.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			PrintHelp()
			return cfg, true
		}
		log.Error("flag parse: %v", err)
		return cfg, true
	}

	// Handle --file=<value> style (fs.Parse only handles --file value)
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--file=") {
			cfg.ConfigPath = strings.TrimPrefix(arg, "--file=")
		}
		if strings.HasPrefix(arg, "-f=") {
			cfg.ConfigPath = strings.TrimPrefix(arg, "-f=")
		}
	}

	// Normalize path (expand ~)
	if strings.HasPrefix(cfg.ConfigPath, "~/") {
		home := os.Getenv("HOME")
		if home != "" {
			cfg.ConfigPath = home + cfg.ConfigPath[1:]
		}
	}

	// Set log verbosity
	if !log.SetLevelFromString(cfg.Verbosity) {
		log.Error("unknown verbosity %q (allowed: error, warn, info, debug)", cfg.Verbosity)
		PrintHelp()
		return cfg, true
	}

	// Apply config path override
	if cfg.ConfigPath != "" {
		log.Debug("using config file: %s", cfg.ConfigPath)
		config.SetPath(cfg.ConfigPath)
	} else {
		log.Debug("using default config file: %s", config.DefaultPath())
	}

	// Handle help/version
	if cfg.Help {
		PrintHelp()
		return cfg, true
	}
	if cfg.Version {
		PrintVersion()
		return cfg, true
	}

	return cfg, false
}

// PrintHelp shows the full --help text.
func PrintHelp() {
	fmt.Fprint(os.Stderr, `
lazymosh — mosh/ssh launcher TUI

USAGE
  lazymosh [flags]

FLAGS
  -h, --help          show this help
  -v, --version       show version
  -V, --verbosity     log level: error, warn, info, debug (default: info)
  -f, --file PATH     use PATH as servers.json (default: ~/.config/lazymosh/servers.json)

EXAMPLES
  lazymosh                       normal run
  lazymosh -V debug              run with verbose logging
  lazymosh -f ~/.my-servers.json  use a custom config file
  lazymosh --file=~/.servers.json use = syntax

CONFIG
  Servers are stored in ~/.config/lazymosh/servers.json (XDG compliant).
  No passwords are stored — SSH keys are used for authentication.

`)
}

// PrintVersion shows version info.
func PrintVersion() {
	fmt.Fprint(os.Stderr, "lazymosh 0.1.0\n")
}
