---
title: "Monitoring - Temporary File"
weight: 53
---

This can be done by using the [template]({{% relref "/configuration/profiles/templates" %}}) function `tempFile`.

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
{{< /tabs >}}
