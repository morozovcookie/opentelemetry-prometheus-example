package main

import (
	"fmt"
	"os"

	uberzap "go.uber.org/zap"
)

type HTTPConfig struct {
	Address string
}

func NewHTTPConfig() *HTTPConfig {
	return &HTTPConfig{
		Address: "127.0.0.1:8080",
	}
}

func (cfg *HTTPConfig) Parse() error {
	if addr := os.Getenv("SERVER_HTTP_ADDRESS"); addr != "" {
		cfg.Address = addr
	}

	return nil
}

type MonitorConfig struct {
	Address string
}

func NewMonitorConfig() *MonitorConfig {
	return &MonitorConfig{
		Address: "127.0.0.1:9090",
	}
}

func (cfg *MonitorConfig) Parse() error {
	if addr := os.Getenv("SERVER_MONITOR_ADDRESS"); addr != "" {
		cfg.Address = addr
	}

	return nil
}

type PerconaConfig struct {
	Dsn string
}

func NewPerconaConfig() *PerconaConfig {
	return &PerconaConfig{
		Dsn: "",
	}
}

func (cfg *PerconaConfig) Parse() error {
	if dsn := os.Getenv("SERVER_PERCONA_DSN"); dsn != "" {
		cfg.Dsn = dsn
	}

	return nil
}

type Config struct {
	*HTTPConfig
	*MonitorConfig
	*PerconaConfig

	LogLevel string

	ZapLevel uberzap.AtomicLevel
}

func NewConfig() *Config {
	return &Config{
		HTTPConfig:    NewHTTPConfig(),
		MonitorConfig: NewMonitorConfig(),
		PerconaConfig: NewPerconaConfig(),

		LogLevel: "",

		ZapLevel: uberzap.NewAtomicLevelAt(uberzap.ErrorLevel),
	}
}

func (cfg *Config) Parse() error {
	for _, cfg := range []interface {
		Parse() error
	}{
		cfg.HTTPConfig,
		cfg.MonitorConfig,
		cfg.PerconaConfig,
	} {
		if err := cfg.Parse(); err != nil {
			return fmt.Errorf("parse config: %w", err)
		}
	}

	if err := cfg.parse(); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	return nil
}

func (cfg *Config) parse() error {
	var err error

	if logLevel := os.Getenv("SERVER_LOG_LEVEL"); logLevel != "" {
		cfg.LogLevel = logLevel
	}

	if cfg.ZapLevel, err = uberzap.ParseAtomicLevel(cfg.LogLevel); err != nil {
		return err
	}

	return nil
}
