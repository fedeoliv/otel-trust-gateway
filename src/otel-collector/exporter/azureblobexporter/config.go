// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package azureblobexporter

import (
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configretry"
)

type TelemetryConfig struct {
	Logs    string `mapstructure:"logs"`
	Metrics string `mapstructure:"metrics"`
	Traces  string `mapstructure:"traces"`
}

type Encodings struct {
	Logs    *component.ID `mapstructure:"logs"`
	Metrics *component.ID `mapstructure:"metrics"`
	Traces  *component.ID `mapstructure:"traces"`
}

type BlobNameFormat struct {
	MetricsFormat            string            `mapstructure:"metrics_format"`
	LogsFormat               string            `mapstructure:"logs_format"`
	TracesFormat             string            `mapstructure:"traces_format"`
	SerialNumRange           int64             `mapstructure:"serial_num_range"`
	SerialNumBeforeExtension bool              `mapstructure:"serial_num_before_extension"`
	TemplateEnabled          bool              `mapstructure:"template_enabled"`
	Params                   map[string]string `mapstructure:"params"`
}

type AppendBlob struct {
	Enabled   bool   `mapstructure:"enabled"`
	Separator string `mapstructure:"separator"`
}

type Authentication struct {
	// Type is the authentication type. supported values are connection_string, service_principal, system_managed_identity, user_managed_identity, workload_identity, and default_credentials
	Type AuthType `mapstructure:"type"`

	// TenantID is the tenand id for the AAD App. It's only needed when type is service_principal or workload_identity.
	TenantID string `mapstructure:"tenant_id"`

	// ClientID is the AAD Application client id. It's needed when type is service_principal, user_managed_identity or workload_identity
	ClientID string `mapstructure:"client_id"`
	// ClientSecret only needed when auth type is service_principal

	ClientSecret string `mapstructure:"client_secret"`

	// ConnectionString to the endpoint.
	ConnectionString string `mapstructure:"connection_string"`

	// FederatedTokenFile is the path to the file containing the federated token. It's needed when type is workload_identity.
	FederatedTokenFile string `mapstructure:"federated_token_file"`
}

type AuthType string

const (
	ConnectionString      AuthType = "connection_string"
	SystemManagedIdentity AuthType = "system_managed_identity"
	UserManagedIdentity   AuthType = "user_managed_identity"
	ServicePrincipal      AuthType = "service_principal"
	WorkloadIdentity      AuthType = "workload_identity"
	DefaultCredentials    AuthType = "default_credentials"
)

// Config contains the main configuration options for the azure storage blob exporter
type Config struct {
	// URL is the endpoint to the azure storage account. This is only required until there is an azure auth extension in the future.
	URL string `mapstructure:"url"`

	// A container organizes a set of blobs, similar to a directory in a file system.
	Container TelemetryConfig `mapstructure:"container"`
	Auth      Authentication  `mapstructure:"auth"`

	// BlobNameFormat is the format of the blob name. It controls the uploaded blob name, e.g. "2006/01/02/metrics_15_04_05.json"
	BlobNameFormat BlobNameFormat `mapstructure:"blob_name_format"`

	// FormatType is the format of encoded telemetry data. Supported values are json, proto, and parquet.
	FormatType string `mapstructure:"format"`

	// AppendBlob configures append blob behavior
	AppendBlob AppendBlob `mapstructure:"append_blob"`

	// Encoding extension to apply for logs/metrics/traces. If present, overrides the marshaler configuration option and format.
	Encodings Encodings `mapstructure:"encodings"`

	configretry.BackOffConfig `mapstructure:"retry_on_failure"`
}

func (c *Config) Validate() error {
	if c.URL == "" && c.Auth.Type != ConnectionString {
		return errors.New("url cannot be empty when auth type is not connection_string")
	}

	switch c.Auth.Type {
	case ConnectionString:
		if c.Auth.ConnectionString == "" {
			return errors.New("connection_string cannot be empty when auth type is connection_string")
		}
	case ServicePrincipal:
		if c.Auth.TenantID == "" || c.Auth.ClientID == "" || c.Auth.ClientSecret == "" {
			return errors.New("tenant_id, client_id and client_secret cannot be empty when auth type is service-principal")
		}
	case UserManagedIdentity:
		if c.Auth.ClientID == "" {
			return errors.New("client_id cannot be empty when auth type is user_managed_identity")
		}
	case WorkloadIdentity:
		if c.Auth.TenantID == "" || c.Auth.ClientID == "" || c.Auth.FederatedTokenFile == "" {
			return errors.New("tenant_id, client_id and federated_token_file cannot be empty when auth type is workload_identity")
		}
	case DefaultCredentials:
		// No additional fields required for default credentials
		// DefaultAzureCredential will automatically detect credentials from environment
	}

	if c.FormatType != "json" && c.FormatType != "proto" && c.FormatType != "parquet" {
		return errors.New("unknown format type: " + c.FormatType)
	}

	return nil
}
