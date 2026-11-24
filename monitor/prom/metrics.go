package prom

import (
	"errors"
	"io/fs"
	"maps"
	"os"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/monitor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
)

const (
	namespace      = "resticprofile"
	backup         = "backup"
	groupLabel     = "group"
	profileLabel   = "profile"
	goVersionLabel = "goversion"
	versionLabel   = "version"
)

// commands reporting the generic command metrics (backup has its own metrics)
var commands = []string{
	constants.CommandCheck,
	constants.CommandCopy,
	constants.CommandForget,
	constants.CommandPrune,
}

type Metrics struct {
	labels     prometheus.Labels
	registry   *prometheus.Registry
	info       *prometheus.GaugeVec
	resticInfo *prometheus.GaugeVec
	backup     BackupMetrics
	commands   map[string]CommandMetrics
}

func NewMetrics(profile, group, version string, resticversion string, configLabels map[string]string) *Metrics {
	// default labels for all metrics
	labels := prometheus.Labels{profileLabel: profile}
	if group != "" {
		labels[groupLabel] = group
	}
	labels = mergeLabels(labels, configLabels)
	keys := slices.Collect(maps.Keys(labels))

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

	p.resticInfo = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_build_info",
		Help: "restic build information.",
	}, append(keys, versionLabel))
	// send the information about the build right away
	p.resticInfo.With(mergeLabels(cloneLabels(labels), map[string]string{versionLabel: resticversion})).Set(1)

	p.backup = newBackupMetrics(keys)

	registry.MustRegister(
		p.info,
		p.resticInfo,
		p.backup.duration,
		p.backup.filesNew,
		p.backup.filesChanged,
		p.backup.filesUnmodified,
		p.backup.dirNew,
		p.backup.dirChanged,
		p.backup.dirUnmodified,
		p.backup.filesTotal,
		p.backup.bytesAdded,
		p.backup.bytesAddedPacked,
		p.backup.bytesTotal,
		p.backup.status,
		p.backup.time,
	)

	p.commands = make(map[string]CommandMetrics, len(commands))
	for _, command := range commands {
		p.commands[command] = newCommandMetrics(command, keys)
		registry.MustRegister(p.commands[command].collectors()...)
	}
	return p
}

// HasCommand returns true if command reports metrics (other than backup)
func (p *Metrics) HasCommand(command string) bool {
	_, found := p.commands[command]
	return found
}

// CommandResults sets the metrics of a restic command other than backup
func (p *Metrics) CommandResults(command string, status Status, exitCode int, summary monitor.Summary) {
	metrics, found := p.commands[command]
	if !found {
		return
	}
	metrics.duration.With(p.labels).Set(summary.Duration.Seconds())
	metrics.exitCode.With(p.labels).Set(float64(exitCode))
	metrics.status.With(p.labels).Set(float64(status))
	metrics.time.With(p.labels).Set(float64(time.Now().Unix()))
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
	p.backup.bytesAddedPacked.With(p.labels).Set(float64(summary.BytesAddedPacked))
	p.backup.bytesTotal.With(p.labels).Set(float64(summary.BytesTotal))
	p.backup.status.With(p.labels).Set(float64(status))
	p.backup.time.With(p.labels).Set(float64(time.Now().Unix()))
}

// SaveTo writes the metrics to filename. Metrics already in the file are kept
// when this run doesn't report them, so commands of the same profile writing
// to the same file don't overwrite each other:
//   - the build info and the backup metrics replace the previous ones entirely,
//   - the metrics of other commands replace the series with the same labels only.
func (p *Metrics) SaveTo(filename string) error {
	return prometheus.WriteToTextfile(filename, prometheus.GathererFunc(func() ([]*dto.MetricFamily, error) {
		families, err := p.registry.Gather()
		if err != nil {
			return nil, err
		}
		previous, err := loadTextfile(filename)
		if err != nil {
			clog.Warningf("cannot read previous prometheus file %q, its content will be replaced: %v", filename, err)
			previous = nil
		}
		return mergeFamilies(previous, families), nil
	}))
}

// Push sends the metrics of command (and the build info) to the push gateway
func (p *Metrics) Push(url, format, jobName, command string) error {
	var expFmt expfmt.Format

	if format == "protobuf" {
		expFmt = expfmt.NewFormat(expfmt.TypeProtoDelim)
	} else {
		expFmt = expfmt.NewFormat(expfmt.TypeTextPlain)
	}

	return push.New(url, jobName).
		Format(expFmt).
		Gatherer(p.commandGatherer(command)).
		Add()
}

// commandGatherer only returns the build info and the metrics of command
func (p *Metrics) commandGatherer(command string) prometheus.Gatherer {
	prefix := namespace + "_" + command + "_"
	return prometheus.GathererFunc(func() ([]*dto.MetricFamily, error) {
		families, err := p.registry.Gather()
		families = slices.DeleteFunc(families, func(family *dto.MetricFamily) bool {
			return !isBuildInfo(family.GetName()) && !strings.HasPrefix(family.GetName(), prefix)
		})
		return families, err
	})
}

func isBuildInfo(name string) bool {
	return strings.HasSuffix(name, "build_info")
}

// mergesSeries returns true for the metrics of the commands other than backup
func mergesSeries(name string) bool {
	for _, command := range commands {
		if strings.HasPrefix(name, namespace+"_"+command+"_") {
			return true
		}
	}
	return false
}

// loadTextfile reads the metrics saved in filename (no metrics if the file doesn't exist)
func loadTextfile(filename string) (map[string]*dto.MetricFamily, error) {
	file, err := os.Open(filename)
	if errors.Is(err, fs.ErrNotExist) {
		return map[string]*dto.MetricFamily{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	parser := expfmt.NewTextParser(model.UTF8Validation)
	return parser.TextToMetricFamilies(file)
}

// mergeFamilies adds the previous metric families and series not present in current
func mergeFamilies(previous map[string]*dto.MetricFamily, current []*dto.MetricFamily) []*dto.MetricFamily {
	merged := slices.Clone(current)
	for _, family := range current {
		old, found := previous[family.GetName()]
		if !found || !mergesSeries(family.GetName()) || old.GetType() != family.GetType() {
			continue
		}
		seen := make(map[string]bool, len(family.GetMetric()))
		for _, metric := range family.GetMetric() {
			seen[labelsKey(metric)] = true
		}
		for _, metric := range old.GetMetric() {
			if !seen[labelsKey(metric)] {
				family.Metric = append(family.Metric, metric)
			}
		}
		slices.SortFunc(family.Metric, func(a, b *dto.Metric) int {
			return strings.Compare(labelsKey(a), labelsKey(b))
		})
	}
	for name, family := range previous {
		if !slices.ContainsFunc(current, func(f *dto.MetricFamily) bool { return f.GetName() == name }) {
			merged = append(merged, family)
		}
	}
	slices.SortFunc(merged, func(a, b *dto.MetricFamily) int {
		return strings.Compare(a.GetName(), b.GetName())
	})
	return merged
}

func labelsKey(metric *dto.Metric) string {
	pairs := make([]string, 0, len(metric.GetLabel()))
	for _, label := range metric.GetLabel() {
		pairs = append(pairs, label.GetName()+"\xff"+label.GetValue())
	}
	slices.Sort(pairs)
	return strings.Join(pairs, "\xfe")
}

func mergeLabels(labels prometheus.Labels, add map[string]string) prometheus.Labels {
	maps.Copy(labels, add)
	return labels
}

func cloneLabels(labels prometheus.Labels) prometheus.Labels {
	clone := make(prometheus.Labels, len(labels))
	maps.Copy(clone, labels)
	return clone
}
