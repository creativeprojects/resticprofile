---
title: "Prometheus"
date: 2022-05-16T20:19:00+01:00
weight: 5
---



resticprofile can generate a prometheus file, or send the report to a push gateway. For now, only a `backup` command will generate a report.
Here's a configuration example with both options to generate a file and send to a push gateway:

{{< tabs groupId="config-with-json" >}}
{{% tab name="toml" %}}

```toml
[root]
  inherit = "default"
  prometheus-save-to-file = "root.prom"
  prometheus-push = "http://localhost:9091/"

  [root.backup]
    extended-status = true
    no-error-on-warning = true
    source = [ "/" ]
```

{{% /tab %}}
{{% tab name="yaml" %}}

```yaml
root:
  inherit: default
  prometheus-save-to-file: "root.prom"
  prometheus-push: "http://localhost:9091/"
  backup:
    extended-status: true
    no-error-on-warning: true
    source:
      - /
```

{{% /tab %}}
{{% tab name="hcl" %}}

```hcl
"root" = {
  "inherit" = "default"
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
{{% tab name="json" %}}

```json
{
  "root": {
    "inherit": "default",
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
{{% /tabs %}}

{{% notice style="note" %}}
Please note you need to set `extended-status` to `true` if you want all the available metrics. See [Extended status]({{< ref "/status/#-extended-status" >}}) for more information.
{{% /notice %}}

Here's an example of the generated prometheus file:

```
# HELP resticprofile_backup_added_bytes Total number of bytes added to the repository.
# TYPE resticprofile_backup_added_bytes gauge
resticprofile_backup_added_bytes{profile="root"} 35746
# HELP resticprofile_backup_dir_changed Number of directories with changes.
# TYPE resticprofile_backup_dir_changed gauge
resticprofile_backup_dir_changed{profile="root"} 9
# HELP resticprofile_backup_dir_new Number of new directories added to the backup.
# TYPE resticprofile_backup_dir_new gauge
resticprofile_backup_dir_new{profile="root"} 0
# HELP resticprofile_backup_dir_unmodified Number of directories unmodified since last backup.
# TYPE resticprofile_backup_dir_unmodified gauge
resticprofile_backup_dir_unmodified{profile="root"} 314
# HELP resticprofile_backup_duration_seconds The backup duration (in seconds).
# TYPE resticprofile_backup_duration_seconds gauge
resticprofile_backup_duration_seconds{profile="root"} 0.946567354
# HELP resticprofile_backup_files_changed Number of files with changes.
# TYPE resticprofile_backup_files_changed gauge
resticprofile_backup_files_changed{profile="root"} 3
# HELP resticprofile_backup_files_new Number of new files added to the backup.
# TYPE resticprofile_backup_files_new gauge
resticprofile_backup_files_new{profile="root"} 0
# HELP resticprofile_backup_files_processed Total number of files scanned by the backup for changes.
# TYPE resticprofile_backup_files_processed gauge
resticprofile_backup_files_processed{profile="root"} 3925
# HELP resticprofile_backup_files_unmodified Number of files unmodified since last backup.
# TYPE resticprofile_backup_files_unmodified gauge
resticprofile_backup_files_unmodified{profile="root"} 3922
# HELP resticprofile_backup_processed_bytes Total number of bytes scanned for changes.
# TYPE resticprofile_backup_processed_bytes gauge
resticprofile_backup_processed_bytes{profile="root"} 3.8524672e+07
# HELP resticprofile_backup_status Backup status: 0=fail, 1=warning, 2=success.
# TYPE resticprofile_backup_status gauge
resticprofile_backup_status{profile="root"} 1
# HELP resticprofile_build_info resticprofile build information.
# TYPE resticprofile_build_info gauge
resticprofile_build_info{goversion="go1.16.6",version="0.16.0"} 1

```

## User defined labels

You can add your own prometheus labels. Please note they will be applied to **all** the metrics.
Here's an example:

{{< tabs groupId="config-with-json" >}}
{{% tab name="toml" %}}

```toml
[root]
  inherit = "default"
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
{{% tab name="yaml" %}}

```yaml
root:
  inherit: default
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
{{% tab name="hcl" %}}

```hcl
"root" = {
  "inherit" = "default"
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
{{% tab name="json" %}}

```json
{
  "root": {
    "inherit": "default",
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
{{% /tabs %}}


which will add the `host` label to all your metrics.


