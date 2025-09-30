const { NodeSDK } = require('@opentelemetry/sdk-node');
const { OTLPTraceExporter } = require('@opentelemetry/exporter-trace-otlp-http');
const { OTLPMetricExporter } = require('@opentelemetry/exporter-metrics-otlp-http');
const { PeriodicExportingMetricReader } = require('@opentelemetry/sdk-metrics');
const { Resource } = require('@opentelemetry/resources');
const { ATTR_SERVICE_NAME, ATTR_SERVICE_VERSION } = require('@opentelemetry/semantic-conventions');
const opentelemetry = require('@opentelemetry/api');

// Configuration
const COLLECTOR_URL = process.env.COLLECTOR_URL || 'http://localhost:4318';
const API_KEY = process.env.API_KEY || 'mobile-app-secret-key-123';
const APP_TOKEN = process.env.APP_TOKEN || 'my-mobile-app-token';

// Custom headers to be sent with every request
const customHeaders = {
  'X-API-Key': API_KEY,
  'X-App-Token': APP_TOKEN,
};

console.log('Starting Mobile App with OTel SDK');
console.log(`Collector URL: ${COLLECTOR_URL}`);
console.log(`Using API Key: ${API_KEY}`);
console.log(`Using App Token: ${APP_TOKEN}`);

// Create resource with custom attributes
// These attributes will be added to all telemetry data
const resource = new Resource({
  [ATTR_SERVICE_NAME]: 'mobile-app-sample',
  [ATTR_SERVICE_VERSION]: '1.0.0',
  'deployment.environment': 'development',
  'device.platform': 'mobile',
  // Pass the custom headers as resource attributes
  // so they are available to the processor
  'X-API-Key': API_KEY,
  'X-App-Token': APP_TOKEN,
});

// Initialize trace exporter with custom headers
const traceExporter = new OTLPTraceExporter({
  url: `${COLLECTOR_URL}/v1/traces`,
  headers: customHeaders,
});

// Initialize metrics exporter with custom headers
const metricExporter = new OTLPMetricExporter({
  url: `${COLLECTOR_URL}/v1/metrics`,
  headers: customHeaders,
});

// Create a metric reader
const metricReader = new PeriodicExportingMetricReader({
  exporter: metricExporter,
  exportIntervalMillis: 5000,
});

// Initialize the SDK
const sdk = new NodeSDK({
  resource: resource,
  traceExporter: traceExporter,
  metricReader: metricReader,
});

// Start the SDK
sdk.start();

console.log('OpenTelemetry SDK started');

// Graceful shutdown
process.on('SIGTERM', () => {
  sdk.shutdown()
    .then(() => console.log('OpenTelemetry SDK shut down successfully'))
    .catch((error) => console.error('Error shutting down OpenTelemetry SDK', error))
    .finally(() => process.exit(0));
});

// ====================================
// Sample application code starts here
// ====================================

const tracer = opentelemetry.trace.getTracer('mobile-app-tracer');
const meter = opentelemetry.metrics.getMeter('mobile-app-meter');

// Create a counter metric
const requestCounter = meter.createCounter('app.requests', {
  description: 'Count of app requests',
});

// Create a histogram metric
const requestDuration = meter.createHistogram('app.request.duration', {
  description: 'Duration of app requests in ms',
  unit: 'ms',
});

// Simulate mobile app activity
function simulateActivity() {
  const span = tracer.startSpan('mobile-app-activity');
  
  try {
    span.setAttribute('user.action', 'button_click');
    span.setAttribute('screen.name', 'home');
    span.addEvent('Activity started');
    
    // Simulate some work
    const duration = Math.random() * 1000;
    const startTime = Date.now();
    
    setTimeout(() => {
      const actualDuration = Date.now() - startTime;
      
      // Record metrics
      requestCounter.add(1, {
        'action.type': 'button_click',
        'status': 'success',
      });
      
      requestDuration.record(actualDuration, {
        'action.type': 'button_click',
      });
      
      span.addEvent('Activity completed');
      span.end();
      
      console.log(`Activity completed in ${actualDuration}ms`);
    }, duration);
    
  } catch (error) {
    span.recordException(error);
    span.setStatus({ code: opentelemetry.SpanStatusCode.ERROR });
    span.end();
  }
}

// Run sample activities
console.log('\nSending sample telemetry data...\n');

// Send multiple activities
for (let i = 0; i < 5; i++) {
  setTimeout(() => {
    console.log(`\n[Activity ${i + 1}] Starting...`);
    simulateActivity();
  }, i * 2000);
}

// Test with invalid API key after some time
setTimeout(() => {
  console.log('\n\n=== Testing with INVALID API Key ===\n');
  
  // Create a new exporter with invalid credentials
  const invalidTraceExporter = new OTLPTraceExporter({
    url: `${COLLECTOR_URL}/v1/traces`,
    headers: {
      'X-API-Key': 'invalid-key',
      'X-App-Token': APP_TOKEN,
    },
  });
  
  // Note: In a real app, you would need to reconfigure the SDK
  // This is just to demonstrate that invalid keys should be rejected
  console.log('Attempting to send trace with invalid API key...');
  console.log('(This would be rejected by the trust gateway processor)\n');
  
}, 12000);

// Keep the app running for a bit, then exit
setTimeout(() => {
  console.log('\n\nShutting down mobile app...');
  sdk.shutdown()
    .then(() => {
      console.log('Mobile app stopped successfully');
      process.exit(0);
    })
    .catch((error) => {
      console.error('Error during shutdown:', error);
      process.exit(1);
    });
}, 15000);
