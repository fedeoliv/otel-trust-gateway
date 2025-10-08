module github.com/fedeoliv/custom-otel-collector/exporter/azureblobexporter

go 1.24.7

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.17.0
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.9.0
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.5.0
	go.opentelemetry.io/collector/component v1.42.0
	go.opentelemetry.io/collector/config/configretry v1.42.0
	go.opentelemetry.io/collector/consumer v1.42.0
	go.opentelemetry.io/collector/exporter v1.42.0
	go.opentelemetry.io/collector/exporter/exporterhelper v0.136.0
	go.opentelemetry.io/collector/pdata v1.42.0
	go.opentelemetry.io/collector/pipeline v1.42.0
	go.uber.org/zap v1.27.0
)
