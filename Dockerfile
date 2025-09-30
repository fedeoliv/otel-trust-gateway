# Build stage
FROM golang:1.24-alpine AS builder

# Install ca-certificates and git
RUN apk --no-cache add ca-certificates git

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
COPY processor/trustgatewayprocessor/go.mod processor/trustgatewayprocessor/go.sum ./processor/trustgatewayprocessor/

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the collector
RUN CGO_ENABLED=0 GOOS=linux go build -o otelcol-custom .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the collector binary
COPY --from=builder /app/otelcol-custom .
# Copy the config file
COPY --from=builder /app/config.yaml .

# Expose ports
# OTLP gRPC
EXPOSE 4317
# OTLP HTTP
EXPOSE 4318
# Health check
EXPOSE 13133

# Run the collector
ENTRYPOINT ["./otelcol-custom"]
CMD ["--config", "config.yaml"]
