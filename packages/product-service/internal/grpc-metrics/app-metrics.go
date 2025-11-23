package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
)

type AppMetrics struct {
	Stage    prometheus.Gauge
	Duration *prometheus.HistogramVec
	Errors   *prometheus.CounterVec
}

type AppMetricsResult struct {
	fx.Out
	Metrics *AppMetrics
}

type AppMetricsParams struct {
	fx.In
	Registry *prometheus.Registry
}

func NewAppMetrics(p AppMetricsParams) AppMetricsResult {
	m := &AppMetrics{
		Stage: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "myapp",
			Name:      "stage",
		}),
		Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "myapp",
			Name:      "request_duration_seconds",
		}, []string{"op", "db"}),
		Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "myapp",
			Name:      "errors_total",
		}, []string{"op", "db"}),
	}

	p.Registry.MustRegister(m.Stage, m.Duration, m.Errors)

	return AppMetricsResult{Metrics: m}
}

var AppMetricsModule = fx.Provide(NewAppMetrics)
