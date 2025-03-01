use std::env;

use opentelemetry::{global, KeyValue};
use opentelemetry_otlp::WithExportConfig;
use opentelemetry_sdk::{propagation::TraceContextPropagator, trace::SdkTracerProvider, Resource};
use pyroscope::pyroscope::PyroscopeAgentRunning;
use pyroscope::PyroscopeAgent;
use pyroscope_pprofrs::{pprof_backend, PprofConfig};

pub fn start_pyroscope() -> PyroscopeAgent<PyroscopeAgentRunning> {
    let pprof_config = PprofConfig::new().sample_rate(100);
    let backend_impl = pprof_backend(pprof_config);

    let service_name = env::var("OTEL_SERVICE_NAME").expect("OTEL_SERVICE_NAME is not defined");
    let endpoint = env::var("PYROSCOPE_ENDPOINT").expect("PYROSCOPE_ENDPOINT is not defined");
    let gbl_env = env::var("RUST_ENV").expect("RUST_ENV is not defined");

    let agent = PyroscopeAgent::builder(&endpoint, &service_name)
        .backend(backend_impl)
        .tags(vec![("env", &gbl_env)])
        .build()
        .unwrap();

    agent.start().unwrap()
}

pub fn start_tracing() -> SdkTracerProvider {
    let service_name = env::var("OTEL_SERVICE_NAME").expect("OTEL_SERVICE_NAME is not defined");
    let endpoint = env::var("OTEL_ENDPOINT").expect("OTEL_ENDPOINT is not defined");
    let gbl_env = env::var("RUST_ENV").expect("RUST_ENV is not defined");

    global::set_text_map_propagator(TraceContextPropagator::new());
    let service_name_resource = Resource::builder_empty()
        .with_attribute(KeyValue::new("service.name", service_name))
        .with_attribute(KeyValue::new("env", gbl_env))
        .build();

    println!("{}", endpoint.as_str());

    let tracer = SdkTracerProvider::builder()
        .with_batch_exporter(
            opentelemetry_otlp::SpanExporter::builder()
                .with_tonic()
                .with_endpoint(endpoint.as_str())
                .build()
                .unwrap(),
        )
        .with_resource(service_name_resource)
        .build();

    global::set_tracer_provider(tracer.clone());

    tracer
}
