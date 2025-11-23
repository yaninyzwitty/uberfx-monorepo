package metrics

import "go.uber.org/fx"

var Module = fx.Module(
	"metrics",
	RegistryModule,
	GRPCMetricsModule,
	AppMetricsModule,
	PromHTTPModule,
)
