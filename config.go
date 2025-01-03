package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	Root     string `json:"root"`
	Hostname string `json:"hostname"`
	Port     int    `json:"port"`
}

const ConfigFile = "config.json"

func SaveDefaultConfig() (*Config, error) {
	var cfg Config

	cfg.SetDefaults()

	b, err := json.MarshalIndent(cfg, "", "\t")
	if err != nil {
		return nil, err
	}

	err = os.WriteFile(ConfigFile, b, 0644)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func ReadConfig() (*Config, error) {
	if _, err := os.Stat(ConfigFile); os.IsNotExist(err) {
		return SaveDefaultConfig()
	}

	b, err := os.ReadFile(ConfigFile)
	if err != nil {
		return nil, err
	}

	var cfg Config

	err = json.Unmarshal(b, &cfg)
	if err != nil {
		return nil, err
	}

	cfg.SetDefaults()

	return &cfg, nil
}

func (c *Config) SetDefaults() {
	// Default root is "./storage"
	if c.Root == "" {
		c.Root = "./storage"
	}

	// Default port is 4994
	if c.Port == 0 {
		c.Port = 4994
	}

	// Default hostname is 127.0.0.1
	if c.Hostname == "" {
		c.Hostname = "127.0.0.1"
	}
}
