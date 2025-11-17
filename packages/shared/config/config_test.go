package config

import (
	"os"
	"testing"

	"go.uber.org/zap"
)

func TestNewConfig(t *testing.T) {
	// Create a temporary config file
	configContent := `
database:
  host: test-host
  port: 5432
  username: test-user
  database: test-db
server:
  port: 8080
`

	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Fatalf("Failed to temporary file name: %v", err)

		}

	}()

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close the temp file: %v", err)
	}

	// Set environment variable
	if err := os.Setenv("CONFIG_PATH", tmpFile.Name()); err != nil {
		t.Fatalf("Failed to set the config path: %v", err)

	}
	defer func() {
		if err := os.Unsetenv("CONFIG_PATH"); err != nil {
			t.Fatalf("Failed to unset the config path: %v", err)

		}

	}()

	logger := zap.NewNop()
	cfg, err := NewConfig(logger)
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}

	if cfg.DbConfig.Host != "test-host" {
		t.Errorf("Expected DB host 'test-host', got '%s'", cfg.DbConfig.Host)
	}
}
