package trustgatewayprocessor

import (
	"go.opentelemetry.io/collector/component"
)

// Config defines the configuration for the trust gateway processor
type Config struct {
	// RequiredHeaders are the HTTP headers that must be present
	RequiredHeaders []string `mapstructure:"required_headers"`
	// ValidAPIKeys are the valid API keys for authentication
	ValidAPIKeys []string `mapstructure:"valid_api_keys"`
}

var _ component.Config = (*Config)(nil)

// Validate checks if the processor configuration is valid
func (cfg *Config) Validate() error {
	return nil
}
