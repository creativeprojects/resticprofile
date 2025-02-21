---
title: "HTTP Hooks"
tags: ["v0.18.0"]
weight: 22
---


## Send HTTP messages before and after a job

As well as being able to run [shell commands]({{% relref "run_hooks" %}}), you can now send HTTP messages before, after (success or failure) running a restic command.

The sections that allow sending HTTP hooks are:
- backup
- copy
- check
- forget
- prune

{{% notice tip %}}
You might notice that's the same sections that can also be scheduled
{{% /notice %}}

Each of these commands can send 4 different types of hooks:

- send-before
- send-after
- send-after-fail
- send-finally

The configuration is the same for each of these 4 types of hooks:

| Name | Required | Default | Notes |
|:-----|:---------|:--------|:------|
| url | Yes | None | URL of your Webhook |
| method | No | GET | This is the HTTP method (GET, POST, HEAD, etc.) |
| skip-tls-verification | No | False | **This is not recommended**: Use only if you're using your own server with a self-signed certificate |
| headers | No | User-Agent set to resticprofile | This is a subsection with a list of `name` and `value` |
| body | No | Empty | Used to send data to the Webhook (POST, PUT, PATCH) |
| body-template | No | None | Template file to generate the body (in go template format) |


### Example sending monitoring information to healthchecks.io:

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


A few environment variables will be available to construct the url and the body:
- `PROFILE_NAME`
- `PROFILE_COMMAND`: backup, check, forget, etc.

Additionally, for the `send-after-fail` hooks, these environment variables will be available:
- `ERROR` containing the latest error message
- `ERROR_COMMANDLINE` containing the command line that failed
- `ERROR_EXIT_CODE` containing the exit code of the command line that failed
- `ERROR_STDERR` containing any message that the failed command sent to the standard error (stderr)

The `send-finally` hooks are also getting the environment of `send-after-fail` when any previous operation has failed (except any `send` operation).

Failures in any `send-*` are logged but do not influence environment or return code.

### order of `send-*`

Here's the flow of HTTP hooks:

```mermaid
graph TD
  LOCK(set resticprofile lock)
  UNLOCK(delete resticprofile lock)
  LOCK --> SB
  SB('send-before') --> RUN
  RUN(run restic command, or group of commands)
  RUN -->|Success| SA
  RUN -->|Error| SAF
  SA('send-after') --> SF
  SAF('send-after-fail') --> SF
  SF('send-finally')
  SF --> UNLOCK
```

{{% notice style="warning" title="resticprofile lock" %}}
The local resticprofile lock is surrounding the whole process. It means that the `run-after-fail` target is not called if the lock cannot be obtained. This is a limitation of the current implementation. 
{{% /notice %}}


### body-template

You can use a standard go template to build the webhook body. It has to be defined in a separate file (otherwise it would clash with the configuration as a go template itself).

The object passed as an argument to the template is:

- `ProfileName`    **string**
- `ProfileCommand` **string**
- `Error`          **ErrorContext**
- `Stdout`         **string**

The type **ErrorContext** is available after an error occurred (otherwise all fields are blank):
- `Message`     **string**
- `CommandLine` **string**
- `ExitCode`    **string**
- `Stderr`      **string**

Here's an example of a body file:

<!-- checkdoc-ignore -->
```json
{
  "profileName": "{{ .ProfileName }}",
  "profileCommand": "{{ .ProfileCommand }}",
  "exitCode": "{{ .Error.ExitCode }}"
}
```

The field `exitCode` will be blank if no error occured.

And here's an example of a configuration using a body template:


{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[profile]

  [profile.backup]
  source = "/source"

    [profile.backup.send-finally]
    method = "POST"
    url = "https://my/monitoring.example.com/"
    body-template = "body-template.json"

      [[profile.backup.send-finally.headers]]
      name = "Content-Type"
      value = "application/json"

```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

profile:

    backup:
        source: "/source"

        send-finally:
            method: POST
            url: https://my/monitoring.example.com/
            body-template: body-template.json
            headers:
              - name: Content-Type
                value: "application/json"


```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
"profile" {

  "backup" = {
    "source" = "/source"

    "send-finally" = {
      "method" = "POST"
      "url" = "https://my/monitoring.example.com/"
      "body-template" = "body-template.json"
      "headers" = {
        "name" = "Content-Type"
        "value" = "application/json"
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
      "send-finally": {
        "method": "POST",
        "url": "https://my/monitoring.example.com/",
        "body-template": "body-template.json",
        "headers": [
          {
            "name": "Content-Type",
            "value": "application/json"
          }
        ]
      }
    }
  }
}
```

{{% /tab %}}
{{< /tabs >}}

### Self-signed certificates

If your monitoring system is using self-signed certificates, you can import them in resticprofile (and you don't need to rely on the `skip-tls-verification` flag)

The parameter is in the `global` section and is called `ca-certificates`: it contains a list of certificate files (PEM).

### timeout

The default timeout for all HTTP requests is 30 seconds.

You can change the default timeout in the `global` section with a parameter called `send-timeout`.

The format is like:
- 30s
- 2m
- 1m20s

### global configuration example


{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[global]
  send-timeout = "10s"
  ca-certificates = [ "ca-chain.cert.pem" ]
```


{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

global:
  send-timeout: 10s
  ca-certificates:
    - ca-chain.cert.pem

```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl

global {
  send-timeout = "10s"
  ca-certificates = "ca-chain.cert.pem"
}
```

{{% /tab %}}
{{% tab title="json" %}}

```json
{
  "version": "1",
  "global": {
    "send-timeout": "10s",
    "ca-certificates": [
      "ca-chain.cert.pem"
    ]
  }
}
```

{{% /tab %}}
{{< /tabs >}}
