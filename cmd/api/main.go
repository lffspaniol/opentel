package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	calingapi "opentel/calingApi"
	config "opentel/config"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var res = resource.NewWithAttributes(
	semconv.SchemaURL,
	semconv.ServiceNameKey.String("api"),
)

var Config *config.Config

func initProvider() (func(), error) {
	log.Println("initProvider")
	ctx := context.Background()

	otelAgentAddr := viper.GetString("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otelAgentAddr == "" {
		otelAgentAddr = "0.0.0.0:4317"
	}
	log.Println(otelAgentAddr)

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, otelAgentAddr,
		// Note the use of insecure transport here. TLS is recommended in production.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	p, err := Metrics(ctx, conn, otelAgentAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get up the metrics %w", err)
	}

	t, err := Tracer(ctx, conn, otelAgentAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get up tracer %w", err)
	}

	return func() {
		p()
		t()
	}, nil
}

func main() {
	Config = config.Load()
	f, err := initProvider()
	if err != nil {
		log.Fatal(err)
	}

	defer f()

	fmt.Println(Config)
	r := mux.NewRouter()
	r.Use(otelmux.Middleware("my-server"))

	r.HandleFunc("/", HelloServer)
	r.HandleFunc("/remote", RemoteRequest)

	http.Handle("/", r)
	log.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
}

func RemoteRequest(w http.ResponseWriter, r *http.Request) {
	calingapi.Request(r.Context())
	time.Sleep(1 * time.Second)
}

func Tracer(ctx context.Context, conn *grpc.ClientConn, uri string) (func(), error) {
	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	batch := sdktrace.NewBatchSpanProcessor(traceExporter)

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(batch),
	)
	otel.SetTracerProvider(tracerProvider)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return func() {
		tracerProvider.Shutdown(ctx)
	}, nil
}

func Metrics(ctx context.Context, conn *grpc.ClientConn, uri string) (func(), error) {
	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithGRPCConn(conn),
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(uri),
	)

	if err != nil {
		log.Panicln(uri, err)
		return nil, err
	}
	read := metric.NewPeriodicReader(metricExporter, metric.WithInterval(1*time.Second))
	provider := metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(read))

	global.SetMeterProvider(provider)

	err = runtime.Start(runtime.WithMinimumReadMemStatsInterval(time.Second))
	if err != nil {
		log.Fatal(err)
	}
	return func() {
		provider.Shutdown(ctx)
	}, nil
}
