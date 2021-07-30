package prom

import (
	"runtime"

	"github.com/creativeprojects/resticprofile/progress"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

const namespace = "resticprofile"
const backup = "backup"
const groupLabel = "group"
const profileLabel = "profile"
const goVersionLabel = "goversion"
const versionLabel = "version"

type Metrics struct {
	group    string
	registry *prometheus.Registry
	info     *prometheus.GaugeVec
	backup   BackupMetrics
}

func NewMetrics(group, version string) *Metrics {
	registry := prometheus.NewRegistry()
	p := &Metrics{
		group:    group,
		registry: registry,
	}
	p.info = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "build_info",
		Help:      "resticprofile build information.",
	}, []string{goVersionLabel, versionLabel})
	p.info.With(prometheus.Labels{goVersionLabel: runtime.Version(), versionLabel: version})

	p.backup = newBackupMetrics(group)

	registry.MustRegister(
		p.info,
		p.backup.duration,
		p.backup.filesNew,
		p.backup.filesChanged,
		p.backup.filesUnmodified,
		p.backup.dirNew,
		p.backup.dirChanged,
		p.backup.dirUnmodified,
		p.backup.filesTotal,
		p.backup.bytesAdded,
		p.backup.bytesTotal,
		p.backup.status,
	)
	return p
}

func (p *Metrics) BackupResults(profile string, status Status, summary progress.Summary) {
	labels := prometheus.Labels{profileLabel: profile}
	if p.group != "" {
		labels[groupLabel] = p.group
	}
	p.backup.duration.With(labels).Set(summary.Duration.Seconds())

	p.backup.filesNew.With(labels).Set(float64(summary.FilesNew))
	p.backup.filesChanged.With(labels).Set(float64(summary.FilesChanged))
	p.backup.filesUnmodified.With(labels).Set(float64(summary.FilesUnmodified))

	p.backup.dirNew.With(labels).Set(float64(summary.DirsNew))
	p.backup.dirChanged.With(labels).Set(float64(summary.DirsChanged))
	p.backup.dirUnmodified.With(labels).Set(float64(summary.DirsUnmodified))

	p.backup.filesTotal.With(labels).Set(float64(summary.FilesTotal))
	p.backup.bytesAdded.With(labels).Set(float64(summary.BytesAdded))
	p.backup.bytesTotal.With(labels).Set(float64(summary.BytesTotal))
	p.backup.status.With(labels).Set(float64(status))
}

func (p *Metrics) SaveTo(filename string) error {
	return prometheus.WriteToTextfile(filename, p.registry)
}

func (p *Metrics) Push(url, jobName string) error {
	return push.New(url, jobName).
		Gatherer(p.registry).
		Add()
}
