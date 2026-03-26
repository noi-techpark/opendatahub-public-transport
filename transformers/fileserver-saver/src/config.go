// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type SinkConfig struct {
	MQQueue        string `yaml:"mq_queue"`
	MQExchange     string `yaml:"mq_exchange"`
	MQKey          string `yaml:"mq_key"`
	FileserverPath string `yaml:"fileserver_path"`
}

type Config struct {
	Sinks []SinkConfig `yaml:"sinks"`
}

// SiriPayload matches what feed-fetcher publishes as rawdata.
type SiriPayload struct {
	Format   string            `json:"format"`
	Protocol string            `json:"protocol"`
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
	if len(cfg.Sinks) == 0 {
		return nil, fmt.Errorf("config has no sinks defined")
	}
	for i, s := range cfg.Sinks {
		if s.MQQueue == "" {
			return nil, fmt.Errorf("sink %d: mq_queue is required", i)
		}
		if s.MQExchange == "" {
			return nil, fmt.Errorf("sink %d: mq_exchange is required", i)
		}
		if s.MQKey == "" {
			return nil, fmt.Errorf("sink %d: mq_key is required", i)
		}
		if s.FileserverPath == "" {
			return nil, fmt.Errorf("sink %d: fileserver_path is required", i)
		}
	}
	return &cfg, nil
}
