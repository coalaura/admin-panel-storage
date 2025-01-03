package main

import (
	"encoding/json"
	"net"
	"os"
)

type Config struct {
	Root       string   `json:"root"`
	Hostname   string   `json:"hostname"`
	Port       int      `json:"port"`
	AllowedIPs []string `json:"allowed_ips"`
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

	// Default hostname is 0.0.0.0
	if c.Hostname == "" {
		c.Hostname = "0.0.0.0"
	}

	// Default allowed_ips is ["*"]
	if len(c.AllowedIPs) == 0 {
		c.AllowedIPs = []string{"*"}
	}
}

func (c *Config) IsIPAllowed(ip string) bool {
	if len(c.AllowedIPs) == 0 || c.AllowedIPs[0] == "*" {
		return true
	}

	address, _, err := net.SplitHostPort(ip)
	if err != nil {
		log.Warningf("Failed to parse IP address: %s\n", err)

		return false
	}

	for _, allowed := range c.AllowedIPs {
		if allowed == "*" || allowed == address {
			return true
		}
	}

	return false
}
