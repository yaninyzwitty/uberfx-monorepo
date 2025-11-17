package telemetry

import (
	"fmt"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"tracer-provider",
	fx.Provide(NewTracer),
	fx.Invoke(RegisterShutdown),
)

type TracerOut struct {
	fx.Out

	TP     *sdktrace.TracerProvider
	Tracer trace.Tracer
}

func NewTracer() (TracerOut, error) {
	tp, err := NewTracerProvider()
	if err != nil {
		return TracerOut{}, fmt.Errorf("tracerprovider error: %w", err)
	}

	return TracerOut{
		TP:     tp,
		Tracer: otel.Tracer("product-service"),
	}, nil
}
