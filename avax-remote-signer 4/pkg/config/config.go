package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Backend BackendConfig `mapstructure:"backend"`
	Logging LoggingConfig `mapstructure:"logging"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	HTTPAddress string    `mapstructure:"http_address"`
	GRPCAddress string    `mapstructure:"grpc_address"`
	Address     string    `mapstructure:"address"` // deprecated, use grpc_address
	TLS         TLSConfig `mapstructure:"tls"`
}

// TLSConfig holds TLS certificate configuration
type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

// BackendConfig holds key management backend configuration
type BackendConfig struct {
	Type   string                 `mapstructure:"type"`
	Config map[string]interface{} `mapstructure:"config"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"` // "json" or "console"
}

// LoadConfig loads configuration from file and environment
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("server.grpc_address", "0.0.0.0:9090")
	v.SetDefault("server.http_address", "0.0.0.0:8080")
	v.SetDefault("server.address", "0.0.0.0:9090") // deprecated
	v.SetDefault("server.tls.enabled", false)
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "console")
	v.SetDefault("backend.type", "file")

	// Read from config file if provided
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Read from environment variables
	v.SetEnvPrefix("AVAX_SIGNER")
	v.AutomaticEnv()

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// validateConfig performs basic validation on the configuration
func validateConfig(config *Config) error {
	// Support legacy address field
	if config.Server.GRPCAddress == "" && config.Server.Address != "" {
		config.Server.GRPCAddress = config.Server.Address
	}
	if config.Server.GRPCAddress == "" {
		return fmt.Errorf("server grpc_address cannot be empty")
	}
	if config.Server.HTTPAddress == "" {
		return fmt.Errorf("server http_address cannot be empty")
	}

	if config.Server.TLS.Enabled {
		if config.Server.TLS.CertFile == "" {
			return fmt.Errorf("TLS cert file required when TLS is enabled")
		}
		if config.Server.TLS.KeyFile == "" {
			return fmt.Errorf("TLS key file required when TLS is enabled")
		}
	}

	if config.Backend.Type == "" {
		return fmt.Errorf("backend type cannot be empty")
	}

	return nil
}
