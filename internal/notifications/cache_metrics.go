package notifications

import "github.com/prometheus/client_golang/prometheus"

func init() {
	prometheus.MustRegister(configsMetric)
	prometheus.MustRegister(clientsMetric)
}

var (
	configsMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "otf",
		Subsystem: "notifier_cache",
		Name:      "configs",
		Help:      "Number of configs in notifier cache",
	})
	clientsMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "otf",
		Subsystem: "notifier_cache",
		Name:      "clients",
		Help:      "Number of clients in notifier cache",
	})
)
