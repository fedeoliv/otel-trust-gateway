// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package azureblobexporter

import (
	"bytes"
	"fmt"

	"github.com/parquet-go/parquet-go"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// Parquet schema structs for OpenTelemetry data

// ParquetSpan represents a trace span in Parquet format
type ParquetSpan struct {
	TraceID            string            `parquet:"trace_id"`
	SpanID             string            `parquet:"span_id"`
	ParentSpanID       string            `parquet:"parent_span_id,optional"`
	Name               string            `parquet:"name"`
	Kind               int32             `parquet:"kind"`
	StartTimeUnixNano  int64             `parquet:"start_time_unix_nano"`
	EndTimeUnixNano    int64             `parquet:"end_time_unix_nano"`
	StatusCode         int32             `parquet:"status_code"`
	StatusMessage      string            `parquet:"status_message,optional"`
	ResourceAttributes map[string]string `parquet:"resource_attributes,optional"`
	SpanAttributes     map[string]string `parquet:"span_attributes,optional"`
	ScopeName          string            `parquet:"scope_name,optional"`
	ScopeVersion       string            `parquet:"scope_version,optional"`
}

// ParquetLog represents a log record in Parquet format
type ParquetLog struct {
	Timestamp          int64             `parquet:"timestamp_unix_nano"`
	ObservedTimestamp  int64             `parquet:"observed_timestamp_unix_nano"`
	SeverityNumber     int32             `parquet:"severity_number"`
	SeverityText       string            `parquet:"severity_text,optional"`
	Body               string            `parquet:"body"`
	TraceID            string            `parquet:"trace_id,optional"`
	SpanID             string            `parquet:"span_id,optional"`
	Flags              uint32            `parquet:"flags"`
	ResourceAttributes map[string]string `parquet:"resource_attributes,optional"`
	LogAttributes      map[string]string `parquet:"log_attributes,optional"`
	ScopeName          string            `parquet:"scope_name,optional"`
	ScopeVersion       string            `parquet:"scope_version,optional"`
}

// ParquetMetric represents a metric data point in Parquet format
type ParquetMetric struct {
	Name               string            `parquet:"name"`
	Description        string            `parquet:"description,optional"`
	Unit               string            `parquet:"unit,optional"`
	Type               string            `parquet:"type"` // gauge, sum, histogram, etc.
	TimeUnixNano       int64             `parquet:"time_unix_nano"`
	ValueType          string            `parquet:"value_type"` // int, double
	IntValue           int64             `parquet:"int_value,optional"`
	DoubleValue        float64           `parquet:"double_value,optional"`
	ResourceAttributes map[string]string `parquet:"resource_attributes,optional"`
	MetricAttributes   map[string]string `parquet:"metric_attributes,optional"`
	ScopeName          string            `parquet:"scope_name,optional"`
	ScopeVersion       string            `parquet:"scope_version,optional"`
	// For Sum metrics
	IsMonotonic            bool   `parquet:"is_monotonic,optional"`
	AggregationTemporality string `parquet:"aggregation_temporality,optional"`
	StartTimeUnixNano      int64  `parquet:"start_time_unix_nano,optional"`
}

type parquetMarshaller struct{}

func newParquetMarshaller() *parquetMarshaller {
	return &parquetMarshaller{}
}

func (p *parquetMarshaller) MarshalTraces(td ptrace.Traces) ([]byte, error) {
	var spans []ParquetSpan

	for i := 0; i < td.ResourceSpans().Len(); i++ {
		rs := td.ResourceSpans().At(i)
		resourceAttrs := attributesToMap(rs.Resource().Attributes())

		for j := 0; j < rs.ScopeSpans().Len(); j++ {
			ss := rs.ScopeSpans().At(j)
			scopeName := ss.Scope().Name()
			scopeVersion := ss.Scope().Version()

			for k := 0; k < ss.Spans().Len(); k++ {
				span := ss.Spans().At(k)

				parentSpanID := ""
				if !span.ParentSpanID().IsEmpty() {
					parentSpanID = span.ParentSpanID().String()
				}

				parquetSpan := ParquetSpan{
					TraceID:            span.TraceID().String(),
					SpanID:             span.SpanID().String(),
					ParentSpanID:       parentSpanID,
					Name:               span.Name(),
					Kind:               int32(span.Kind()),
					StartTimeUnixNano:  int64(span.StartTimestamp()),
					EndTimeUnixNano:    int64(span.EndTimestamp()),
					StatusCode:         int32(span.Status().Code()),
					StatusMessage:      span.Status().Message(),
					ResourceAttributes: resourceAttrs,
					SpanAttributes:     attributesToMap(span.Attributes()),
					ScopeName:          scopeName,
					ScopeVersion:       scopeVersion,
				}
				spans = append(spans, parquetSpan)
			}
		}
	}

	return marshalToParquet(spans)
}

func (p *parquetMarshaller) MarshalLogs(ld plog.Logs) ([]byte, error) {
	var logs []ParquetLog

	for i := 0; i < ld.ResourceLogs().Len(); i++ {
		rl := ld.ResourceLogs().At(i)
		resourceAttrs := attributesToMap(rl.Resource().Attributes())

		for j := 0; j < rl.ScopeLogs().Len(); j++ {
			sl := rl.ScopeLogs().At(j)
			scopeName := sl.Scope().Name()
			scopeVersion := sl.Scope().Version()

			for k := 0; k < sl.LogRecords().Len(); k++ {
				logRecord := sl.LogRecords().At(k)

				traceID := ""
				if !logRecord.TraceID().IsEmpty() {
					traceID = logRecord.TraceID().String()
				}

				spanID := ""
				if !logRecord.SpanID().IsEmpty() {
					spanID = logRecord.SpanID().String()
				}

				parquetLog := ParquetLog{
					Timestamp:          int64(logRecord.Timestamp()),
					ObservedTimestamp:  int64(logRecord.ObservedTimestamp()),
					SeverityNumber:     int32(logRecord.SeverityNumber()),
					SeverityText:       logRecord.SeverityText(),
					Body:               logRecord.Body().AsString(),
					TraceID:            traceID,
					SpanID:             spanID,
					Flags:              uint32(logRecord.Flags()),
					ResourceAttributes: resourceAttrs,
					LogAttributes:      attributesToMap(logRecord.Attributes()),
					ScopeName:          scopeName,
					ScopeVersion:       scopeVersion,
				}
				logs = append(logs, parquetLog)
			}
		}
	}

	return marshalToParquet(logs)
}

func (p *parquetMarshaller) MarshalMetrics(md pmetric.Metrics) ([]byte, error) {
	var metrics []ParquetMetric

	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rm := md.ResourceMetrics().At(i)
		resourceAttrs := attributesToMap(rm.Resource().Attributes())

		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)
			scopeName := sm.Scope().Name()
			scopeVersion := sm.Scope().Version()

			for k := 0; k < sm.Metrics().Len(); k++ {
				metric := sm.Metrics().At(k)

				// Handle different metric types
				switch metric.Type() {
				case pmetric.MetricTypeGauge:
					metrics = append(metrics, extractGaugeMetrics(metric, resourceAttrs, scopeName, scopeVersion)...)
				case pmetric.MetricTypeSum:
					metrics = append(metrics, extractSumMetrics(metric, resourceAttrs, scopeName, scopeVersion)...)
				case pmetric.MetricTypeHistogram:
					metrics = append(metrics, extractHistogramMetrics(metric, resourceAttrs, scopeName, scopeVersion)...)
				case pmetric.MetricTypeSummary:
					metrics = append(metrics, extractSummaryMetrics(metric, resourceAttrs, scopeName, scopeVersion)...)
				case pmetric.MetricTypeExponentialHistogram:
					metrics = append(metrics, extractExponentialHistogramMetrics(metric, resourceAttrs, scopeName, scopeVersion)...)
				}
			}
		}
	}

	return marshalToParquet(metrics)
}

