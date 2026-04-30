package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"sync"
)

// ─── Path override (set via --file flag) ───────────────────────────────────

var (
	pathOverride string
	pathMu       sync.Mutex
)

// SetPath overrides the config file path. Mostly useful for testing.
func SetPath(p string) {
	pathMu.Lock()
	pathOverride = p
	pathMu.Unlock()
}

// Path returns the resolved config file path.
func Path() string {
	pathMu.Lock()
	defer pathMu.Unlock()
	if pathOverride != "" {
		return pathOverride
	}
	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg == "" {
		u, err := user.Current()
		if err != nil {
			return ""
		}
		xdg = filepath.Join(u.HomeDir, ".config")
	}
	return filepath.Join(xdg, "lazymosh", "servers.json")
}

// DefaultPath returns the default XDG path (ignores override).
func DefaultPath() string {
	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg == "" {
		u, err := user.Current()
		if err != nil {
			return ""
		}
		xdg = filepath.Join(u.HomeDir, ".config")
	}
	return filepath.Join(xdg, "lazymosh", "servers.json")
}

// ─── Config types ────────────────────────────────────────────────────────────

// Server represents a single mosh/ssh target.
type Server struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Locality string `json:"locality"` // e.g. "us-east", "eu-berlin"
}

// Store is the top-level config file.
type Store struct {
	Servers []Server `json:"servers"`
}

// Load reads the config file. Returns an empty Store if missing or unreadable.
func Load() (*Store, error) {
	path := Path()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Store{}, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var s Store
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &s, nil
}

// Save writes the store to the config file.
func Save(s *Store) error {
	path := Path()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
