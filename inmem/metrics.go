package inmem

import "github.com/prometheus/client_golang/prometheus"

var cacheSize = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "otf",
	Subsystem: "cache",
	Name:      "size",
	Help:      "Maximum size of cache.",
})

var cacheUsed = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "otf",
	Subsystem: "cache",
	Name:      "used",
	Help:      "Number of bytes stored in cache.",
})

var cacheEntries = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "otf",
	Subsystem: "cache",
	Name:      "entries",
	Help:      "Number of entries in cache.",
})

var cacheHits = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "otf",
	Subsystem: "cache",
	Name:      "hits",
	Help:      "Number of successfully found keys in cache.",
})

var cacheMisses = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "otf",
	Subsystem: "cache",
	Name:      "misses",
	Help:      "Number of not found keys in cache.",
})

var cacheDelHits = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "otf",
	Subsystem: "cache",
	Name:      "del_hits",
	Help:      "Number of successfully deleted keys in cache.",
})

var cacheDelMisses = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "otf",
	Subsystem: "cache",
	Name:      "del_misses",
	Help:      "Number of not deleted keys in cache.",
})

var cacheCollisions = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "otf",
	Subsystem: "cache",
	Name:      "collisions",
	Help:      "Number of happened key-collisions in cache.",
})

func init() {
	prometheus.MustRegister(cacheSize)
	prometheus.MustRegister(cacheUsed)
	prometheus.MustRegister(cacheEntries)
	prometheus.MustRegister(cacheHits)
	prometheus.MustRegister(cacheMisses)
	prometheus.MustRegister(cacheDelHits)
	prometheus.MustRegister(cacheDelMisses)
	prometheus.MustRegister(cacheCollisions)
}