func (p *parquetMarshaller) format() string {
	return formatTypeParquet
}

// Helper functions

func attributesToMap(attrs pcommon.Map) map[string]string {
	result := make(map[string]string, attrs.Len())
	attrs.Range(func(k string, v pcommon.Value) bool {
		result[k] = v.AsString()
		return true
	})
	return result
}

func extractGaugeMetrics(metric pmetric.Metric, resourceAttrs map[string]string, scopeName, scopeVersion string) []ParquetMetric {
	var metrics []ParquetMetric
	gauge := metric.Gauge()

	for i := 0; i < gauge.DataPoints().Len(); i++ {
		dp := gauge.DataPoints().At(i)
		pm := ParquetMetric{
			Name:               metric.Name(),
			Description:        metric.Description(),
			Unit:               metric.Unit(),
			Type:               "gauge",
			TimeUnixNano:       int64(dp.Timestamp()),
			ResourceAttributes: resourceAttrs,
			MetricAttributes:   attributesToMap(dp.Attributes()),
			ScopeName:          scopeName,
			ScopeVersion:       scopeVersion,
		}

		switch dp.ValueType() {
		case pmetric.NumberDataPointValueTypeInt:
			pm.ValueType = "int"
			pm.IntValue = dp.IntValue()
		case pmetric.NumberDataPointValueTypeDouble:
			pm.ValueType = "double"
			pm.DoubleValue = dp.DoubleValue()
		}

		metrics = append(metrics, pm)
	}

	return metrics
}

