package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestDurationSecondsHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "cart",
			Name:      "http_request_duration_seconds",
			Help:      "Histogram of http request duration in seconds.",
		},
		[]string{"handler", "status"},
	)

	externalCallDurationSecondsHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "cart",
			Name:      "external_call_duration_seconds",
			Help:      "Histogram of external resource request duration in seconds.",
		},
		[]string{"handler", "status"},
	)

	InMemoryElementsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "cart",
			Name:      "in_memory_elements_total",
			Help:      "Total number of in-memory elements.",
		},
	)
)

func SaveRequestDuration(startTime time.Time, handler string, status int) {
	httpRequestDurationSecondsHistogram.
		WithLabelValues(handler, strconv.Itoa(status)).
		Observe(time.Since(startTime).Seconds())
}

func SaveExternalCallDuration(startTime time.Time, handler string, status int) {
	externalCallDurationSecondsHistogram.
		WithLabelValues(handler, strconv.Itoa(status)).
		Observe(time.Since(startTime).Seconds())
}
