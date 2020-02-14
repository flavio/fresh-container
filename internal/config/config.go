package config

import (
	"encoding/json"
	"os"
)

const DEFAULT_CACHE_TTL_HOURS = 2

type RegistryConfig struct {
	AuthDomain string `json:"auth_domain"`
	Insecure   bool   `json:"insecure"`
	NonSSL     bool   `json:"non_ssl"`
	SkipPing   bool   `json:"skip_ping"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

type Config struct {
	Registries    map[string]RegistryConfig `json:"registries"`
	CacheTTLHours int                       `json:"cache_ttl_hours"`
}

func (c *Config) GetRegistryConfig(domain string) RegistryConfig {
	rc, found := c.Registries[domain]
	if found {
		return rc
	}

	return RegistryConfig{
		AuthDomain: domain,
		Insecure:   false,
		NonSSL:     false,
		SkipPing:   false,
	}
}

func NewConfig() Config {
	cfg := Config{}
	cfg.fixDefaults()

	return cfg
}

func NewFromFile(path string) (cfg Config, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		return
	}
	cfg.fixDefaults()

	return
}

func (c *Config) fixDefaults() {
	if c.CacheTTLHours == 0 {
		c.CacheTTLHours = DEFAULT_CACHE_TTL_HOURS
	}
}
