---
title: "Status file"
slug: status
weight: 5
tags: [ "monitoring" ]
aliases:
  - /status
---

If you need to send your backup results to a monitoring system, use the `run-after` and `run-after-fail` scripts.

For simpler needs, resticprofile can generate a JSON file with details of the latest backup, forget, or check command. For example, I use a Zabbix agent to [check this file](https://github.com/creativeprojects/resticprofile/tree/master/contrib/zabbix) daily. Any monitoring system that reads JSON files can be integrated.

To enable this, add the status file location as a parameter in your profile.

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[profile]
  status-file = "backup-status.json"
```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

profile:
  status-file: backup-status.json
```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
"profile" {
  "status-file" = "backup-status.json"
}
```

{{% /tab %}}
{{% tab title="json" %}}

```json
{
  "version": "1",
  "profile": {
    "status-file": "backup-status.json"
  }
}
```

{{% /tab %}}
{{< /tabs >}}


Here is an example of a generated file showing the last `check` failed, while the last `backup` succeeded:

```json
{
  "profiles": {
    "self": {
      "backup": {
        "success": true,
        "time": "2021-03-24T16:36:56.831077Z",
        "error": "",
        "stderr": "",
        "duration": 16,
        "files_new": 215,
        "files_changed": 0,
        "files_unmodified": 0,
        "dirs_new": 58,
        "dirs_changed": 0,
        "dirs_unmodified": 0,
        "files_total": 215,
        "bytes_added": 296536447,
        "bytes_total": 362952485
      },
      "check": {
        "success": false,
        "time": "2021-03-24T15:23:40.270689Z",
        "error": "exit status 1",
        "stderr": "unable to create lock in backend: repository is already locked exclusively by PID 18534 on dingo by cloud_user (UID 501, GID 20)\nlock was created at 2021-03-24 15:23:29 (10.42277s ago)\nstorage ID 1bf636d2\nthe `unlock` command can be used to remove stale locks\n",
        "duration": 1
      }
    }
  }
}
```

## ⚠️ Extended status

In the backup section above, you can see fields like `files_new` and `files_total`. This information is available only when resticprofile's output is redirected or when the `extended-status` flag is added to your backup configuration.

This limitation ensures restic displays terminal output correctly.

The following fields do **not** require `extended-status` or stdout redirection:
- success
- time
- error
- stderr
- duration

The `extended-status` flag is **disabled by default because it suppresses restic's output**.

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[profile]
  status-file = "/home/backup/status.json"

  [profile.backup]
    extended-status = true
    source = "/go"
    exclude = [ "/**/.git/" ]
```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

profile:
  status-file: /home/backup/status.json
  backup:
    extended-status: true
    source: /go
    exclude:
      - "/**/.git/"

```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
"profile" = {
  "status-file" = "/home/backup/status.json"

  "backup" = {
    "extended-status" = true
    "source" = "/go"
    "exclude" = ["/**/.git/"]
  }
}
```

{{% /tab %}}
{{% tab title="json" %}}

```json
{
  "version": "1",
  "profile": {
    "status-file": "/home/backup/status.json",
    "backup": {
      "extended-status": true,
      "source": "/go",
      "exclude": [
        "/**/.git/"
      ]
    }
  }
}
```

{{% /tab %}}
{{< /tabs >}}
