---
title: "Scheduling"
weight: 20
alwaysopen: false
---

Here's an example of a scheduling configuration:

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
[default]
  repository = "d:\\backup"
  password-file = "key"

[self]
  inherit = "default"

  [self.retention]
    after-backup = true
    keep-within = "14d"

  [self.backup]
    source = "."
    schedule = [ "Mon..Fri *:00,15,30,45", "Sat,Sun 0,12:00" ]
    schedule-permission = "user"
    schedule-lock-wait = "10m"

  [self.prune]
    schedule = "sun 3:30"
    schedule-permission = "user"
    schedule-lock-wait = "1h"
```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
default:
  repository: "d:\\backup"
  password-file: key

self:
  inherit: default
  retention:
    after-backup: true
    keep-within: 14d
  backup:
    source: "."
    schedule:
      - "Mon..Fri *:00,15,30,45" # every 15 minutes on weekdays
      - "Sat,Sun 0,12:00"        # twice a day on week-ends
    schedule-permission: user
    schedule-lock-wait: 10m
  prune:
    schedule: "sun 3:30"
    schedule-permission: user
    schedule-lock-wait: 1h
```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
"default" = {
  "repository" = "d:\\backup"
  "password-file" = "key"
}

"self" = {
  "inherit" = "default"

  "retention" = {
    "after-backup" = true
    "keep-within" = "14d"
  }

  "backup" = {
    "source" = "."
    "schedule" = ["Mon..Fri *:00,15,30,45", "Sat,Sun 0,12:00"]
    "schedule-permission" = "user"
    "schedule-lock-wait" = "10m"
  }

  "prune" = {
    "schedule" = "sun 3:30"
    "schedule-permission" = "user"
    "schedule-lock-wait" = "1h"
  }
}
```

{{% /tab %}}
{{% tab title="json" %}}

```json
{
  "default": {
    "repository": "d:\\backup",
    "password-file": "key"
  },
  "self": {
    "inherit": "default",
    "retention": {
      "after-backup": true,
      "keep-within": "14d"
    },
    "backup": {
      "source": ".",
      "schedule": [
        "Mon..Fri *:00,15,30,45",
        "Sat,Sun 0,12:00"
      ],
      "schedule-permission": "user",
      "schedule-lock-wait": "10m"
    },
    "prune": {
      "schedule": "sun 3:30",
      "schedule-permission": "user",
      "schedule-lock-wait": "1h"
    }
  }
}
```

{{% /tab %}}
{{< /tabs >}}

For more examples, refer to:
{{% children sort="weight" %}}
