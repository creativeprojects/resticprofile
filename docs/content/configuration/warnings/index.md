---
title: "Warnings"
date: 2022-05-16T20:24:23+01:00
weight: 30
---

## Warnings from restic

Until version 0.13.0, resticprofile was always considering a restic warning as an error. This will remain the **default**.
But the version 0.13.0 introduced a parameter to avoid this behaviour and consider that the backup was successful instead.

A restic warning occurs when it cannot read some files, but a snapshot was successfully created.

### no-error-on-warning

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[profile]

  [profile.backup]
    no-error-on-warning = true

```

{{% /tab %}}
{{% tab title="yaml" %}}


```yaml
version: "1"

profile:
    backup:
        no-error-on-warning: true
```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
"profile" = {

  "backup" = {
    "no-error-on-warning" = true
  }
}
```

{{% /tab %}}
{{% tab title="json" %}}

```json
{
  "version": "1",
  "profile": {
    "backup": {
      "no-error-on-warning": true
    }
  }
}
```

{{% /tab %}}
{{% /tabs %}}