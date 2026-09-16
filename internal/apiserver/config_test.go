package apiserver_test

import (
	"log/slog"
	"testing"

	"github.com/PoCInnovation/SecureEnv/internal/apiserver"
)

func env(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

func TestLoadConfigDefaults(t *testing.T) {
	t.Parallel()

	cfg, err := apiserver.LoadConfig(env(nil))
	if err != nil {
		t.Fatal(err)
	}
	want := apiserver.Config{ListenAddr: ":8080", VaultMount: "secret", LogLevel: slog.LevelInfo}
	if cfg != want {
		t.Fatalf("LoadConfig() = %+v, want %+v", cfg, want)
	}
	if cfg.TLSEnabled() {
		t.Error("TLS must be disabled by default")
	}
}

func TestLoadConfigOverrides(t *testing.T) {
	t.Parallel()

	cfg, err := apiserver.LoadConfig(env(map[string]string{
		apiserver.EnvListenAddr:  "127.0.0.1:9000",
		apiserver.EnvVaultMount:  "kv",
		apiserver.EnvVaultPrefix: "secureenv",
		apiserver.EnvTLSCertFile: "cert.pem",
		apiserver.EnvTLSKeyFile:  "key.pem",
		apiserver.EnvLogLevel:    "debug",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ListenAddr != "127.0.0.1:9000" || cfg.VaultMount != "kv" || cfg.VaultPrefix != "secureenv" ||
		!cfg.TLSEnabled() || cfg.LogLevel != slog.LevelDebug {
		t.Fatalf("LoadConfig() = %+v", cfg)
	}
}

func TestLoadConfigErrors(t *testing.T) {
	t.Parallel()

	for name, values := range map[string]map[string]string{
		"cert without key": {apiserver.EnvTLSCertFile: "cert.pem"},
		"key without cert": {apiserver.EnvTLSKeyFile: "key.pem"},
		"bad log level":    {apiserver.EnvLogLevel: "verbose"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := apiserver.LoadConfig(env(values)); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
