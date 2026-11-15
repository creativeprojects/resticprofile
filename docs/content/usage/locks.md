---
title: "Locks"
weight: 20
---



restic is already using a lock to avoid running some operations at the same time.

Since resticprofile can run several commands in a profile, it could be better to run the whole batch in a lock so nobody can interfere in the meantime.

For this to happen you can specify a lock file in each profile:

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[src]
  lock = "/tmp/resticprofile-profile-src.lock"

  [src.backup]
    check-before = true
    exclude = [ "/**/.git" ]
    source = [ "~/go" ]

  [src.retention]
    after-backup = true
    before-backup = false
    compact = false
    keep-within = "30d"
    prune = true
```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

src:
  lock: "/tmp/resticprofile-profile-src.lock"
  backup:
    check-before: true
    exclude:
    - /**/.git
    source:
    - ~/go
  retention:
    after-backup: true
    before-backup: false
    compact: false
    keep-within: 30d
    prune: true
```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
"src" = {
  "lock" = "/tmp/resticprofile-profile-src.lock"

  "backup" = {
    "check-before" = true
    "exclude" = ["/**/.git"]
    "source" = ["~/go"]
  }

  "retention" = {
    "after-backup" = true
    "before-backup" = false
    "compact" = false
    "keep-within" = "30d"
    "prune" = true
  }
}
```

{{% /tab %}}
{{% tab title="json" %}}

```json
{
  "version": "1",
  "src": {
    "lock": "/tmp/resticprofile-profile-src.lock",
    "backup": {
      "check-before": true,
      "exclude": [
        "/**/.git"
      ],
      "source": [
        "~/go"
      ]
    },
    "retention": {
      "after-backup": true,
      "before-backup": false,
      "compact": false,
      "keep-within": "30d",
      "prune": true
    }
  }
}
```

{{% /tab %}}
{{< /tabs >}}


For this profile, a lock will be set using the file `/tmp/resticprofile-profile-src.lock` for the duration of the profile: *check*, *backup* and *retention* (via the forget command)

**Please note restic locks and resticprofile locks are completely independent**

## Stale locks

In some cases, resticprofile as well as restic may leave a lock behind if the process died (or the machine rebooted).

For that matter, if you add the flag `force-inactive-lock` to your profile, resticprofile will detect and remove stale locks: 
* **resticprofile locks**: Check for the presence of a process with the PID indicated in the lockfile. If it can't find any, it will try to delete the lock and continue the operation (locking again, running profile and so on...)
* **restic locks**: Evaluate if a restic command failed on acquiring a lock. If the lock is older than `restic-stale-lock-age`, invoke `restic unlock` and retry the command that failed (can be disabled by setting `restic-stale-lock-age` to 0, default is 1h).

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[global]
  restic-stale-lock-age = "2h"

[src]
  lock = "/tmp/resticprofile-profile-src.lock"
  force-inactive-lock = true
```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

global:
  restic-stale-lock-age: 2h

src:
  lock: "/tmp/resticprofile-profile-src.lock"
  force-inactive-lock: true
```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
"global" = {
  "restic-stale-lock-age" = "2h"
}

"src" = {
  "lock" = "/tmp/resticprofile-profile-src.lock"
  "force-inactive-lock" = true
}
```

{{% /tab %}}
{{% tab title="json" %}}

```json
{
  "version": "1",
  "global": {
    "restic-stale-lock-age": "2h"
  },
  "src": {
    "lock": "/tmp/resticprofile-profile-src.lock",
    "force-inactive-lock": true
  }
}
```

{{% /tab %}}
{{< /tabs >}}

## Lock wait

By default, restic and resticprofile fail when a lock cannot be acquired as another process is currently holding it.

Depending on the use case (e.g. scheduled backups), it may be more appropriate to wait on another process to finish instead of failing immediately.

For that matter, if you add the commandline flag `--lock-wait` or configure schedules with `schedule-lock-wait`, resticprofile will wait on other backup processes:
* **resticprofile locks**: Retry acquiring the lockfile until it either succeeds (when the other resticprofile process released the lock) or fail as the lock-wait duration has passed without success.
* **restic locks**: Evaluate if a restic command failed on acquiring a lock. If the lock is not considered stale, retry the restic command every `restic-lock-retry-after` (default 1 minute) until it acquired the lock, or fail as the lock-wait duration has passed.

Note: The lock wait duration is cumulative. If various locks in one profile-run require lock wait, the total wait time may not exceed the duration that was specified. 

### Interactive sessions

`schedule-lock-mode` and `schedule-lock-wait` only apply when resticprofile is started by a schedule (through the [run-schedule]({{% relref "/schedules/commands/index.html#run-schedule-command" %}}) command). They are ignored when you run a profile yourself from the command line: without a lock wait duration, resticprofile fails immediately when another process is holding a lock.

This is usually what you want in an interactive session, since you are in front of the console to inspect the lock and remove it manually.

To wait for a lock in an interactive session, add the `--lock-wait` flag to the command line (or set the `RESTICPROFILE_LOCK_WAIT` environment variable):

```shell
resticprofile --lock-wait 15m profile.backup
```

The same goes for `restic-lock-retry-after` in the `global` section: a restic command that failed on acquiring a lock is only retried every `restic-lock-retry-after` when a lock wait duration is set, either by `--lock-wait` or by `schedule-lock-wait` in a scheduled run.

The [stale lock](#stale-locks) settings (`force-inactive-lock` and `restic-stale-lock-age`) apply to every run, scheduled or not: resticprofile removes a stale lock and retries the command, with or without a lock wait duration.

The `lock` file of the profile is also used in every run, unless you skip it with the `--no-lock` flag, or with `schedule-lock-mode` set to `ignore` in a scheduled run.

{{% notice style="tip" %}}
To test the lock settings of a schedule from the command line, run it the same way the scheduler does: `resticprofile run-schedule backup@profile`
{{% /notice %}}

## restic lock management

resticprofile can retry restic commands that fail on acquiring a lock and can also ask restic to unlock stale locks. The behaviour is controlled by 2 settings inside the `global` section:

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[global]
  # Retry a restic command that failed on acquiring a lock every minute 
  # (at least), for up to the time specified in "--lock-wait duration". 
  restic-lock-retry-after = "1m"
  # Ask restic to unlock a stale lock when its age is more than 2 hours
  # and the option "force-inactive-lock" is enabled in the profile.
  restic-stale-lock-age = "2h"
```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

global:
  # Retry a restic command that failed on acquiring a lock every minute 
  # (at least), for up to the time specified in "--lock-wait duration". 
  restic-lock-retry-after: 1m
  # Ask restic to unlock a stale lock when its age is more than 2 hours
  # and the option "force-inactive-lock" is enabled in the profile.
  restic-stale-lock-age: 2h
```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
"global" = {
  # Retry a restic command that failed on acquiring a lock every minute 
  # (at least), for up to the time specified in "--lock-wait duration". 
  "restic-lock-retry-after" = "1m"
  # Ask restic to unlock a stale lock when its age is more than 2 hours
  # and the option "force-inactive-lock" is enabled in the profile.
  "restic-stale-lock-age" = "2h"
}
```

{{% /tab %}}
{{% tab title="json" %}}

```json
{
  "version": "1",
  "global": {
    "restic-lock-retry-after": "1m",
    "restic-stale-lock-age": "2h"
  }
}
```

{{% /tab %}}
{{< /tabs >}}


If restic lock management is not desired, it can be disabled by setting both values to **0**.
