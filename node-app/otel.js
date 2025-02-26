const { NodeSDK } = require('@opentelemetry/sdk-node');
const { getNodeAutoInstrumentations } = require('@opentelemetry/auto-instrumentations-node');
const { OTLPTraceExporter } = require('@opentelemetry/exporter-trace-otlp-proto');
const { OTLPMetricExporter } = require('@opentelemetry/exporter-metrics-otlp-proto');
const { OTLPLogExporter } = require('@opentelemetry/exporter-logs-otlp-proto');
const { PeriodicExportingMetricReader } = require('@opentelemetry/sdk-metrics');
const { SimpleLogRecordProcessor, ConsoleLogRecordExporter } = require('@opentelemetry/sdk-logs');
const { LoggerProvider } = require('@opentelemetry/sdk-logs');
const { Resource } = require("@opentelemetry/resources");
const { logs, SeverityNumber } = require('@opentelemetry/api-logs');
const {
    ATTR_SERVICE_NAME,
    ATTR_SERVICE_VERSION,
} = require('@opentelemetry/semantic-conventions');
const dotenv = require('dotenv');
dotenv.config();

const endPoint = process.env.OTEL_ENDPOINT;
const serviceName = process.env.OTEL_SERVICE_NAME;
const logEnabled = process.env.OTEL_LOG_ENABLED == "true";

console.log("🔧 OpenTelemetry configuration")
console.log(`🔗 Endpoint: ${endPoint}`)
console.log(`📦 Service name: ${serviceName}`)
console.log(`📝 Log enabled: ${logEnabled}`)
console.log("")
console.log("🔗 OpenTelemetry starting")

const resource = new Resource({
    [ATTR_SERVICE_NAME]: serviceName || 'unknown-app',
    [ATTR_SERVICE_VERSION]: '1.0',
})

const traceExporter = new OTLPTraceExporter({
    url: endPoint ? `${endPoint}/v1/traces` : 'http://localhost:4318/v1/traces',
});

const metricExporter = new OTLPMetricExporter({
    url: endPoint ? `${endPoint}/v1/metrics` : 'http://localhost:4318/v1/metrics',
});

const loggerProvider = new LoggerProvider({ resource });
if (logEnabled) {
    loggerProvider.addLogRecordProcessor(new SimpleLogRecordProcessor(new OTLPLogExporter({
        url: endPoint ? `${endPoint}/v1/logs` : "http://localhost:4318/v1/logs",
    })));
    logs.setGlobalLoggerProvider(loggerProvider);

    const loggerOtel = loggerProvider.getLogger('default');

    loggerOtel.emit({
        severityNumber: SeverityNumber.INFO,
        severityText: "INFO",
        body: "Hello OpenTelemetry",
        attributes: {
            'logs.type': 'LogRecord'
        }
    })

    const logger = logs.getLogger("console");
    const oldConsoleLog = console.log;
    const oldConsoleWarn = console.warn;
    const oldConsoleError = console.error;

    console.log = (...args) => {
        logger.emit({
            severityNumber: SeverityNumber.INFO,
            severityText: "INFO",
            body: args.map(String).join(" "), // Convertit les arguments en string
        });
        oldConsoleLog(...args); // Affiche aussi dans la console normale
    };

    console.warn = (...args) => {
        logger.emit({
            severityNumber: SeverityNumber.WARN,
            severityText: "WARN",
            body: args.map(String).join(" "),
        });
        oldConsoleWarn(...args);
    }

    console.error = (...args) => {
        logger.emit({
            severityNumber: SeverityNumber.ERROR,
            severityText: "ERROR",
            body: args.map(String).join(" "),
        });
        oldConsoleError(...args);
    }
}


const sdk = new NodeSDK({
    resource,
    traceExporter,
    metricReader: new PeriodicExportingMetricReader({
        exporter: metricExporter,
        exportIntervalMillis: 60000, // Envoi des métriques toutes les 60s
    }),
    instrumentations: [getNodeAutoInstrumentations()], // Auto-instrumentation
    loggerProvider,
});



try {
    console.log('🚀 OpenTelemetry started');
    sdk.start()
} catch (error) {
    console.error(error);
    process.exit(1);
}


process.on('SIGTERM', () => {
    sdk.shutdown().then(() => console.log('🛑 OpenTelemetry gracefully shutdown')).catch(console.error);
});
