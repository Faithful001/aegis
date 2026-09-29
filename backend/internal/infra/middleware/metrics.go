package middleware

import (
	"strconv"
	"time"

	"github.com/Faithful001/aegis/internal/infra/observability"
	"github.com/gin-gonic/gin"
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		observability.AppMetrics.HTTPInFlight.Inc()
		start := time.Now()

		c.Next()

		observability.AppMetrics.HTTPInFlight.Dec()
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		if path == "" {
			path = c.Request.URL.Path
		}

		observability.AppMetrics.HTTPRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		observability.AppMetrics.HTTPRequestDuration.WithLabelValues(c.Request.Method, path, status).Observe(duration)
	}
}
