package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Env string

const (
	Development Env = "development"
	Production  Env = "production"
)

type Config struct {
	Env          Env
	Host         string
	Port         int
	BaseURL      string
	SiteName     string
	TemplatesDir string
	StaticDir    string
}

const envPrefix = "SITEKIT_"

func Load() (*Config, error) {
	cfg := &Config{
		Env:          Env(getEnv("ENV", string(Development))),
		Host:         getEnv("HOST", "0.0.0.0"),
		Port:         getEnvInt("PORT", 8080),
		BaseURL:      getEnv("BASE_URL", "http://localhost:8080"),
		SiteName:     getEnv("SITE_NAME", "Sitekit Site"),
		TemplatesDir: getEnv("TEMPLATES_DIR", "templates"),
		StaticDir:    getEnv("STATIC_DIR", "static"),
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	switch c.Env {
	case Development, Production:
	default:
		return fmt.Errorf("invalid %sENV value %q: must be %q or %q", envPrefix, c.Env, Development, Production)
	}

	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid %sPORT value %d: must be between 1 and 65535", envPrefix, c.Port)
	}

	if strings.TrimSpace(c.BaseURL) == "" {
		return fmt.Errorf("%sBASE_URL must not be empty", envPrefix)
	}

	return nil
}

func (c *Config) IsDevelopment() bool {
	return c.Env == Development
}

func (c *Config) IsProduction() bool {
	return c.Env == Production
}

func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(envPrefix + key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	raw := getEnv(key, "")
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}
