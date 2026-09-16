// Package apiserver wires the API server together: configuration, Vault
// client, HTTP handler and lifecycle.
package apiserver

import (
	"errors"
	"fmt"
	"log/slog"
)

// Config holds the API server settings. Vault connection settings (address,
// CA certificate, TLS options...) use the standard VAULT_* variables read by
// the Vault client itself.
type Config struct {
	ListenAddr  string
	VaultMount  string
	VaultPrefix string
	TLSCertFile string
	TLSKeyFile  string
	LogLevel    slog.Level
}

// Environment variables read by LoadConfig.
const (
	EnvListenAddr  = "SECURE_ENV_LISTEN_ADDR"
	EnvVaultMount  = "SECURE_ENV_VAULT_MOUNT"
	EnvVaultPrefix = "SECURE_ENV_VAULT_PREFIX"
	EnvTLSCertFile = "SECURE_ENV_TLS_CERT_FILE"
	EnvTLSKeyFile  = "SECURE_ENV_TLS_KEY_FILE"
	EnvLogLevel    = "SECURE_ENV_LOG_LEVEL"
)

// LoadConfig builds a Config from getenv, usually os.Getenv.
func LoadConfig(getenv func(string) string) (Config, error) {
	cfg := Config{
		ListenAddr:  valueOr(getenv(EnvListenAddr), ":8080"),
		VaultMount:  valueOr(getenv(EnvVaultMount), "secret"),
		VaultPrefix: getenv(EnvVaultPrefix),
		TLSCertFile: getenv(EnvTLSCertFile),
		TLSKeyFile:  getenv(EnvTLSKeyFile),
	}

	if (cfg.TLSCertFile == "") != (cfg.TLSKeyFile == "") {
		return Config{}, fmt.Errorf("%s and %s must be set together", EnvTLSCertFile, EnvTLSKeyFile)
	}
	if raw := getenv(EnvLogLevel); raw != "" {
		if err := cfg.LogLevel.UnmarshalText([]byte(raw)); err != nil {
			return Config{}, errors.Join(fmt.Errorf("invalid %s", EnvLogLevel), err)
		}
	}
	return cfg, nil
}

// TLSEnabled reports whether the server terminates TLS itself.
func (c Config) TLSEnabled() bool { return c.TLSCertFile != "" }

func valueOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
