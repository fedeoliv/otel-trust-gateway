package main

import (
	"log"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/envprovider"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/confmap/provider/httpprovider"
	"go.opentelemetry.io/collector/confmap/provider/httpsprovider"
	"go.opentelemetry.io/collector/confmap/provider/yamlprovider"
	"go.opentelemetry.io/collector/exporter"
	debugexporter "go.opentelemetry.io/collector/exporter/debugexporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	batchprocessor "go.opentelemetry.io/collector/processor/batchprocessor"
	memorylimiterprocessor "go.opentelemetry.io/collector/processor/memorylimiterprocessor"
	"go.opentelemetry.io/collector/receiver"
	otlpreceiver "go.opentelemetry.io/collector/receiver/otlpreceiver"

	azureblobexporter "github.com/fedeoliv/custom-otel-collector/exporter/azureblobexporter"
	azureeventhubsexporter "github.com/fedeoliv/custom-otel-collector/exporter/azureeventhubsexporter"
	"github.com/fedeoliv/custom-otel-collector/processor/trustgatewayprocessor"
	azuremonitorexporter "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/azuremonitorexporter"
	healthcheckextension "github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension"
	probabilisticsamplerprocessor "github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor"
)

func main() {
	info := component.BuildInfo{
		Command:     "otelcol-custom",
		Description: "Custom OpenTelemetry Collector with Trust Gateway",
		Version:     "0.1.0",
	}

	set := otelcol.CollectorSettings{
		BuildInfo: info,
		Factories: func() (otelcol.Factories, error) {
			factories, err := components()
			if err != nil {
				return otelcol.Factories{}, err
			}
			return factories, nil
		},
		ConfigProviderSettings: otelcol.ConfigProviderSettings{
			ResolverSettings: confmap.ResolverSettings{
				URIs: []string{"file:config.yaml"},
				ProviderFactories: []confmap.ProviderFactory{
					fileprovider.NewFactory(),
					envprovider.NewFactory(),
					yamlprovider.NewFactory(),
					httpprovider.NewFactory(),
					httpsprovider.NewFactory(),
				},
			},
		},
	}

	if err := run(set); err != nil {
		log.Fatal(err)
	}
}

func run(params otelcol.CollectorSettings) error {
	cmd := otelcol.NewCommand(params)
	return cmd.Execute()
}

func components() (otelcol.Factories, error) {
	factories := otelcol.Factories{}

	// Receivers
	factories.Receivers = map[component.Type]receiver.Factory{
		otlpreceiver.NewFactory().Type(): otlpreceiver.NewFactory(),
	}

	// Exporters
	factories.Exporters = map[component.Type]exporter.Factory{
		debugexporter.NewFactory().Type():          debugexporter.NewFactory(),
		azuremonitorexporter.NewFactory().Type():   azuremonitorexporter.NewFactory(),
		azureblobexporter.NewFactory().Type():      azureblobexporter.NewFactory(),
		azureeventhubsexporter.NewFactory().Type(): azureeventhubsexporter.NewFactory(),
	}

	// Processors
	factories.Processors = map[component.Type]processor.Factory{
		batchprocessor.NewFactory().Type():                batchprocessor.NewFactory(),
		memorylimiterprocessor.NewFactory().Type():        memorylimiterprocessor.NewFactory(),
		trustgatewayprocessor.NewFactory().Type():         trustgatewayprocessor.NewFactory(),
		probabilisticsamplerprocessor.NewFactory().Type(): probabilisticsamplerprocessor.NewFactory(),
	}

	// Extensions
	factories.Extensions = map[component.Type]extension.Factory{
		healthcheckextension.NewFactory().Type(): healthcheckextension.NewFactory(),
	}

	return factories, nil
}
