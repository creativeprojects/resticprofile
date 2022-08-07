+++
chapter = true
pre = "<b>4. </b>"
title = "Schedules"
weight = 20
+++


# Scheduled backups

resticprofile is capable of managing scheduled backups for you using:
- **launchd** on macOS X
- **Task Scheduler** on Windows
- **systemd** where available (Linux and other BSDs)
- **crond** on supported platforms (Linux and other BSDs)

On unixes (except macOS) resticprofile is using **systemd** by default. **crond** can be used instead if configured in `global` `scheduler` parameter:

{{< tabs groupId="config-with-json" >}}
{{% tab name="toml" %}}

```toml
[global]
  scheduler = "crond"
```

{{% /tab %}}
{{% tab name="yaml" %}}

```yaml
---
global:
    scheduler: crond
```

{{% /tab %}}
{{% tab name="hcl" %}}

```hcl
"global" = {
  "scheduler" = "crond"
}
```

{{% /tab %}}
{{% tab name="json" %}}

```json
{
  "global": {
    "scheduler": "crond"
  }
}
```

{{% /tab %}}
{{% /tabs %}}




Each profile can be scheduled independently (groups are not available for scheduling yet - it will be available in version '2' of the configuration file).

These 5 profile sections are accepting a schedule configuration:
- backup
- check
- forget (version 0.11.0)
- prune (version 0.11.0)
- copy (version 0.16.0)

which mean you can schedule `backup`, `forget`, `prune`, `check` and `copy` independently (I recommend to use a local `lock` in this case).

## retention schedule is deprecated
**Important**:
starting from version 0.11.0 the schedule of the `retention` section is **deprecated**: Use the `forget` section instead.

