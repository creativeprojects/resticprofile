---
title: "Retry"
weight: 31
---

A backup to a remote repository can fail because of a temporary problem: the network is down for a minute, or the storage provider limits the rate of requests.
restic already retries failed requests to the repository for a while, but when the problem lasts longer, the whole command fails.

resticprofile can run the restic command again after such a failure. Set `retry` to the number of times to try again, and `retry-wait` to the time to wait before each new attempt (one minute by default).

These parameters are available in the `backup`, `check`, `copy`, `forget` and `prune` sections.

## Which failures are retried

resticprofile only runs the command again when restic exits with code **1** (a fatal error). Any other exit code means that running the same command again cannot help, so it is never retried:

| Exit code | Meaning | Retried |
|-----------|---------|---------|
| 1 | Fatal error, for example the repository cannot be reached | Yes |
| 2 | Go runtime error | No |
| 3 | The snapshot was created, but some source files could not be read | No (see [Warnings]({{% relref "/configuration/warnings" %}})) |
| 10 | The repository does not exist | No |
| 11 | The repository is locked | No (see [Locks]({{% relref "/usage/locks" %}})) |
| 12 | Wrong password | No |
| 130 | Interrupted | No |

A failure to lock the repository is handled by `lock-wait` and `restic-lock-retry-after` (see [Locks]({{% relref "/usage/locks" %}})), never by `retry`.

## What runs again

Only the restic command runs again. The `run-before` and `run-after` commands of the section are not repeated, and the `run-after-fail` and `send-after-fail` hooks only run once, after the last attempt failed.

A backup that reads its `stdin` directly cannot be retried, as the input can only be read once. With `stdin-command`, the command runs again for each attempt.

## Example

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[profile]

  [profile.backup]
    # run the backup again up to 3 times after a fatal error
    retry = 3
    # wait 5 minutes before each new attempt
    retry-wait = "5m"

```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

profile:
    backup:
        # run the backup again up to 3 times after a fatal error
        retry: 3
        # wait 5 minutes before each new attempt
        retry-wait: 5m
```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
"profile" = {

  "backup" = {
    # run the backup again up to 3 times after a fatal error
    "retry" = 3
    # wait 5 minutes before each new attempt
    "retry-wait" = "5m"
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
      "retry": 3,
      "retry-wait": "5m"
    }
  }
}
```

{{% /tab %}}
{{< /tabs >}}
