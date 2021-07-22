package prom

import (
	"github.com/creativeprojects/resticprofile/shell"
	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "resticprofile"
const backup = "backup"

type Status uint

const (
	StatusFailed Status = iota
	StatusWarning
	StatusSuccess
)

type Backup struct {
	duration        prometheus.Gauge
	filesNew        prometheus.Gauge
	filesChanged    prometheus.Gauge
	filesUnmodified prometheus.Gauge
	dirNew          prometheus.Gauge
	dirChanged      prometheus.Gauge
	dirUnmodified   prometheus.Gauge
	filesTotal      prometheus.Gauge
	bytesAdded      prometheus.Gauge
	bytesTotal      prometheus.Gauge
	status          prometheus.Gauge
	registry        *prometheus.Registry
}

func NewBackup() *Backup {
	registry := prometheus.NewRegistry()
	p := &Backup{
		registry: registry,
	}
	createMetrics(p)

	registry.MustRegister(
		p.duration,
		p.filesNew,
		p.filesChanged,
		p.filesUnmodified,
		p.dirNew,
		p.dirChanged,
		p.dirUnmodified,
		p.filesTotal,
		p.bytesAdded,
		p.bytesTotal,
		p.status,
	)
	return p
}

func createMetrics(p *Backup) {
	p.duration = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: backup,
		Name:      "duration_seconds",
		Help:      "The duration of backup in seconds.",
	})
	p.filesNew = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: backup,
		Name:      "files_new",
		Help:      "Number of new files added to the backup.",
	})
	p.filesChanged = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: backup,
		Name:      "files_changed",
		Help:      "Number of files with changes.",
	})
	p.filesUnmodified = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: backup,
		Name:      "files_unmodified",
		Help:      "Number of files unmodified since last backup.",
	})
	p.dirNew = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: backup,
		Name:      "dir_new",
		Help:      "Number of new directories added to the backup.",
	})
	p.dirChanged = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: backup,
		Name:      "dir_changed",
		Help:      "Number of directories with changes.",
	})
	p.dirUnmodified = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: backup,
		Name:      "dir_unmodified",
		Help:      "Number of directories unmodified since last backup.",
	})
	p.filesTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: backup,
		Name:      "files_processed",
		Help:      "Total number of files scanned by the backup for changes.",
	})
	p.bytesAdded = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: backup,
		Name:      "added_bytes",
		Help:      "Total number of bytes added to the repository.",
	})
	p.bytesTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: backup,
		Name:      "processed_bytes",
		Help:      "Total number of bytes scanned for changes.",
	})
	p.status = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: backup,
		Name:      "status",
		Help:      "Backup status: 0=fail, 1=warning, 2=success.",
	})
}

func (p *Backup) Results(status Status, summary shell.Summary) {
	p.bytesAdded.Set(float64(summary.BytesAdded))
	p.bytesTotal.Set(float64(summary.BytesTotal))
	p.status.Set(float64(status))
}

func (p *Backup) SaveTo(filename string) {
	prometheus.WriteToTextfile(filename, p.registry)
}
