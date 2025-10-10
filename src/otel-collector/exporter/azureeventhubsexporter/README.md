# Azure Event Hubs Exporter

The Azure Event Hubs exporter exports traces, metrics, and logs to Azure Event Hubs.

## Configuration

The following configuration options are supported:

- **namespace** (required if not using connection_string): The Event Hubs namespace endpoint (e.g., "my-namespace.servicebus.windows.net")
- **event_hub**: Event Hub names for different telemetry types
  - **logs** (default: "logs"): Event Hub name for logs
  - **metrics** (default: "metrics"): Event Hub name for metrics  
  - **traces** (default: "traces"): Event Hub name for traces
- **auth**: Authentication configuration
  - **type**: Authentication type. Supported values: `connection_string`, `service_principal`, `system_managed_identity`, `user_managed_identity`, `workload_identity`, `default_credentials`
  - **connection_string**: Connection string to the Event Hubs namespace or Event Hub (required when type is `connection_string`)
  - **tenant_id**: Tenant ID for Azure AD authentication (required for `service_principal` and `workload_identity`)
  - **client_id**: Client ID (required for `service_principal`, `user_managed_identity`, and `workload_identity`)
  - **client_secret**: Client secret (required for `service_principal`)
  - **federated_token_file**: Path to federated token file (required for `workload_identity`)
- **format** (default: "json"): Format of encoded telemetry data. Supported values: `json`, `proto`
- **partition_key**: Partition key configuration for Event Hub partitioning
  - **source**: How the partition key is generated. Options: `static`, `resource_attribute`, `trace_id`, `span_id`, `random`
  - **value**: Used when source is `static` or specifies the attribute name when source is `resource_attribute`
- **max_event_size** (default: 1048576): Maximum size of an event in bytes (max: 1MB for Event Hubs)
- **batch_size** (default: 100): Number of events to batch before sending
- **retry_on_failure**: Retry configuration
  - **enabled** (default: true): Whether to retry on failure
  - **initial_interval** (default: 5s): Initial retry interval
  - **max_interval** (default: 30s): Maximum retry interval
  - **max_elapsed_time** (default: 5m): Maximum elapsed time for retries

## Example Configuration

### Using Connection String

```yaml
exporters:
  azureeventhubs:
    auth:
      type: connection_string
      connection_string: "Endpoint=sb://my-namespace.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=my-key"
    event_hub:
      logs: "otel-logs"
      metrics: "otel-metrics"
      traces: "otel-traces"
    format: json
    partition_key:
      source: trace_id
    max_event_size: 1048576
    batch_size: 100
```

### Using Service Principal

```yaml
exporters:
  azureeventhubs:
    namespace: "my-namespace.servicebus.windows.net"
    auth:
      type: service_principal
      tenant_id: "your-tenant-id"
      client_id: "your-client-id"
      client_secret: "your-client-secret"
    event_hub:
      logs: "otel-logs"
      metrics: "otel-metrics"
      traces: "otel-traces"
    format: proto
    partition_key:
      source: resource_attribute
      value: "service.name"
```

### Using Managed Identity

```yaml
exporters:
  azureeventhubs:
    namespace: "my-namespace.servicebus.windows.net"
    auth:
      type: system_managed_identity
    event_hub:
      logs: "otel-logs"
      metrics: "otel-metrics"
      traces: "otel-traces"
    format: json
    partition_key:
      source: random
```

## Partition Key Strategies

The exporter supports different partition key strategies:

- **static**: Uses a fixed partition key value
- **resource_attribute**: Uses the value of a specified resource attribute
- **trace_id**: Uses the trace ID from the telemetry data (traces and logs only)
- **span_id**: Uses the span ID from the telemetry data (traces only)
- **random**: Generates a random partition key for even distribution

## Authentication Types

- **connection_string**: Uses a connection string to authenticate
- **service_principal**: Uses Azure AD service principal authentication
- **system_managed_identity**: Uses system-assigned managed identity
- **user_managed_identity**: Uses user-assigned managed identity
- **workload_identity**: Uses workload identity (for Kubernetes environments)
- **default_credentials**: Uses DefaultAzureCredential which tries multiple authentication methods

## Event Hub Requirements

- Event Hubs must exist before starting the collector
- The authentication principal must have "Azure Event Hubs Data Sender" role on the Event Hubs
- Maximum event size is 1MB for Event Hubs Standard tier

## Performance Considerations

- Use appropriate batch sizes to optimize throughput
- Consider partition key strategy for load distribution
- Monitor Event Hub metrics for throttling and errors
- Use connection pooling by sharing Event Hub clients when possible