/**
 * Mobile App Telemetry Simulator
 * 
 * This application simulates a high-volume mobile app generating OpenTelemetry data.
 * 
 * Simulation Details:
 * - Generates 1000 activities at 10 activities/second (100 seconds total)
 * - 6 activity types: button_click, p// Run sample activities
 * - 7 different screens: home, profile, settings, checkout, products, cart, login
 * - 50 simulated users (user-0 to user-49)
 * - 75% success rate, 25% error rate
 * - Variable duration: 0-500ms per activity
 * 
 * Telemetry Sent:
 * - Traces: 1000 spans with attributes (user.action, screen.name, user.id, etc.)
 * - Logs: 2000-3000 log records WITH TRACE CONTEXT (trace_id, span_id)
 *   - Each activity generates 2-3 logs: start, complete/error
 *   - Logs are correlated with their parent trace via trace_id
 * - Metrics: Counters and histograms for request counts and durations
 * 
 * Sampling & Correlation:
 * With 10% probabilistic sampling configured in the collector:
 * - ~100 traces will be sent to Azure Application Insights
 * - ~200-300 logs will be sent (same trace_ids as the sampled traces)
 * - Logs and traces are CORRELATED - same trace_id means they're kept together
 * - 100% of metrics are sent (no sampling)
 * - When other exporters are enabled, they receive 100% of all data
 */

const { NodeSDK } = require('@opentelemetry/sdk-node');
const { OTLPTraceExporter } = require('@opentelemetry/exporter-trace-otlp-http');
const { OTLPMetricExporter } = require('@opentelemetry/exporter-metrics-otlp-http');
const { OTLPLogExporter } = require('@opentelemetry/exporter-logs-otlp-http');
const { PeriodicExportingMetricReader } = require('@opentelemetry/sdk-metrics');
const { LoggerProvider, BatchLogRecordProcessor } = require('@opentelemetry/sdk-logs');
const { logs } = require('@opentelemetry/api-logs');
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

// Initialize logs exporter with custom headers
const logExporter = new OTLPLogExporter({
  url: `${COLLECTOR_URL}/v1/logs`,
  headers: customHeaders,
});

// Create logger provider with batch processor
const loggerProvider = new LoggerProvider({ resource });
loggerProvider.addLogRecordProcessor(new BatchLogRecordProcessor(logExporter));
logs.setGlobalLoggerProvider(loggerProvider);

// Initialize the SDK
const sdk = new NodeSDK({
  resource: resource,
  traceExporter: traceExporter,
  metricReader: metricReader,
  logRecordProcessor: new BatchLogRecordProcessor(logExporter),
});

// Start the SDK
sdk.start();

console.log('OpenTelemetry SDK started (traces, metrics, and logs)');

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
const logger = logs.getLogger('mobile-app-logger');

// Create a counter metric
const requestCounter = meter.createCounter('app.requests', {
  description: 'Count of app requests',
});

// Create a histogram metric
const requestDuration = meter.createHistogram('app.request.duration', {
  description: 'Duration of app requests in ms',
  unit: 'ms',
});

