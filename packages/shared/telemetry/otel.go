package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"

	"go.uber.org/zap"
)

func NewTracerProvider() (*sdktrace.TracerProvider, error) {
	// TODO-implement otlp for production use fx.in if possible
	// ctx := context.Background() // use for otlp
	exporter, err := stdout.New(stdout.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter %w", err)

	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
	)

	// set tracer defaults
	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil

}

// clean shutdown tracer provider

func RegisterShutdown(lc fx.Lifecycle, logger *zap.Logger, tp *sdktrace.TracerProvider) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			if err := tp.Shutdown(ctx); err != nil {
				logger.Error("failed to shutdown tracer provider", zap.Error(err))
			} else {
				logger.Info("tracer provider shutdown successfully")
			}
			return nil
		},
	})
}
