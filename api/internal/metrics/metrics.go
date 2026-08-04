package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "statuspage_api_requests_total",
			Help: "Total number of HTTP requests processed by StatusPage API",
		},
		[]string{"method", "endpoint", "status"},
	)

	ActiveMonitorsCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "statuspage_monitors_count",
			Help: "Current number of active monitors managed by StatusPage",
		},
		[]string{"tenant_id"},
	)

	CheckLatencyHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "statuspage_check_latency_seconds",
			Help:    "Latency of monitor HTTP check executions in seconds",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 10), // 10ms to ~10s
		},
		[]string{"monitor_id"},
	)

	DowntimeEventsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "statuspage_downtime_events_total",
			Help: "Total count of detected monitor downtime transitions",
		},
		[]string{"tenant_id"},
	)
)