// Simulate mobile app activity with variety
function simulateActivity(activityNumber) {
  // Randomly select activity type and screen
  const activityTypes = ['button_click', 'page_view', 'api_call', 'user_input', 'scroll', 'swipe'];
  const screens = ['home', 'profile', 'settings', 'checkout', 'products', 'cart', 'login'];
  const statuses = ['success', 'success', 'success', 'error']; // 75% success, 25% error
  
  const activityType = activityTypes[Math.floor(Math.random() * activityTypes.length)];
  const screen = screens[Math.floor(Math.random() * screens.length)];
  const status = statuses[Math.floor(Math.random() * statuses.length)];
  
  const span = tracer.startSpan(`mobile-app-${activityType}`);
  const userId = `user-${Math.floor(Math.random() * 50)}`;
  
  // Create context with the active span so logs inherit trace_id and span_id
  const ctx = opentelemetry.trace.setSpan(opentelemetry.context.active(), span);
  
  try {
    span.setAttribute('user.action', activityType);
    span.setAttribute('screen.name', screen);
    span.setAttribute('activity.number', activityNumber);
    span.setAttribute('user.id', userId);
    span.addEvent('Activity started');
    
    // Emit log WITH trace context - this log will have the same trace_id as the span
    opentelemetry.context.with(ctx, () => {
      logger.emit({
        severityText: 'INFO',
        body: `Activity started: ${activityType} on ${screen}`,
        attributes: {
          'log.type': 'activity_start',
          'user.action': activityType,
          'screen.name': screen,
          'user.id': userId,
          'activity.number': activityNumber,
        },
      });
    });
    
    // Simulate some work with varying duration
    const duration = Math.random() * 500; // 0-500ms
    const startTime = Date.now();
    
    setTimeout(() => {
      const actualDuration = Date.now() - startTime;
      
      // Simulate occasional errors
      if (status === 'error') {
        const error = new Error(`Simulated error in ${activityType}`);
        span.recordException(error);
        span.setStatus({ code: opentelemetry.SpanStatusCode.ERROR });
        
        // Emit ERROR log WITH trace context
        opentelemetry.context.with(ctx, () => {
          logger.emit({
            severityText: 'ERROR',
            body: `Error in ${activityType}: ${error.message}`,
            attributes: {
              'log.type': 'error',
              'error.message': error.message,
              'user.action': activityType,
              'screen.name': screen,
              'user.id': userId,
            },
          });
        });
      } else {
        span.setStatus({ code: opentelemetry.SpanStatusCode.OK });
        
        // Emit SUCCESS log WITH trace context
        opentelemetry.context.with(ctx, () => {
          logger.emit({
            severityText: 'INFO',
            body: `Activity completed: ${activityType} on ${screen} (${actualDuration.toFixed(0)}ms)`,
            attributes: {
              'log.type': 'activity_complete',
              'user.action': activityType,
              'screen.name': screen,
              'user.id': userId,
              'duration.ms': actualDuration,
              'status': status,
            },
          });
        });
      }
      
      // Record metrics
      requestCounter.add(1, {
        'action.type': activityType,
        'status': status,
        'screen': screen,
      });
      
      requestDuration.record(actualDuration, {
        'action.type': activityType,
        'screen': screen,
      });
      
      span.addEvent('Activity completed', { status });
      span.end();
      
      if (activityNumber % 100 === 0) {
        console.log(`[${activityNumber}/1000] ${activityType} on ${screen} - ${status} (${actualDuration.toFixed(0)}ms)`);
      }
    }, duration);
    
  } catch (error) {
    span.recordException(error);
    span.setStatus({ code: opentelemetry.SpanStatusCode.ERROR });
    
    // Emit ERROR log WITH trace context
    opentelemetry.context.with(ctx, () => {
      logger.emit({
        severityText: 'ERROR',
        body: `Exception in activity: ${error.message}`,
        attributes: {
          'log.type': 'exception',
          'error.message': error.message,
          'user.action': activityType,
          'screen.name': screen,
        },
      });
    });
    
    span.end();
  }
}

// Run sample activities
console.log('\n========================================');
console.log('Generating HIGH VOLUME telemetry data');
console.log('Sending 1000 activities with random types');
console.log('========================================\n');

// Generate 1000 activities rapidly to simulate high volume
console.log('Starting activity generation...\n');

for (let i = 1; i <= 1000; i++) {
  // Stagger the activities slightly to avoid overwhelming the system
  // But keep them close together to simulate high volume
  setTimeout(() => {
    simulateActivity(i);
  }, i * 100); // 100ms between each activity = 10 activities/second
}

// Show progress
setTimeout(() => {
  console.log('\n--- 25% complete (250 activities) ---');
}, 25000);

setTimeout(() => {
  console.log('--- 50% complete (500 activities) ---');
}, 50000);

setTimeout(() => {
  console.log('--- 75% complete (750 activities) ---');
}, 75000);

setTimeout(() => {
  console.log('--- 100% complete ---\n');
  console.log('All 1000 activities dispatched!');
  console.log('\nNote: With 10% sampling, approximately 100 traces will be sent to Azure Monitor.');
  console.log('Check your Application Insights in a few moments to see the sampled data.\n');
}, 100000);

// Keep the app running to allow all telemetry to be exported
setTimeout(() => {
  console.log('Waiting for telemetry export to complete...\n');
}, 105000);

// Graceful shutdown after all data is sent
setTimeout(() => {
  console.log('Shutting down mobile app...');
  sdk.shutdown()
    .then(() => {
      console.log('✅ Mobile app stopped successfully');
      console.log('All telemetry data has been sent to the collector.');
      process.exit(0);
    })
    .catch((error) => {
      console.error('❌ Error during shutdown:', error);
      process.exit(1);
    });
}, 120000);
