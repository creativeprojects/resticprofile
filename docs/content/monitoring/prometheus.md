---
title: "Prometheus"
slug: prometheus
weight: 10
tags: [ "monitoring" ]
---

Resticprofile can generate a Prometheus file or send the report to a Pushgateway. Currently, only the `backup` command generates a report. Below is a configuration example for generating a file and sending it to a Pushgateway:

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[root]
  prometheus-save-to-file = "root.prom"
  prometheus-push = "http://localhost:9091/"

  [root.backup]
    extended-status = true
    no-error-on-warning = true
    source = [ "/" ]
```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

root:
  prometheus-save-to-file: "root.prom"
  prometheus-push: "http://localhost:9091/"
  backup:
    extended-status: true
    no-error-on-warning: true
    source:
      - /
```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
"root" = {
  "prometheus-save-to-file" = "root.prom"
  "prometheus-push" = "http://localhost:9091/"

  "backup" = {
    "extended-status" = true
    "no-error-on-warning" = true
    "source" = ["/"]
  }
}
```

{{% /tab %}}
{{% tab title="json" %}}

```json
{
  "version": "1",
  "root": {
    "prometheus-save-to-file": "root.prom",
    "prometheus-push": "http://localhost:9091/",
    "backup": {
      "extended-status": true,
      "no-error-on-warning": true,
      "source": [
        "/"
      ]
    }
  }
}
```

{{% /tab %}}
{{< /tabs >}}

{{% notice style="note" %}}
Set `extended-status` to `true` to access all available metrics. For details, see [Extended status]({{% relref "/monitoring/status/index.html#-extended-status" %}}).
{{% /notice %}}

Here's an example of a generated prometheus file:

```
# HELP restic_build_info restic build information.
# TYPE restic_build_info gauge
restic_build_info{profile="prom",version="0.18.0"} 1
# HELP resticprofile_backup_added_bytes Total number of bytes added to the repository.
# TYPE resticprofile_backup_added_bytes gauge
resticprofile_backup_added_bytes{profile="prom"} 96167
# HELP resticprofile_backup_dir_changed Number of directories with changes.
# TYPE resticprofile_backup_dir_changed gauge
resticprofile_backup_dir_changed{profile="prom"} 8
# HELP resticprofile_backup_dir_new Number of new directories added to the backup.
# TYPE resticprofile_backup_dir_new gauge
resticprofile_backup_dir_new{profile="prom"} 0
# HELP resticprofile_backup_dir_unmodified Number of directories unmodified since last backup.
# TYPE resticprofile_backup_dir_unmodified gauge
resticprofile_backup_dir_unmodified{profile="prom"} 1060
# HELP resticprofile_backup_duration_seconds The backup duration (in seconds).
# TYPE resticprofile_backup_duration_seconds gauge
resticprofile_backup_duration_seconds{profile="prom"} 0.986296416
# HELP resticprofile_backup_files_changed Number of files with changes.
# TYPE resticprofile_backup_files_changed gauge
resticprofile_backup_files_changed{profile="prom"} 2
# HELP resticprofile_backup_files_new Number of new files added to the backup.
# TYPE resticprofile_backup_files_new gauge
resticprofile_backup_files_new{profile="prom"} 0
# HELP resticprofile_backup_files_processed Total number of files scanned by the backup for changes.
# TYPE resticprofile_backup_files_processed gauge
resticprofile_backup_files_processed{profile="prom"} 7723
# HELP resticprofile_backup_files_unmodified Number of files unmodified since last backup.
# TYPE resticprofile_backup_files_unmodified gauge
resticprofile_backup_files_unmodified{profile="prom"} 7721
# HELP resticprofile_backup_processed_bytes Total number of bytes scanned for changes.
# TYPE resticprofile_backup_processed_bytes gauge
resticprofile_backup_processed_bytes{profile="prom"} 2.935621558e+09
# HELP resticprofile_backup_status Backup status: 0=fail, 1=warning, 2=success.
# TYPE resticprofile_backup_status gauge
resticprofile_backup_status{profile="prom"} 2
# HELP resticprofile_backup_time_seconds Last backup run (unixtime).
# TYPE resticprofile_backup_time_seconds gauge
resticprofile_backup_time_seconds{profile="prom"} 1.747673785e+09
# HELP resticprofile_build_info resticprofile build information.
# TYPE resticprofile_build_info gauge
resticprofile_build_info{goversion="go1.24.3",profile="prom",version="0.31.0"} 1


