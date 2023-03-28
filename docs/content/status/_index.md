+++
chapter = true
pre = "<b>5. </b>"
title = "Status"
weight = 30
+++


# Status file for easy monitoring

If you need to escalate the result of your backup to a monitoring system, you can definitely use the `run-after` and `run-after-fail` scripting.

But sometimes we just need something simple that a monitoring system can regularly check. For that matter, resticprofile can generate a simple JSON file with the details of the latest backup/forget/check command. For example I have a Zabbix agent [checking this file](https://github.com/creativeprojects/resticprofile/tree/master/contrib/zabbix) once a day, and so you can hook up any monitoring system that can interpret a JSON file.

In your profile, you simply need to add a new parameter, which is the location of your status file

{{< tabs groupId="config-with-json" >}}
{{% tab name="toml" %}}

```toml
version = "1"

[profile]
  status-file = "backup-status.json"
```

{{% /tab %}}
{{% tab name="yaml" %}}

```yaml
version: "1"

profile:
  status-file: backup-status.json
```

{{% /tab %}}
{{% tab name="hcl" %}}

```hcl
"profile" {
  "status-file" = "backup-status.json"
}
```

{{% /tab %}}
{{% tab name="json" %}}

```json
{
  "version": "1",
  "profile": {
    "status-file": "backup-status.json"
  }
}
```

{{% /tab %}}
{{% /tabs %}}


Here's an example of a generated file, where you can see that the last `check` failed, whereas the last `backup` succeeded:

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

In the backup section above you can see some fields like `files_new`, `files_total`, etc. This information is only available when resticprofile's output is either *not* sent to the terminal (e.g. redirected) or when you add the flag `extended-status` to your backup configuration.

This is a technical limitation to ensure restic displays terminal output correctly. 

`extended-status` or stdout redirection is **not needed** for these fields:
- success
- time
- error
- stderr
- duration

`extended-status` is **not set by default because it hides any output from restic**

{{< tabs groupId="config-with-json" >}}
{{% tab name="toml" %}}

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
{{% tab name="yaml" %}}

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
{{% tab name="hcl" %}}

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
{{% tab name="json" %}}

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
{{% /tabs %}}
