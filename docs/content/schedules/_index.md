+++
archetype = "chapter"
pre = "<b>4. </b>"
title = "Schedules"
weight = 4
+++


## Scheduler

resticprofile manages scheduled backups using:
- **[launchd]({{% relref "/schedules/launchd" %}})** on macOS
- **[Task Scheduler]({{% relref "/schedules/task_scheduler" %}})** on Windows
- **[systemd]({{% relref "/schedules/systemd" %}})** on Linux and other BSDs
- **[crond]({{% relref "/schedules/cron" %}})** as a fallback (requires `crontab` binary)
- **[crontab]({{% relref "/schedules/cron" %}})** files (with or without a user column)

On Unix systems (excluding macOS), resticprofile uses **systemd** if available, otherwise it falls back to **crond**.

See [reference / global section]({{% relref "/reference/global" %}}) for scheduler configuration options.

Each profile can be scheduled independently. Within each profile, these sections can be scheduled:
- **backup**
- **check**
- **forget**
- **prune**
- **copy**

## Deprecation
Scheduling the `retention` section directly is **deprecated**. Use the `forget` section instead.

The retention section should be associated with a `backup` section, not scheduled independently.

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
[profile.retention]
  # deprecated
  schedule = "daily"

# use the forget target instead
[profile.forget]
  schedule = "daily"

```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
---
profile:
  retention:
    # deprecated
    schedule: daily

  # use the forget target instead
  forget:
    schedule: daily
```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
"profile" = {
  "retention" = {
     # deprecated
    schedule = "daily"
  }

  # use the forget target instead
  "forget" = {
    schedule = "daily"
  }
}
```

{{% /tab %}}
{{% tab title="json" %}}

```json
{
  "profile": {
    "forget": {
      "schedule": "daily"
    }
  }
}
```

{{% /tab %}}
{{< /tabs >}}


## More information

{{% children  %}}