func extractSumMetrics(metric pmetric.Metric, resourceAttrs map[string]string, scopeName, scopeVersion string) []ParquetMetric {
	var metrics []ParquetMetric
	sum := metric.Sum()

	aggregationTemporality := "unspecified"
	switch sum.AggregationTemporality() {
	case pmetric.AggregationTemporalityDelta:
		aggregationTemporality = "delta"
	case pmetric.AggregationTemporalityCumulative:
		aggregationTemporality = "cumulative"
	}

	for i := 0; i < sum.DataPoints().Len(); i++ {
		dp := sum.DataPoints().At(i)
		pm := ParquetMetric{
			Name:                   metric.Name(),
			Description:            metric.Description(),
			Unit:                   metric.Unit(),
			Type:                   "sum",
			TimeUnixNano:           int64(dp.Timestamp()),
			StartTimeUnixNano:      int64(dp.StartTimestamp()),
			ResourceAttributes:     resourceAttrs,
			MetricAttributes:       attributesToMap(dp.Attributes()),
			ScopeName:              scopeName,
			ScopeVersion:           scopeVersion,
			IsMonotonic:            sum.IsMonotonic(),
			AggregationTemporality: aggregationTemporality,
		}

		switch dp.ValueType() {
		case pmetric.NumberDataPointValueTypeInt:
			pm.ValueType = "int"
			pm.IntValue = dp.IntValue()
		case pmetric.NumberDataPointValueTypeDouble:
			pm.ValueType = "double"
			pm.DoubleValue = dp.DoubleValue()
		}

		metrics = append(metrics, pm)
	}

	return metrics
}

func extractHistogramMetrics(metric pmetric.Metric, resourceAttrs map[string]string, scopeName, scopeVersion string) []ParquetMetric {
	var metrics []ParquetMetric
	histogram := metric.Histogram()

	aggregationTemporality := "unspecified"
	switch histogram.AggregationTemporality() {
	case pmetric.AggregationTemporalityDelta:
		aggregationTemporality = "delta"
	case pmetric.AggregationTemporalityCumulative:
		aggregationTemporality = "cumulative"
	}

	for i := 0; i < histogram.DataPoints().Len(); i++ {
		dp := histogram.DataPoints().At(i)

		// For histograms, we store the sum as the primary value
		pm := ParquetMetric{
			Name:                   metric.Name(),
			Description:            metric.Description(),
			Unit:                   metric.Unit(),
			Type:                   "histogram",
			TimeUnixNano:           int64(dp.Timestamp()),
			StartTimeUnixNano:      int64(dp.StartTimestamp()),
			ValueType:              "double",
			DoubleValue:            dp.Sum(),
			ResourceAttributes:     resourceAttrs,
			MetricAttributes:       attributesToMap(dp.Attributes()),
			ScopeName:              scopeName,
			ScopeVersion:           scopeVersion,
			AggregationTemporality: aggregationTemporality,
		}

		metrics = append(metrics, pm)
	}

	return metrics
}

func extractSummaryMetrics(metric pmetric.Metric, resourceAttrs map[string]string, scopeName, scopeVersion string) []ParquetMetric {
	var metrics []ParquetMetric
	summary := metric.Summary()

	for i := 0; i < summary.DataPoints().Len(); i++ {
		dp := summary.DataPoints().At(i)

		// Store summary sum as the primary value
		pm := ParquetMetric{
			Name:               metric.Name(),
			Description:        metric.Description(),
			Unit:               metric.Unit(),
			Type:               "summary",
			TimeUnixNano:       int64(dp.Timestamp()),
			StartTimeUnixNano:  int64(dp.StartTimestamp()),
			ValueType:          "double",
			DoubleValue:        dp.Sum(),
			ResourceAttributes: resourceAttrs,
			MetricAttributes:   attributesToMap(dp.Attributes()),
			ScopeName:          scopeName,
			ScopeVersion:       scopeVersion,
		}

		metrics = append(metrics, pm)
	}

	return metrics
}

func extractExponentialHistogramMetrics(metric pmetric.Metric, resourceAttrs map[string]string, scopeName, scopeVersion string) []ParquetMetric {
	var metrics []ParquetMetric
	expHistogram := metric.ExponentialHistogram()

	aggregationTemporality := "unspecified"
	switch expHistogram.AggregationTemporality() {
	case pmetric.AggregationTemporalityDelta:
		aggregationTemporality = "delta"
	case pmetric.AggregationTemporalityCumulative:
		aggregationTemporality = "cumulative"
	}

	for i := 0; i < expHistogram.DataPoints().Len(); i++ {
		dp := expHistogram.DataPoints().At(i)

		pm := ParquetMetric{
			Name:                   metric.Name(),
			Description:            metric.Description(),
			Unit:                   metric.Unit(),
			Type:                   "exponential_histogram",
			TimeUnixNano:           int64(dp.Timestamp()),
			StartTimeUnixNano:      int64(dp.StartTimestamp()),
			ValueType:              "double",
			DoubleValue:            dp.Sum(),
			ResourceAttributes:     resourceAttrs,
			MetricAttributes:       attributesToMap(dp.Attributes()),
			ScopeName:              scopeName,
			ScopeVersion:           scopeVersion,
			AggregationTemporality: aggregationTemporality,
		}

		metrics = append(metrics, pm)
	}

	return metrics
}

func marshalToParquet[T any](rows []T) ([]byte, error) {
	if len(rows) == 0 {
		return []byte{}, nil
	}

	buf := new(bytes.Buffer)
	// Create writer with Snappy compression
	writer := parquet.NewGenericWriter[T](buf, parquet.Compression(&parquet.Snappy))

	_, err := writer.Write(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to write parquet data: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close parquet writer: %w", err)
	}

	return buf.Bytes(), nil
}
