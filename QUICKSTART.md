# Quick Start Guide

This guide will help you get the custom OpenTelemetry Collector and mobile app sample running in minutes.

## Prerequisites

- Go 1.24+ (for building the collector)
- Node.js 20+ (for running the mobile app)
- Docker (optional, for containerization)

## Quick Start (5 minutes)

### Step 1: Build the Collector

```bash
# Navigate to the collector directory
cd src/otel-collector

# Build the collector binary
go build -o otelcol-custom .
```

### Step 2: Start the Collector

```bash
# Run the collector with the default configuration
./otelcol-custom --config config.yaml
```

You should see output indicating the collector is running:
```
Starting otelcol-custom...
Starting GRPC server endpoint=[::]:4317
Starting HTTP server endpoint=[::]:4318
Everything is ready. Begin running and processing data.
```

### Step 3: Run the Mobile App Sample

In a new terminal:

```bash
# Navigate to mobile app directory
cd src/mobile-app

# Install dependencies
npm install

# Run the sample app
npm start
```

The mobile app will:
1. Initialize OpenTelemetry SDK with custom headers
2. Send 5 sample traces and metrics
3. Demonstrate API key validation
4. Shutdown gracefully after 15 seconds

### Step 4: Verify

In the collector terminal, you should see:
- ✅ "Telemetry validation passed" messages
- ✅ Detailed trace and metric data being exported
- ✅ Resource attributes including `X-API-Key` and `X-App-Token`

## Docker Quick Start

```bash
# Navigate to collector directory
cd src/otel-collector

# Build the Docker image
docker build -t otelcol-custom .

# Run the container
docker run -p 4317:4317 -p 4318:4318 -p 13133:13133 otelcol-custom
```

Or use Docker Compose:

```bash
cd src/otel-collector
docker-compose up
```

## Testing Different Scenarios

### Test with Valid Credentials

```bash
cd src/mobile-app
npm start
```

Expected: ✅ Telemetry data is accepted and processed

### Test with Invalid API Key

```bash
cd src/mobile-app
API_KEY=invalid-key npm start
```

Expected: ⚠️ Collector logs show "invalid API key" warnings

### Test with Different Endpoint

```bash
cd src/mobile-app
COLLECTOR_URL=http://your-collector:4318 npm start
```

## Verification Checklist

- [ ] Collector starts without errors
- [ ] HTTP endpoint is accessible: `curl http://localhost:13133` (health check)
- [ ] Mobile app sends telemetry successfully
- [ ] Collector validates and accepts telemetry with valid credentials
- [ ] Collector rejects telemetry with invalid credentials

## Next Steps

1. Read the full [README.md](README.md) for detailed documentation
2. Modify `src/otel-collector/config.yaml` to add more API keys or change validation rules
3. Extend the processor to add custom validation logic
4. Add additional exporters to send data to external systems

## Troubleshooting

### Port Already in Use

If you see "address already in use" errors:

```bash
# Find and kill processes using the ports
lsof -ti:4317 | xargs kill -9
lsof -ti:4318 | xargs kill -9
lsof -ti:8888 | xargs kill -9
```

### Collector Not Receiving Data

1. Check collector is running: `curl http://localhost:13133`
2. Verify mobile app URL: Check `COLLECTOR_URL` environment variable
3. Check firewall rules for ports 4317 and 4318

### Build Errors

```bash
# Clean and rebuild
cd src/otel-collector
rm otelcol-custom
go mod tidy
go build -o otelcol-custom .
```

## Support

For issues, questions, or contributions, please open an issue on GitHub.
