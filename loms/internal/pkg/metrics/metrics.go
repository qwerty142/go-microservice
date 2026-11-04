package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc/status"
)

var (
	requestDurationSecondsHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "loms",
			Name:      "request_duration_seconds",
			Help:      "Histogram of request duration in seconds.",
		},
		[]string{"handler", "status"},
	)

	databaseRequestDurationSecondsHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "loms",
			Name:      "database_request_duration_seconds",
			Help:      "Histogram of database request duration in seconds.",
		},
		[]string{"error", "operation"},
	)
)

func SaveRequestDuration(startTime time.Time, handler string, error error) {
	statusCodeInt := 0
	if error != nil {
		statusCode, _ := status.FromError(error)
		statusCodeInt = int(statusCode.Code())
	}
	requestDurationSecondsHistogram.WithLabelValues(handler, strconv.Itoa(statusCodeInt)).Observe(time.Since(startTime).Seconds())
}

func SaveDatabaseRequestDuration(startTime time.Time, error error, operation string) {
	var errorMessage string
	if error != nil {
		errorMessage = error.Error()
	}

	databaseRequestDurationSecondsHistogram.WithLabelValues(errorMessage, operation).Observe(time.Since(startTime).Seconds())
}
