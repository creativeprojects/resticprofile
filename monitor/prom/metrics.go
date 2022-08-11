package prom

import (
	"runtime"
	"time"

	"github.com/creativeprojects/resticprofile/monitor"
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
	group        string
	configLabels map[string]string
	registry     *prometheus.Registry
	info         *prometheus.GaugeVec
	backup       BackupMetrics
}

func NewMetrics(group, version string, configLabels map[string]string) *Metrics {
	registry := prometheus.NewRegistry()
	p := &Metrics{
		group:        group,
		configLabels: configLabels,
		registry:     registry,
	}
	p.info = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "build_info",
		Help:      "resticprofile build information.",
	}, mergeKeys([]string{goVersionLabel, versionLabel}, configLabels))
	p.info.With(mergeLabels(prometheus.Labels{goVersionLabel: runtime.Version(), versionLabel: version}, configLabels)).Set(1)

	p.backup = newBackupMetrics(group, configLabels)

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
		p.backup.time,
	)
	return p
}

func (p *Metrics) BackupResults(profile string, status Status, summary monitor.Summary) {
	labels := prometheus.Labels{profileLabel: profile}
	if p.group != "" {
		labels[groupLabel] = p.group
	}
	labels = mergeLabels(labels, p.configLabels)
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
	p.backup.time.With(labels).Set(float64(time.Now().Unix()))
}

func (p *Metrics) SaveTo(filename string) error {
	return prometheus.WriteToTextfile(filename, p.registry)
}

func (p *Metrics) Push(url, jobName string) error {
	return push.New(url, jobName).
		Gatherer(p.registry).
		Add()
}

func mergeKeys(keys []string, add map[string]string) []string {
	for key := range add {
		keys = append(keys, key)
	}
	return keys
}

func mergeLabels(labels prometheus.Labels, add map[string]string) prometheus.Labels {
	for key, value := range add {
		labels[key] = value
	}
	return labels
}
