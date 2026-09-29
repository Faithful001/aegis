package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPInFlight        prometheus.Gauge

	InferenceRequestsTotal *prometheus.CounterVec
	InferenceDuration      *prometheus.HistogramVec
	TokensTotal            *prometheus.CounterVec
	RateLimitHitsTotal     *prometheus.CounterVec
	ActiveWorkers          prometheus.Gauge
}

var AppMetrics = newMetrics()

func newMetrics() *Metrics {
	return &Metrics{
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "aegis",
				Subsystem: "http",
				Name:      "requests_total",
				Help:      "Total number of HTTP requests processed by Aegis.",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "aegis",
				Subsystem: "http",
				Name:      "request_duration_seconds",
				Help:      "Histogram of HTTP request latencies in seconds.",
				Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30},
			},
			[]string{"method", "path", "status"},
		),
		HTTPInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "aegis",
				Subsystem: "http",
				Name:      "in_flight_requests",
				Help:      "Current number of in-flight HTTP requests.",
			},
		),
		InferenceRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "aegis",
				Subsystem: "inference",
				Name:      "requests_total",
				Help:      "Total number of AI inference requests.",
			},
			[]string{"provider", "model", "status"},
		),
		InferenceDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "aegis",
				Subsystem: "inference",
				Name:      "duration_seconds",
				Help:      "Duration of AI model inference requests in seconds.",
				Buckets:   []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0, 20.0, 45.0, 60.0},
			},
			[]string{"provider", "model"},
		),
		TokensTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "aegis",
				Subsystem: "inference",
				Name:      "tokens_total",
				Help:      "Total number of tokens consumed by type, provider, and model.",
			},
			[]string{"type", "provider", "model"}, // type: prompt | completion
		),
		RateLimitHitsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "aegis",
				Subsystem: "ratelimit",
				Name:      "hits_total",
				Help:      "Total rate limiter evaluations by result (allowed vs blocked).",
			},
			[]string{"status", "tier"},
		),
		ActiveWorkers: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "aegis",
				Subsystem: "worker",
				Name:      "active_count",
				Help:      "Number of currently active/healthy worker nodes.",
			},
		),
	}
}
