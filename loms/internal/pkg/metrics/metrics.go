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

// SaveRequestDuration records the grpc request duration labeled with the returned status code.
func SaveRequestDuration(startTime time.Time, handler string, err error) {
	statusCodeInt := 0
	if err != nil {
		statusCode, _ := status.FromError(err)
		statusCodeInt = int(statusCode.Code())
	}
	requestDurationSecondsHistogram.WithLabelValues(handler, strconv.Itoa(statusCodeInt)).Observe(time.Since(startTime).Seconds())
}

// SaveDatabaseRequestDuration records the database request duration labeled with the error and operation.
func SaveDatabaseRequestDuration(startTime time.Time, err error, operation string) {
	var errorMessage string
	if err != nil {
		errorMessage = err.Error()
	}

	databaseRequestDurationSecondsHistogram.WithLabelValues(errorMessage, operation).Observe(time.Since(startTime).Seconds())
}
