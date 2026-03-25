// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type FeedConfig struct {
	Endpoint string            `yaml:"endpoint"`
	Provider string            `yaml:"provider"`
	Cron     string            `yaml:"cron"`
	Format   string            `yaml:"format"`
	Metadata map[string]string `yaml:"metadata,omitempty"`
	Headers  map[string]string `yaml:"headers,omitempty"`
}

type Config struct {
	Feeds []FeedConfig `yaml:"feeds"`
}

type SiriPayload struct {
	Format   string            `json:"format"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Payload  string            `json:"payload"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	if len(cfg.Feeds) == 0 {
		return nil, fmt.Errorf("config has no feeds defined")
	}
	for i, f := range cfg.Feeds {
		if f.Endpoint == "" {
			return nil, fmt.Errorf("feed %d: endpoint is required", i)
		}
		if f.Provider == "" {
			return nil, fmt.Errorf("feed %d: provider is required", i)
		}
		if f.Cron == "" {
			return nil, fmt.Errorf("feed %d: cron is required", i)
		}
		if f.Format == "" {
			return nil, fmt.Errorf("feed %d: format is required", i)
		}
	}
	return &cfg, nil
}
