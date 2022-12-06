package calingapi

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("calingapi")

func Request(ctx context.Context) {
	_, span := tracer.Start(ctx, "Request")
	defer span.End()
	time.Sleep(1 * time.Second)
}
