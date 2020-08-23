package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Counter to track failed ssh login attempts
var failedAttempts = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "failed_conn_attempts_total",
		Help: "Number of failed ssh connection attempts",
	},
	[]string{"country"},
)

// Reporter defines the behaviour for failed connection event reporters
type Reporter interface {
	Report(e FailedConnEvent) error
}

type prometheusReporter struct{}

func (pr prometheusReporter) Report(e FailedConnEvent) error {
	failedAttempts.WithLabelValues(e.Country).Inc()

	return nil
}

func init() {
	prometheus.MustRegister(failedAttempts)
}
