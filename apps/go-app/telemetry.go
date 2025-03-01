package telemetry

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/grafana/pyroscope-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var serviceName = semconv.ServiceNameKey.String("go-app")

func OtelMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add a trace
		tracer := otel.Tracer("default")
		ctx, span := tracer.Start(c.Request.Context(), fmt.Sprintf("%s-handler", c.FullPath()))
		defer span.End()

		// Add a metric
		meter := otel.Meter("default")
		counter, err := meter.Int64Counter("ping_count")
		if err != nil {
			log.Fatal(err)
		}
		counter.Add(ctx, 1, metric.WithAttributes(attribute.String("endpoint", c.FullPath())))

		// Continue to the next handler
		c.Next()
	}
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func InitOtel(ctx context.Context) func(ctx context.Context) error {
	conn, err := InitConn()
	if err != nil {
		log.Fatal(err)
	}

	var shutdownFuncs []func(context.Context) error

	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)
	res, err := resource.New(ctx,
		resource.WithAttributes(
			// The service name used to display traces in backends
			serviceName,
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	tracerProvider, err := InitTracerProvider(ctx, res, conn)
	if err != nil {
		log.Fatal(err)
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)

	otel.SetTracerProvider(tracerProvider)

	meterProvider, err := InitMeterProvider(ctx, res, conn)
	if err != nil {
		log.Fatal(err)
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)

	otel.SetMeterProvider(meterProvider)

	return shutdown
}

// Initialize a gRPC connection to be used by both the tracer and meter providers.
func InitConn() (*grpc.ClientConn, error) {
	log.Println("Initializing gRPC connection to OTEL collector...")
	conn, err := grpc.Dial("localhost:4317",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}
	log.Println("gRPC connection established.")
	return conn, err
}

// Initializes an OTLP exporter, and configures the corresponding trace provider.
func InitTracerProvider(ctx context.Context, res *resource.Resource, conn *grpc.ClientConn) (*sdktrace.TracerProvider, error) {
	log.Println("Initializing trace provider...")
	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	// Set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	log.Println("Trace provider initialized.")
	// Shutdown will flush any remaining spans and shut down the exporter.
	return tracerProvider, nil
}

// Initializes an OTLP exporter, and configures the corresponding meter provider.
func InitMeterProvider(ctx context.Context, res *resource.Resource, conn *grpc.ClientConn) (*sdkmetric.MeterProvider, error) {
	log.Println("Initializing meter provider...")
	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics exporter: %w", err)
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)

	log.Println("Meter provider initialized.")
	return meterProvider, nil
}

func InitPyroscope() *pyroscope.Profiler {
	log.Println("Initializing Pyroscope...")
	// These 2 lines are only required if you're using mutex or block profiling
	// Read the explanation below for how to set these rates:
	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)

	profiler, err := pyroscope.Start(pyroscope.Config{
		ApplicationName: os.Getenv("OTEL_SERVICE_NAME"),

		// replace this with the address of pyroscope server
		ServerAddress: os.Getenv("PYROSCOPE_ENDPOINT"),

		// you can disable logging by setting this to nil
		Logger: pyroscope.StandardLogger,

		// you can provide static tags via a map:
		Tags: map[string]string{"app": "go-app"},

		ProfileTypes: []pyroscope.ProfileType{
			// these profile types are enabled by default:
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,

			// these profile types are optional:
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	})

	if err != nil {
		log.Fatalf("could not start pyroscope: %v", err)
	}

	return profiler
}
