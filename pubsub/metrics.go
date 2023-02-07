package pubsub

import "github.com/prometheus/client_golang/prometheus"

func init() {
	prometheus.MustRegister(totalSubscribers)
}

var totalSubscribers = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "otf",
	Subsystem: "pub_sub",
	Name:      "total_subscribers",
	Help:      "Total number of subscribers.",
})
