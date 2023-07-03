package middleware

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

const msMultiplier = 1000

var (
	requestsDuration *prometheus.HistogramVec
	requestsCounter  *prometheus.CounterVec
)

func init() {
	var err error

	requestsDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "request_duration_endpoint_milliseconds",
		Help:    "Time taken to process request",
		Buckets: []float64{20, 50, 100, 500},
	}, []string{"endpoint", "method"})

	err = prometheus.Register(requestsDuration)
	if err != nil {
		log.Error().
			Err(err).
			Fields(map[string]interface{}{"metric": "request_duration_endpoint_milliseconds"}).
			Msg("unable to register prometheus metric")
	}

	requestsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "request_count_endpoint",
		Help: "How many HTTP requests processed, partitioned by status code and HTTP method.",
	}, []string{"endpoint", "method"})

	err = prometheus.Register(requestsCounter)
	if err != nil {
		log.Error().
			Err(err).
			Fields(map[string]interface{}{"metric": "request_count_endpoint"}).
			Msg("unable to register prometheus metric")
	}
}

func Prometheus(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// detect execution time
		endpoint := mux.CurrentRoute(r).GetName()
		timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
			requestsDuration.WithLabelValues(endpoint, r.Method).Observe(v * msMultiplier)
		}))

		defer timer.ObserveDuration()

		requestsCounter.WithLabelValues(endpoint, r.Method).Inc()

		next.ServeHTTP(w, r)
	})
}
