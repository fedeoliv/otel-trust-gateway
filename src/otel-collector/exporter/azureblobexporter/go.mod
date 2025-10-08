module github.com/fedeoliv/custom-otel-collector/exporter/azureblobexporter

go 1.24.7

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.18.0
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.9.0
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.5.0
	github.com/parquet-go/parquet-go v0.25.1
	go.opentelemetry.io/collector/component v1.42.0
	go.opentelemetry.io/collector/config/configretry v1.42.0
	go.opentelemetry.io/collector/consumer v1.42.0
	go.opentelemetry.io/collector/exporter v1.42.0
	go.opentelemetry.io/collector/exporter/exporterhelper v0.136.0
	go.opentelemetry.io/collector/pdata v1.42.0
	go.opentelemetry.io/collector/pipeline v1.42.0
	go.uber.org/zap v1.27.0
)

require (
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	golang.org/x/sys v0.35.0 // indirect
)
