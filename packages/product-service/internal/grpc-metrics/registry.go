package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
)

type RegistryResult struct {
	fx.Out
	Registry *prometheus.Registry
}

func NewRegistry() RegistryResult {
	return RegistryResult{
		Registry: prometheus.NewRegistry(),
	}
}

var RegistryModule = fx.Provide(NewRegistry)
