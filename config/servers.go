package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

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

func configPath() (string, error) {
	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg == "" {
		u, err := user.Current()
		if err != nil {
			return "", err
		}
		xdg = filepath.Join(u.HomeDir, ".config")
	}
	return filepath.Join(xdg, "lazymosh", "servers.json"), nil
}

// Load reads the config file. Returns an empty Store if missing or unreadable.
func Load() (*Store, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Store{}, nil
		}
		return nil, err
	}
	var s Store
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse servers.json: %w", err)
	}
	return &s, nil
}

// Save writes the store to the config file.
func Save(s *Store) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
