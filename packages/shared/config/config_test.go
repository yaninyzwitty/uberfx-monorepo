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
	defer os.Remove(tmpFile.Name())
	
	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tmpFile.Close()

	// Set environment variable
	os.Setenv("CONFIG_PATH", tmpFile.Name())
	defer os.Unsetenv("CONFIG_PATH")

	logger := zap.NewNop()
	cfg, err := NewConfig(logger)
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}

	if cfg.DbConfig.Host != "test-host" {
		t.Errorf("Expected DB host 'test-host', got '%s'", cfg.DbConfig.Host)
	}
}
