---
title: "Prometheus"
slug: prometheus
weight: 10
tags: [ "monitoring" ]
---

Resticprofile can generate a Prometheus file or send the report to a Pushgateway. The `backup` command generates a detailed report, and the `check`, `copy`, `forget` and `prune` commands generate a short one (see [other commands](#other-commands)). Below is a configuration example for generating a file and sending it to a Pushgateway:

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

## Other commands

The `check`, `copy`, `forget` and `prune` commands report these metrics, where `<command>` is the name of the command:

| Metric | Description |
| ------ | ----------- |
| `resticprofile_<command>_duration_seconds` | Duration of the command (in seconds) |
| `resticprofile_<command>_exit_code` | Exit code of restic (`-1` when restic did not exit normally) |
| `resticprofile_<command>_status` | `0`=fail, `1`=warning, `2`=success |
| `resticprofile_<command>_time_seconds` | Last run of the command (unixtime) |

The exit code tells apart failures such as `11` (restic could not lock the repository) and `1` (fatal error, e.g. `check` found errors in the repository).

The `forget` metrics are also reported when the `retention` section runs before or after a backup, and the `check` metrics when `check-before` or `check-after` is set.

Example of the metrics added to the file after a `check` run:

```
# HELP resticprofile_check_duration_seconds The check duration (in seconds).
# TYPE resticprofile_check_duration_seconds gauge
resticprofile_check_duration_seconds{profile="prom"} 12.437316709
# HELP resticprofile_check_exit_code Exit code of restic check (-1 when restic did not exit normally).
# TYPE resticprofile_check_exit_code gauge
resticprofile_check_exit_code{profile="prom"} 0
# HELP resticprofile_check_status The check status: 0=fail, 1=warning, 2=success.
# TYPE resticprofile_check_status gauge
resticprofile_check_status{profile="prom"} 2
# HELP resticprofile_check_time_seconds Last check run (unixtime).
# TYPE resticprofile_check_time_seconds gauge
resticprofile_check_time_seconds{profile="prom"} 1.795479011e+09
```

### Several commands writing to the same file

When `prometheus-save-to-file` already exists, resticprofile updates it instead of replacing it, so running `copy` and then `check` on the same profile keeps the metrics of both commands:

- the metrics of the command that just ran replace its previous metrics,
- the metrics of the other commands stay in the file,
- the backup metrics and the build information are replaced as a whole, like before.

For `check`, `copy`, `forget` and `prune`, only the series with the same labels are replaced. This way, two kinds of checks of the same profile can be told apart with a label, for example a quick check and a check reading the data, using an environment variable set by the scheduler:

```yaml
version: "1"

root:
  prometheus-save-to-file: "root.prom"
  prometheus-labels:
    - check: '{{ or .Env.CHECK_KIND "quick" }}'
```

`CHECK_KIND=full resticprofile root.check --read-data` then reports `resticprofile_check_status{check="full",profile="root"}` next to the series of the quick check.

{{% notice style="note" %}}
A series stays in the file until a run with the same labels replaces it. Delete the file after changing `prometheus-labels` to remove the old series.
{{% /notice %}}

## Prometheus Pushgateway

Prometheus Pushgateway uses the job label as a grouping key. Metrics with the same grouping key are replaced when pushed. To prevent overwriting metrics from different profiles, the default job label is set to `<profile_name>.<command>` (e.g., `root.backup`).

For more control over the job label, use the `prometheus-push-job` property. This property supports the `$command` placeholder, which is replaced with the executed command's name. Each push only contains the metrics of the command that ran (plus the build information). The metrics of a `retention` section are pushed as the `forget` command (e.g. `root.forget`).

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