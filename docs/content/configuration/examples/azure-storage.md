---
title: "Simple Configuration with Azure Storage"
---

Here's a simple configuration file using a Microsoft Azure backend. You will notice that the `env` section lets you define environment variables:

{{< tabs groupid="config-with-hcl" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[default]
  repository = "azure:restic:/"
  password-file = "key"
  option = "azure.connections=3"

  [default.env]
    AZURE_ACCOUNT_NAME = "my_storage_account"
    AZURE_ACCOUNT_KEY = "my_super_secret_key"

  [default.backup]
    exclude-file = "excludes"
    exclude-caches = true
    one-file-system = true
    tag = [ "root" ]
    source = [ "/", "/var" ]
    schedule = "daily"
    schedule-after-network-online = true
```
{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

default:
  repository: "azure:restic:/"
  password-file: "key"
  option: "azure.connections=3"

  env:
    AZURE_ACCOUNT_NAME: "my_storage_account"
    AZURE_ACCOUNT_KEY: "my_super_secret_key"

  backup:
    exclude-file: "excludes"
    exclude-caches: true
    one-file-system: true
    tag:
      - "root"
    source:
      - "/"
      - "/var"
    schedule: "daily"
    schedule-after-network-online: true
```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
default {
    repository = "azure:restic:/"
    password-file = "key"
    options = "azure.connections=3"

    env {
      AZURE_ACCOUNT_NAME = "my_storage_account"
      AZURE_ACCOUNT_KEY = "my_super_secret_key"
    }

    backup = {
        exclude-file = "excludes"
        exclude-caches = true
        one-file-system = true
        tag = [ "root" ]
        source = [ "/", "/var" ]
    }
}
```

{{% /tab %}}
{{< /tabs >}}
