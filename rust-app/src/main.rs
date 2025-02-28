use actix_web::web::Path;
use actix_web::{web, App, HttpRequest, HttpServer};
use actix_web_opentelemetry::RequestTracing;
use opentelemetry::{global, KeyValue};
use opentelemetry_otlp::WithExportConfig;
use opentelemetry_sdk::{propagation::TraceContextPropagator, trace::SdkTracerProvider, Resource};

async fn index(_req: HttpRequest, _path: Path<String>) -> &'static str {
    "Hello world!"
}

#[actix_web::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    env_logger::init();
    global::set_text_map_propagator(TraceContextPropagator::new());
    let service_name_resource = Resource::builder_empty()
        .with_attribute(KeyValue::new("service.name", "actix_server"))
        .build();

    let tracer = SdkTracerProvider::builder()
        .with_batch_exporter(
            opentelemetry_otlp::SpanExporter::builder()
                .with_tonic()
                .with_endpoint("http://127.0.0.1:4317")
                .build()?,
        )
        .with_resource(service_name_resource)
        .build();

    global::set_tracer_provider(tracer.clone());

    log::info!("📡 Server starting ! Listening");
    HttpServer::new(move || {
        App::new()
            .wrap(RequestTracing::new())
            .service(web::resource("/users/{id}").to(index))
    })
    .bind("127.0.0.1:8080")?
    .run()
    .await?;

    // Ensure all spans have been reported
    tracer.shutdown()?;

    Ok(())
}
