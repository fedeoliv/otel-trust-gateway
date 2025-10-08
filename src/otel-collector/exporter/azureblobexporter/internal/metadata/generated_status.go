// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package metadata

import (
	"go.opentelemetry.io/collector/component"
)

var (
	Type      = component.MustNewType("azureblob")
	ScopeName = "github.com/fedeoliv/custom-otel-collector/exporter/azureblobexporter"
)

const (
	TracesStability  = component.StabilityLevelBeta
	MetricsStability = component.StabilityLevelBeta
	LogsStability    = component.StabilityLevelBeta
)
