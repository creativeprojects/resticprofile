---
title: "Sending to healthchecks.io"
weight: 20
---

For more information, refer to [Hooks - HTTP]({{% relref "configuration/hooks/http_hooks/#send-http-messages-before-and-after-a-job" %}})

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[profile]

  [profile.backup]
  source = "/source"
  exclude = [ "/**/.git/" ]
  schedule = [ "*:00,30" ]
  schedule-permission = "user"

    # you can have more than one target

    [[profile.backup.send-before]]
    method = "HEAD"
    url = "https://hc-ping.com/831e288e-1293-46f8-ac31-70ea7f875650/start"

    [[profile.backup.send-before]]
    method = "HEAD"
    url = "https://httpstat.us/400"

    # you can have more than one target

    [[profile.backup.send-after]]
    method = "HEAD"
    url = "https://hc-ping.com/831e288e-1293-46f8-ac31-70ea7f875650"

    [[profile.backup.send-after]]
    method = "HEAD"
    url = "https://httpstat.us/500"

    [profile.backup.send-after-fail]
    method = "POST"
    url = "https://hc-ping.com/831e288e-1293-46f8-ac31-70ea7f875650/fail"
    body = "${ERROR}\n\n${ERROR_STDERR}"

      [[profile.backup.send-after-fail.headers]]
      name = "Content-Type"
      value = "text/plain; charset=UTF-8"

  [profile.check]
  schedule = [ "*:15" ]

    [profile.check.send-before]
    method = "HEAD"
    url = "https://hc-ping.com/e0f62e41-b75f-450f-8cdd-7f25e466d2dc/start"

    [profile.check.send-after]
    method = "HEAD"
    url = "https://hc-ping.com/e0f62e41-b75f-450f-8cdd-7f25e466d2dc"

  [profile.retention]
  after-backup = true


```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

profile:

    backup:
        source: "/source"
        exclude:
          - "/**/.git/"
        schedule:
          - "*:00,30"
        schedule-permission: user

        # you can have more than one target
        send-before:
          - method: HEAD
            url: https://hc-ping.com/831e288e-1293-46f8-ac31-70ea7f875650/start
          - method: HEAD
            url: https://httpstat.us/400

        # you can have more than one target
        send-after:
          - method: HEAD
            url: https://hc-ping.com/831e288e-1293-46f8-ac31-70ea7f875650
          - method: HEAD
            url: https://httpstat.us/500

        send-after-fail:
            method: POST
            url: https://hc-ping.com/831e288e-1293-46f8-ac31-70ea7f875650/fail
            body: "${ERROR}\n\n${ERROR_STDERR}"
            headers:
              - name: Content-Type
                value: "text/plain; charset=UTF-8"
    check:
        schedule:
          - "*:15"

        send-before:
          method: HEAD
          url: https://hc-ping.com/e0f62e41-b75f-450f-8cdd-7f25e466d2dc/start

        send-after:
          method: HEAD
          url: https://hc-ping.com/e0f62e41-b75f-450f-8cdd-7f25e466d2dc
          
    retention:
        after-backup: true

```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
"profile" {

  "backup" = {
    "source" = "/source"
    "exclude" = ["/**/.git/"]
    "schedule" = ["*:00,30"]
    "schedule-permission" = "user"

    "send-before" = {
      "method" = "HEAD"
      "url" = "https://hc-ping.com/831e288e-1293-46f8-ac31-70ea7f875650/start"
    }

    "send-before" = {
      "method" = "HEAD"
      "url" = "https://httpstat.us/400"
    }

    "send-after" = {
      "method" = "HEAD"
      "url" = "https://hc-ping.com/831e288e-1293-46f8-ac31-70ea7f875650"
    }

    "send-after" = {
      "method" = "HEAD"
      "url" = "https://httpstat.us/500"
    }

    "send-after-fail" = {
      "method" = "POST"
      "url" = "https://hc-ping.com/831e288e-1293-46f8-ac31-70ea7f875650/fail"
      "body" = "${ERROR}\n\n${ERROR_STDERR}"
      "headers" = {
        "name" = "Content-Type"
        "value" = "text/plain; charset=UTF-8"
      }
    }
  }

  "check" = {
    "schedule" = ["*:15"]

    "send-before" = {
      "method" = "HEAD"
      "url" = "https://hc-ping.com/e0f62e41-b75f-450f-8cdd-7f25e466d2dc/start"
    }

    "send-after" = {
      "method" = "HEAD"
      "url" = "https://hc-ping.com/e0f62e41-b75f-450f-8cdd-7f25e466d2dc"
    }
  }

  "retention" = {
    "after-backup" = true
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
      "exclude": [
        "/**/.git/"
      ],
      "schedule": [
        "*:00,30"
      ],
      "schedule-permission": "user",
      "send-before": [
        {
          "method": "HEAD",
          "url": "https://hc-ping.com/831e288e-1293-46f8-ac31-70ea7f875650/start"
        },
        {
          "method": "HEAD",
          "url": "https://httpstat.us/400"
        }
      ],
      "send-after": [
        {
          "method": "HEAD",
          "url": "https://hc-ping.com/831e288e-1293-46f8-ac31-70ea7f875650"
        },
        {
          "method": "HEAD",
          "url": "https://httpstat.us/500"
        }
      ],
      "send-after-fail": {
        "method": "POST",
        "url": "https://hc-ping.com/831e288e-1293-46f8-ac31-70ea7f875650/fail",
        "body": "${ERROR}\n\n${ERROR_STDERR}",
        "headers": [
          {
            "name": "Content-Type",
            "value": "text/plain; charset=UTF-8"
          }
        ]
      }
    },
    "check": {
      "schedule": [
        "*:15"
      ],
      "send-before": {
        "method": "HEAD",
        "url": "https://hc-ping.com/e0f62e41-b75f-450f-8cdd-7f25e466d2dc/start"
      },
      "send-after": {
        "method": "HEAD",
        "url": "https://hc-ping.com/e0f62e41-b75f-450f-8cdd-7f25e466d2dc"
      }
    },
    "retention": {
      "after-backup": true
    }
  }
}
```

{{% /tab %}}
{{< /tabs >}}

