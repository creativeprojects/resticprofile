package prom

import (
	"runtime"
	"time"

	"github.com/creativeprojects/resticprofile/monitor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/prometheus/common/expfmt"
	"golang.org/x/exp/maps"
)

const (
	namespace      = "resticprofile"
	backup         = "backup"
	groupLabel     = "group"
	profileLabel   = "profile"
	goVersionLabel = "goversion"
	versionLabel   = "version"
)

type Metrics struct {
	labels   prometheus.Labels
	registry *prometheus.Registry
	info     *prometheus.GaugeVec
	backup   BackupMetrics
}

func NewMetrics(profile, group, version string, configLabels map[string]string) *Metrics {
	// default labels for all metrics
	labels := prometheus.Labels{profileLabel: profile}
	if group != "" {
		labels[groupLabel] = group
	}
	labels = mergeLabels(labels, configLabels)
	keys := maps.Keys(labels)

	registry := prometheus.NewRegistry()
	p := &Metrics{
		labels:   labels,
		registry: registry,
	}
	p.info = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "build_info",
		Help:      "resticprofile build information.",
	}, append(keys, goVersionLabel, versionLabel))
	// send the information about the build right away
	p.info.With(mergeLabels(cloneLabels(labels), map[string]string{goVersionLabel: runtime.Version(), versionLabel: version})).Set(1)

	p.backup = newBackupMetrics(keys)

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

func (p *Metrics) BackupResults(status Status, summary monitor.Summary) {
	p.backup.duration.With(p.labels).Set(summary.Duration.Seconds())

	p.backup.filesNew.With(p.labels).Set(float64(summary.FilesNew))
	p.backup.filesChanged.With(p.labels).Set(float64(summary.FilesChanged))
	p.backup.filesUnmodified.With(p.labels).Set(float64(summary.FilesUnmodified))

	p.backup.dirNew.With(p.labels).Set(float64(summary.DirsNew))
	p.backup.dirChanged.With(p.labels).Set(float64(summary.DirsChanged))
	p.backup.dirUnmodified.With(p.labels).Set(float64(summary.DirsUnmodified))

	p.backup.filesTotal.With(p.labels).Set(float64(summary.FilesTotal))
	p.backup.bytesAdded.With(p.labels).Set(float64(summary.BytesAdded))
	p.backup.bytesTotal.With(p.labels).Set(float64(summary.BytesTotal))
	p.backup.status.With(p.labels).Set(float64(status))
	p.backup.time.With(p.labels).Set(float64(time.Now().Unix()))
}

func (p *Metrics) SaveTo(filename string) error {
	return prometheus.WriteToTextfile(filename, p.registry)
}

func (p *Metrics) Push(url, format, jobName string) error {
	var expFmt expfmt.Format

	if format == "protobuf" {
		expFmt = expfmt.FmtProtoDelim
	} else {
		expFmt = expfmt.FmtText
	}

	return push.New(url, jobName).
		Format(expFmt).
		Gatherer(p.registry).
		Add()
}

func mergeLabels(labels prometheus.Labels, add map[string]string) prometheus.Labels {
	for key, value := range add {
		labels[key] = value
	}
	return labels
}

func cloneLabels(labels prometheus.Labels) prometheus.Labels {
	clone := make(prometheus.Labels, len(labels))
	for key, value := range labels {
		clone[key] = value
	}
	return clone
}
