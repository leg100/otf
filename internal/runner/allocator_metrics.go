package runner

import "github.com/prometheus/client_golang/prometheus"

func init() {
	prometheus.MustRegister(currentJobsMetric)
}

const runnerIDLabel = "runner_id"

var currentJobsMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "otf",
	Subsystem: "allocator",
	Name:      "current_jobs",
	Help:      "Current jobs by runner ID",
}, []string{runnerIDLabel})
