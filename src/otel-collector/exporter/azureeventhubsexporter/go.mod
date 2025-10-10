module github.com/fedeoliv/custom-otel-collector/exporter/azureeventhubsexporter

go 1.21

require (
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.9.0
	github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs v1.2.1
	go.opentelemetry.io/collector/component v0.106.1
	go.opentelemetry.io/collector/config/configretry v0.106.1
	go.opentelemetry.io/collector/consumer v0.106.1
	go.opentelemetry.io/collector/exporter v0.106.1
	go.opentelemetry.io/collector/exporter/exporterhelper v0.106.1
	go.opentelemetry.io/collector/pdata v1.12.0
	go.opentelemetry.io/collector/pipeline v0.106.1
	go.uber.org/zap v1.27.0
)