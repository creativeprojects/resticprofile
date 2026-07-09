---
title: "Message Body Template"
weight: 30
---

You can use a standard go template to build the webhook body. It has to be defined in a separate file (otherwise it would clash with the configuration as a go template itself).

More information can be found under [Hooks - HTTP Hooks]({{% relref "configuration/hooks/http_hooks/#body-template" %}})

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
