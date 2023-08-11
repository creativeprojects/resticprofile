---
title: "Logs"
tags: ["v0.21.0"]
weight: 45
---

By default **resticprofile** will display all logs (from itself and **restic**) to the console.

You can redirect the logs to a local file, a temporary file or a syslog server.

## Destination

The log destination syntax is a such:
* `filename` => redirects all the logs to the local file called **filename**
* `temp:filename` => redirects all the logs to a temporary file available during the whole session, and deleted afterwards.
* `tcp://syslog_server:514` or `udp://syslog_server:514` => redirects all the logs to the **syslog** server.

{{% notice tip %}}
If the location cannot be opened, **resticprofile** will default to send the logs to the console.
{{% /notice %}}

## Command line

You can redirect the logs on the command line with the `--log` flag:

```shell
resticprofile --log backup.log backup
```

## On a schedule

You can keep the logs displayed on the console when you run **resticprofile** commands manually, but send the logs somewhere else when **resticprofile** is started from a schedule.


{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[profile]
  [profile.backup]
    schedule = "*:00,30"
    schedule-priority = "background"
    schedule-log = "profile-backup.log"
```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

profile:
  backup:
    schedule: '*:00,30'
    schedule-priority: background
    schedule-log: profile-backup.log
```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
"profile" "backup" {
  "schedule" = "*:00,30"
  "schedule-priority" = "background"
  "schedule-log" = "profile-backup.log"
}
```

{{% /tab %}}
{{% tab title="json" %}}

```json
{
  "version": "1",
  "profile": {
    "backup": {
      "schedule": "*:00,30",
      "schedule-priority": "background",
      "schedule-log": "profile-backup.log"
    }
  }
}
```

{{% /tab %}}
{{% /tabs %}}

## Send logs to a temporary file

This can be done by using the [template]({{< ref "/configuration/templates" >}}) function `tempFile`.

This is to cover a special case when you want to upload the logs one by one to a remote location in a `run-finally` or a `run-after-fail` target.

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[backup_current]
  [backup_current.backup]
    verbose = true
    no-error-on-warning = true
    source = "{{ .CurrentDir }}"
    schedule = "*:44"
    schedule-log = '{{ tempFile "backup.log" }}'
    run-finally = 'cp {{ tempFile "backup.log" }} /logs/backup{{ .Now.Format "2006-01-02T15-04-05" }}.log'
```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

backup_current:
  backup:
    verbose: true
    no-error-on-warning: true
    source: "{{ .CurrentDir }}"
    schedule:
      - "*:44"
    schedule-log: '{{ tempFile "backup.log" }}'
    run-finally: 'cp {{ tempFile "backup.log" }} /logs/backup{{ .Now.Format "2006-01-02T15-04-05" }}.log'
```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
"profile" "backup" {
  "verbose" = true
  "no-error-on-warning" = true
  "source" = "{{ .CurrentDir }}"
  "schedule" = "*:44"
  "schedule-log" = "{{ tempFile "backup.log" }}"
  "run-finally" = "cp {{ tempFile "backup.log" }} /logs/backup{{ .Now.Format "2006-01-02T15-04-05" }}.log"
}
```

{{% /tab %}}
{{% tab title="json" %}}

```json
{
  "version": "1",
  "profile": {
    "backup": {
      "verbose": true,
      "no-error-on-warning": true,
      "source": "{{ .CurrentDir }}",
      "schedule": "*:44",
      "schedule-log": "{{ tempFile "backup.log" }}",
      "run-finally": "cp {{ tempFile "backup.log" }} /logs/backup{{ .Now.Format "2006-01-02T15-04-05" }}.log"
    }
  }
}
```

{{% /tab %}}
{{% /tabs %}}
