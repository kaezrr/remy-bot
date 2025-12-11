package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Database   string `json:"database"`
	SessionDir string `json:"session_dir"`
	Prefix     string `json:"prefix"`
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = json.Unmarshal(b, &cfg)
	return &cfg, err
}