```

## Prometheus Pushgateway

Prometheus Pushgateway uses the job label as a grouping key. Metrics with the same grouping key are replaced when pushed. To prevent overwriting metrics from different profiles, the default job label is set to `<profile_name>.<command>` (e.g., `root.backup`).

For more control over the job label, use the `prometheus-push-job` property. This property supports the `$command` placeholder, which is replaced with the executed command's name.

You can specify the request format using `prometheus-push-format`. The default is `text`, but it can also be set to `protobuf` (see [compatibility with Prometheus](https://prometheus.io/docs/instrumenting/exposition_formats/)).

## User-Defined Labels

You can add custom Prometheus labels, which will apply to **all** metrics. Example:

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[root]
  prometheus-save-to-file = "root.prom"
  prometheus-push = "http://localhost:9091/"

  [[root.prometheus-labels]]
    host = "{{ .Hostname }}"

  [root.backup]
    extended-status = true
    no-error-on-warning = true
    source = [ "/" ]
```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

root:
  prometheus-save-to-file: "root.prom"
  prometheus-push: "http://localhost:9091/"
  prometheus-labels:
    - host: {{ .Hostname }}
  backup:
    extended-status: true
    no-error-on-warning: true
    source:
      - /
```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
"root" = {
  "prometheus-save-to-file" = "root.prom"
  "prometheus-push" = "http://localhost:9091/"

  "prometheus-labels" = {
    "host" = "{{ .Hostname }}"
  }

  "backup" = {
    "extended-status" = true
    "no-error-on-warning" = true
    "source" = ["/"]
  }
}
```

{{% /tab %}}
{{% tab title="json" %}}

```json
{
  "version": "1",
  "root": {
    "prometheus-save-to-file": "root.prom",
    "prometheus-push": "http://localhost:9091/",
    "prometheus-labels": [
      {
        "host": "{{ .Hostname }}"
      }
    ],
    "backup": {
      "extended-status": true,
      "no-error-on-warning": true,
      "source": [
        "/"
      ]
    }
  }
}
```

{{% /tab %}}
{{< /tabs >}}


This adds the `host` label to all your metrics.

## Exposing metrics over HTTP [EXPERIMENTAL]

In addition to the file-based and push-based options, resticprofile can serve the latest metrics on an HTTP endpoint. The endpoint combines the Go runtime collector with the profile's `prometheus-save-to-file` body so a single sidecar can be scraped by Prometheus, Node Exporter's `--collector.textfile.directory`, or anything else that understands the Prometheus text exposition format.

The feature is opt-in and hidden from `--help` until it stabilises. Enable it via the `--metrics-port` flag (or the `RESTICPROFILE_METRICS_PORT` environment variable) and select one of the following integration points:

### `schedule` keeps the process alive and serves `/metrics`

When `resticprofile` is used as the scheduler under `crond` (e.g. inside a container), it can double as a metrics sidecar. The scheduler registers all cron jobs as usual, then blocks on SIGINT/SIGTERM while serving `/metrics` for each profile that has `prometheus-save-to-file` set. Textfile paths are de-duplicated across profiles and groups, so several profiles writing to the same file share a single scrape target.

```sh
resticprofile --metrics-port 19302 schedule --all &
crond -f
```

### `serve-metrics` as a stand-alone sidecar

If you don't use `schedule` (or already have a separate scheduler), run a dedicated process:

```sh
resticprofile --metrics-port 19302 default serve-metrics
```

### Response contract

- `200 OK` — Go runtime metrics followed by the textfile body.
- `503 Service Unavailable` — the textfile does not exist yet (e.g. the first backup has not run). Prometheus reports the target as `up == 0` without logging a 404 on every scrape.
- `500 Internal Server Error` — other I/O errors.

The server binds on `0.0.0.0:<port>` for `network_mode: host` containers.

### Scrape configuration

```yaml
scrape_configs:
  - job_name: resticprofile
    static_configs:
      - targets: ['<host>:19302']
        labels:
          host: backup-server
```