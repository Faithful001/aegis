package metrics

import "github.com/prometheus/client_golang/prometheus/promhttp"

type MetricsHandler struct{}

func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{}
}

func (m *MetricsHandler) Init() {
	promhttp.Handler()
}