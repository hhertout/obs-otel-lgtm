use actix_web::{web, App, HttpServer};
use actix_web_opentelemetry::RequestTracing;
use dotenv::dotenv;

mod telemetry;

async fn ping() -> &'static str {
    "Hello world!"
}

#[actix_web::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    dotenv().ok();

    env_logger::init();

    telemetry::start_pyroscope();
    let tracer = telemetry::start_tracing();

    log::info!("📡 Server starting ! Listening");
    HttpServer::new(move || {
        App::new()
            .wrap(RequestTracing::new())
            .service(web::resource("/ping").to(ping))
    })
    .bind("127.0.0.1:8080")?
    .run()
    .await?;

    // Ensure all spans have been reported
    tracer.shutdown()?;

    Ok(())
}
