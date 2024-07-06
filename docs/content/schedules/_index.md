+++
archetype = "chapter"
pre = "<b>4. </b>"
title = "Schedules"
weight = 4
+++


resticprofile is capable of managing scheduled backups for you. Under the hood it's using:
- **launchd** on macOS X
- **Task Scheduler** on Windows
- **systemd** where available (Linux and other BSDs)
- **crond** as fallback (depends on the availability of a `crontab` binary)
- **crontab** files (low level, with (`*`) or without (`-`) user column)

On unixes (except macOS) resticprofile is using **systemd** if available and falls back to **crond**. 
On any OS a **crond** compatible scheduler can be used instead if configured in `global` / `scheduler`:

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
[global]
  scheduler = "crond"
  # scheduler = "crond:/usr/bin/crontab"
  # scheduler = "crontab:*:/etc/cron.d/resticprofile"
  # scheduler = "crontab:-:/var/spool/cron/crontabs/username"
```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
---
global:
    scheduler: crond
    # scheduler: "crond:/usr/bin/crontab"
    # scheduler: "crontab:*:/etc/cron.d/resticprofile"
    # scheduler: "crontab:-:/var/spool/cron/crontabs/username"
```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
"global" = {
  "scheduler" = "crond"
  # "scheduler" = "crond:/usr/bin/crontab"
  # "scheduler" = "crontab:*:/etc/cron.d/resticprofile"
  # "scheduler" = "crontab:-:/var/spool/cron/crontabs/username"
}
```

{{% /tab %}}
{{% tab title="json" %}}

```json
{
  "global": {
    "scheduler": "crond"
  }
}
```

{{% /tab %}}
{{< /tabs >}}

See also [reference / global section]({{% relref "/reference/global" %}}) for options on how to configure the scheduler.


Each profile can be scheduled independently (groups are not available for scheduling yet - it will be available in version '2' of the configuration file).

These 5 profile sections are accepting a schedule configuration:
- backup
- check
- forget (version 0.11.0)
- prune (version 0.11.0)
- copy (version 0.16.0)

which mean you can schedule `backup`, `forget`, `prune`, `check` and `copy` independently (I recommend using a [local lock]({{% relref "/usage/locks" %}}) in this case).

## retention schedule is deprecated
Starting from version 0.11.0, directly scheduling the `retention` section is **deprecated**: Use the `forget` section for direct schedule instead.

The retention section is designed to be associated with a `backup` section, not to be scheduled independently.

