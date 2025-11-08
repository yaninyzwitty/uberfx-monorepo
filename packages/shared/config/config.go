package config

import (
	"fmt"
	"os"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.yaml.in/yaml/v3"
)

type Config struct {
	DbConfig     DbConfig     `yaml:"database"`
	ServerConfig ServerConfig `yaml:"server"`
}

type DbConfig struct {
	Username string `yaml:"username"`
	Host     string `yaml:"host"`
	Database string `yaml:"database"`
	Port     int    `yaml:"port"`
	SslMode  string `yaml:"sslmode"`
	Password string `yaml:"password"` // Optional: can be overridden by DB_PASSWORD env var
}

type ServerConfig struct {
	GatewayPort        int `yaml:"gateway_port"`
	ProductServicePort int `yaml:"product_service_port"`
}

// Module exports the configuration provider
// Loads configuration from YAML file and provides it to the application
var Module = fx.Module("config",
	fx.Provide(NewConfig),
)

func NewConfig(log *zap.Logger) (*Config, error) {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "config.yaml" // default fallback
	}

	// Try to read the config file
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		// If file doesn't exist in current directory, try common locations
		if os.IsNotExist(err) {
			// Try relative to the executable or common config locations
			alternativePaths := []string{
				"./config.yaml",
				"./packages/product-service/config.yaml",
				"./packages/gateway-service/config.yaml",
			}

			for _, altPath := range alternativePaths {
				if data, err = os.ReadFile(altPath); err == nil {
					cfgPath = altPath
					log.Info("Found config file at alternative path", zap.String("path", altPath))
					break
				}
			}

			if err != nil {
				return nil, fmt.Errorf("failed to read config file (tried: %s and alternatives): %w", cfgPath, err)
			}
		} else {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var conf Config
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	log.Info("configuration loaded successfully",
		zap.String("path", cfgPath),
		zap.String("database_host", conf.DbConfig.Host),
		zap.Int("product_service_port", conf.ServerConfig.ProductServicePort),
	)
	return &conf, nil
}
