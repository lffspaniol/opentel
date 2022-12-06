package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	calingapi "opentel/calingApi"
	config "opentel/config"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

var Config *config.Config

func initTracer(Config *config.Config) (*sdktrace.TracerProvider, error) {
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(Config.Facility),
			semconv.ServiceVersionKey.String("v0.1.0"),
			attribute.String("environment", "demo"),
		),
	)
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(Config.ColectorServer),
	)
	exporter, err := otlptrace.New(context.Background(), client)
	if err != nil {
		return nil, err
	}

	batch := sdktrace.NewBatchSpanProcessor(exporter)

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(r),
		sdktrace.WithSpanProcessor(batch),
	)
	return tracerProvider, nil
}

func main() {
	Config = config.Load()
	tp, err := initTracer(Config)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

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
	_, span := otel.Tracer("HelloServer").Start(r.Context(), "HelloServer")
	defer span.End()
	fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])

}

func RemoteRequest(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("HelloServer").Start(r.Context(), "HelloServer")
	defer span.End()
	calingapi.Request(ctx)
}
