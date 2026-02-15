package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RequestsTotal counts HTTP requests by method, path, and status code.
	RequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "pollex_requests_total",
		Help: "Total HTTP requests processed.",
	}, []string{"method", "path", "status"})

	// PolishDuration tracks inference latency per model.
	PolishDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "pollex_polish_duration_seconds",
		Help:    "Time spent on polish inference.",
		Buckets: []float64{0.5, 1, 2, 5, 10, 20, 30, 60, 120},
	}, []string{"model"})

	// InputChars tracks the distribution of input text lengths.
	InputChars = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "pollex_input_chars",
		Help:    "Number of characters in polish input text.",
		Buckets: []float64{50, 100, 250, 500, 1000, 2500, 5000, 10000},
	})

	// AdapterAvailable tracks whether each adapter is reachable.
	AdapterAvailable = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pollex_adapter_available",
		Help: "Whether an LLM adapter is available (1) or not (0).",
	}, []string{"adapter"})
)
