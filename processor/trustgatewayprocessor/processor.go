package trustgatewayprocessor

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

type trustGatewayProcessor struct {
	config *Config
	logger *zap.Logger
}

// processTraces validates traces based on resource attributes
func (p *trustGatewayProcessor) processTraces(ctx context.Context, td ptrace.Traces) (ptrace.Traces, error) {
	if err := p.validateTelemetry(td.ResourceSpans()); err != nil {
		p.logger.Warn("Trace validation failed", zap.Error(err))
		// Return empty traces on validation failure
		return ptrace.NewTraces(), nil
	}
	p.logger.Debug("Trace validation passed", zap.Int("spans", td.SpanCount()))
	return td, nil
}

// processMetrics validates metrics based on resource attributes
func (p *trustGatewayProcessor) processMetrics(ctx context.Context, md pmetric.Metrics) (pmetric.Metrics, error) {
	if err := p.validateTelemetry(md.ResourceMetrics()); err != nil {
		p.logger.Warn("Metric validation failed", zap.Error(err))
		// Return empty metrics on validation failure
		return pmetric.NewMetrics(), nil
	}
	p.logger.Debug("Metric validation passed", zap.Int("datapoints", md.DataPointCount()))
	return md, nil
}

// processLogs validates logs based on resource attributes
func (p *trustGatewayProcessor) processLogs(ctx context.Context, ld plog.Logs) (plog.Logs, error) {
	if err := p.validateTelemetry(ld.ResourceLogs()); err != nil {
		p.logger.Warn("Log validation failed", zap.Error(err))
		// Return empty logs on validation failure
		return plog.NewLogs(), nil
	}
	p.logger.Debug("Log validation passed", zap.Int("records", ld.LogRecordCount()))
	return ld, nil
}

// validateTelemetry checks if the telemetry data contains valid authentication tokens
// The custom headers are expected to be passed as resource attributes by the receiver
func (p *trustGatewayProcessor) validateTelemetry(resources interface{}) error {
	// Check if we have any required headers configured
	if len(p.config.RequiredHeaders) == 0 && len(p.config.ValidAPIKeys) == 0 {
		p.logger.Debug("No validation rules configured, allowing all telemetry")
		return nil
	}

	var attrs map[string]interface{}
	
	// Extract attributes based on telemetry type
	switch r := resources.(type) {
	case ptrace.ResourceSpansSlice:
		if r.Len() == 0 {
			return fmt.Errorf("no resource spans found")
		}
		attrs = extractAttributes(r.At(0).Resource().Attributes())
	case pmetric.ResourceMetricsSlice:
		if r.Len() == 0 {
			return fmt.Errorf("no resource metrics found")
		}
		attrs = extractAttributes(r.At(0).Resource().Attributes())
	case plog.ResourceLogsSlice:
		if r.Len() == 0 {
			return fmt.Errorf("no resource logs found")
		}
		attrs = extractAttributes(r.At(0).Resource().Attributes())
	default:
		return fmt.Errorf("unknown resource type")
	}

	// Validate required headers are present
	for _, header := range p.config.RequiredHeaders {
		if _, ok := attrs[header]; !ok {
			return fmt.Errorf("missing required header: %s", header)
		}
	}

	// Validate API key if configured
	if len(p.config.ValidAPIKeys) > 0 {
		apiKey, ok := attrs["X-API-Key"]
		if !ok {
			return fmt.Errorf("missing X-API-Key header")
		}
		
		apiKeyStr, ok := apiKey.(string)
		if !ok {
			return fmt.Errorf("X-API-Key must be a string")
		}

		valid := false
		for _, validKey := range p.config.ValidAPIKeys {
			if apiKeyStr == validKey {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid API key")
		}
	}

	return nil
}

// extractAttributes converts pcommon.Map to a regular map
func extractAttributes(attrs interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	// Type assertion to get the actual attributes map
	if attrMap, ok := attrs.(interface{ Range(func(string, interface{}) bool) }); ok {
		attrMap.Range(func(k string, v interface{}) bool {
			result[k] = v
			return true
		})
	}
	
	return result
}
