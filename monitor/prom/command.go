package prom

import (
	"github.com/prometheus/client_golang/prometheus"
)

// CommandMetrics are the metrics reported for a restic command other than backup
type CommandMetrics struct {
	duration *prometheus.GaugeVec
	exitCode *prometheus.GaugeVec
	status   *prometheus.GaugeVec
	time     *prometheus.GaugeVec
}

func newCommandMetrics(command string, labels []string) CommandMetrics {
	commandMetrics := CommandMetrics{
		duration: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: command,
			Name:      "duration_seconds",
			Help:      "The " + command + " duration (in seconds).",
		}, labels),
		exitCode: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: command,
			Name:      "exit_code",
			Help:      "Exit code of restic " + command + " (-1 when restic did not exit normally).",
		}, labels),
		status: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: command,
			Name:      "status",
			Help:      "The " + command + " status: 0=fail, 1=warning, 2=success.",
		}, labels),
		time: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: command,
			Name:      "time_seconds",
			Help:      "Last " + command + " run (unixtime).",
		}, labels),
	}
	return commandMetrics
}

func (m CommandMetrics) collectors() []prometheus.Collector {
	return []prometheus.Collector{m.duration, m.exitCode, m.status, m.time}
}
