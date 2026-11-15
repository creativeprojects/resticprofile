---
title: "Sending to Watchgoose"
weight: 25
tags: [ "monitoring", "watchgoose.com" ]
---

For more information, refer to [Hooks - HTTP]({{% relref "configuration/hooks/http_hooks/#send-http-messages-before-and-after-a-job" %}})

[Watchgoose](https://watchgoose.com/docs/) gives each check a ping URL: send `/start` when the job starts, the bare URL on success, and `/fail` on failure. Replace the UUID with the one of your check. The first 10,000 bytes of a POST body are stored as the ping log, so the `send-after-fail` hook below attaches the error message and the restic error output to the failure.

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[profile]

  [profile.backup]
  source = "/source"
  schedule = [ "*:00,30" ]
  schedule-permission = "user"

    [profile.backup.send-before]
    method = "HEAD"
    url = "https://watchgoose.com/ping/addffa72-da17-40ae-be9c-ff591afb942a/start"

    [profile.backup.send-after]
    method = "HEAD"
    url = "https://watchgoose.com/ping/addffa72-da17-40ae-be9c-ff591afb942a"

    [profile.backup.send-after-fail]
    method = "POST"
    url = "https://watchgoose.com/ping/addffa72-da17-40ae-be9c-ff591afb942a/fail"
    body = "${ERROR}\n\n${ERROR_STDERR}"

      [[profile.backup.send-after-fail.headers]]
      name = "Content-Type"
      value = "text/plain; charset=UTF-8"
```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

profile:

    backup:
        source: "/source"
        schedule:
          - "*:00,30"
        schedule-permission: user

        send-before:
            method: HEAD
            url: https://watchgoose.com/ping/addffa72-da17-40ae-be9c-ff591afb942a/start

        send-after:
            method: HEAD
            url: https://watchgoose.com/ping/addffa72-da17-40ae-be9c-ff591afb942a

        send-after-fail:
            method: POST
            url: https://watchgoose.com/ping/addffa72-da17-40ae-be9c-ff591afb942a/fail
            body: "${ERROR}\n\n${ERROR_STDERR}"
            headers:
              - name: Content-Type
                value: "text/plain; charset=UTF-8"
```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
"profile" {

  "backup" = {
    "source" = "/source"
    "schedule" = ["*:00,30"]
    "schedule-permission" = "user"

    "send-before" = {
      "method" = "HEAD"
      "url" = "https://watchgoose.com/ping/addffa72-da17-40ae-be9c-ff591afb942a/start"
    }

    "send-after" = {
      "method" = "HEAD"
      "url" = "https://watchgoose.com/ping/addffa72-da17-40ae-be9c-ff591afb942a"
    }

    "send-after-fail" = {
      "method" = "POST"
      "url" = "https://watchgoose.com/ping/addffa72-da17-40ae-be9c-ff591afb942a/fail"
      "body" = "${ERROR}\n\n${ERROR_STDERR}"
      "headers" = {
        "name" = "Content-Type"
        "value" = "text/plain; charset=UTF-8"
      }
    }
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
      "source": "/source",
      "schedule": [
        "*:00,30"
      ],
      "schedule-permission": "user",
      "send-before": {
        "method": "HEAD",
        "url": "https://watchgoose.com/ping/addffa72-da17-40ae-be9c-ff591afb942a/start"
      },
      "send-after": {
        "method": "HEAD",
        "url": "https://watchgoose.com/ping/addffa72-da17-40ae-be9c-ff591afb942a"
      },
      "send-after-fail": {
        "method": "POST",
        "url": "https://watchgoose.com/ping/addffa72-da17-40ae-be9c-ff591afb942a/fail",
        "body": "${ERROR}\n\n${ERROR_STDERR}",
        "headers": [
          {
            "name": "Content-Type",
            "value": "text/plain; charset=UTF-8"
          }
        ]
      }
    }
  }
}
```

{{% /tab %}}
{{< /tabs >}}

To record the restic exit code on the check instead of a plain failure, use `https://watchgoose.com/ping/addffa72-da17-40ae-be9c-ff591afb942a/${ERROR_EXIT_CODE}` as the `send-after-fail` URL.
