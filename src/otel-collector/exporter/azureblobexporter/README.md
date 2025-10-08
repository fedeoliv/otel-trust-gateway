# Custom Azure Blob Exporter

This is a custom version of the [Azure Blob exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/azureblobexporter) with additional support for **DefaultAzureCredential** authentication.

## Differences from Upstream

The primary enhancement in this custom exporter is the addition of the `default_credentials` authentication type, which leverages Azure's `DefaultAzureCredential` for automatic credential discovery.

### New Authentication Type: `default_credentials`

This authentication method automatically attempts to authenticate using multiple credential sources in the following order:

1. **Environment Variables** - `AZURE_TENANT_ID`, `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`, etc.
2. **Workload Identity** - For Kubernetes environments
3. **Managed Identity** - System or user-assigned managed identities
4. **Azure CLI** - Credentials from `az login`
5. **Azure PowerShell** - Credentials from `Connect-AzAccount`

This follows the same pattern implemented in the Azure Monitor exporter (PR [#33584](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/33584)).

## Configuration

### Example with Default Credentials

```yaml
exporters:
  azureblob:
    url: "https://<your-storage-account>.blob.core.windows.net"
    auth:
      type: default_credentials
    container:
      metrics: "metrics"
      logs: "logs"
      traces: "traces"
```

### All Supported Authentication Types

The exporter supports the following authentication methods:

1. **Connection String** (default)
   ```yaml
   auth:
     type: connection_string
     connection_string: "DefaultEndpointsProtocol=https;AccountName=..."
   ```

2. **Service Principal**
   ```yaml
   auth:
     type: service_principal
     tenant_id: "your-tenant-id"
     client_id: "your-client-id"
     client_secret: "your-client-secret"
   ```

3. **System Managed Identity**
   ```yaml
   auth:
     type: system_managed_identity
   ```

4. **User Managed Identity**
   ```yaml
   auth:
     type: user_managed_identity
     client_id: "your-managed-identity-client-id"
   ```

5. **Workload Identity**
   ```yaml
   auth:
     type: workload_identity
     tenant_id: "your-tenant-id"
     client_id: "your-client-id"
   ```

6. **Default Credentials** ⭐ (NEW)
   ```yaml
   auth:
     type: default_credentials
   ```

## Complete Configuration Example

```yaml
exporters:
  azureblob:
    url: "https://mystorageaccount.blob.core.windows.net"
    auth:
      type: default_credentials
    container:
      metrics: "otel-metrics"
      logs: "otel-logs"
      traces: "otel-traces"
    blob_name_format:
      metrics_format: "2006/01/02/metrics_15_04_05.json"
      logs_format: "2006/01/02/logs_15_04_05.json"
      traces_format: "2006/01/02/traces_15_04_05.json"
      serial_num_range: 10000
      template_enabled: false
    format_type: "json"
    append_blob:
      enabled: false
      separator: "\n"
```

## Benefits of Default Credentials

- **Simplified Configuration**: No need to specify explicit credentials in configuration files
- **Enhanced Security**: Credentials are never stored in configuration or code
- **Environment Flexibility**: Automatically adapts to different Azure environments (local dev, AKS, VM with managed identity, etc.)
- **Best Practice**: Follows Azure SDK recommendations for credential management

## Use Cases

### Local Development
Set Azure CLI credentials:
```bash
az login
```

### Kubernetes/AKS
Use workload identity or managed identity - no configuration needed.

### Azure VM
Use system or user-assigned managed identity - no configuration needed.

### CI/CD Pipeline
Set service principal credentials via environment variables:
```bash
export AZURE_TENANT_ID="..."
export AZURE_CLIENT_ID="..."
export AZURE_CLIENT_SECRET="..."
```

## Contributing Back

This implementation is designed to be contributed back to the upstream OpenTelemetry Collector Contrib repository. The changes follow the pattern established by the Azure Monitor exporter and maintain compatibility with all existing authentication methods.
