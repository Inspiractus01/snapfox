package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Jobs []BackupJob `json:"jobs"`
}

func getConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".snapfox"), nil
}

func getConfigPath() (string, error) {
	dir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func LoadConfig() (*Config, error) {
	path, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	// If config doesn't exist yet, return empty config
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return &Config{Jobs: []BackupJob{}}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		// Try to migrate older config where job.id was a number
		migrated, migErr := migrateOldConfig(data)
		if migErr != nil {
			// migration failed too – vrátime pôvodnú chybu
			return nil, err
		}
		cfg = *migrated
	}

	if cfg.Jobs == nil {
		cfg.Jobs = []BackupJob{}
	}

	return &cfg, nil
}

// SaveConfig writes config.json atomically
func SaveConfig(cfg *Config) error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	dir, err := getConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}

// migrateOldConfig tries to handle old configs where job.id was a JSON number
func migrateOldConfig(data []byte) (*Config, error) {
	// Generic structure: { "jobs": [ { ... } ] }
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	jobsRaw, ok := raw["jobs"].([]interface{})
	if !ok {
		// no jobs or unexpected format → necháme to tak
		return nil, fmt.Errorf("no jobs field to migrate")
	}

	for _, job := range jobsRaw {
		m, ok := job.(map[string]interface{})
		if !ok {
			continue
		}
		if v, exists := m["id"]; exists {
			switch vv := v.(type) {
			case float64:
				// convert 1 → "1"
				m["id"] = fmt.Sprintf("%.0f", vv)
			}
		}
	}

	// serialize back to JSON and unmarshal do Config
	fixedData, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(fixedData, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}