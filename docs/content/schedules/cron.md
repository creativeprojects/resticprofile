---
title: "Cron & compatible"
weight: 170
---


On any OS, use a **crond** compatible scheduler if configured in `global` / `scheduler`:

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
[global]
  scheduler = "crond"
```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
---
global:
    scheduler: crond
```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
"global" = {
  "scheduler" = "crond"
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


This configuration uses the default `crontab` tool shipped with `crond`.

You can specify the location of the `crontab` tool:

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
[global]
  scheduler = "crond:/usr/bin/crontab"
```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
---
global:
    scheduler: crond:/usr/bin/crontab
```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
"global" = {
  "scheduler" = "crond:/usr/bin/crontab"
}
```

{{% /tab %}}
{{% tab title="json" %}}

```json
{
  "global": {
    "scheduler": "crond:/usr/bin/crontab"
  }
}
```

{{% /tab %}}
{{< /tabs >}}


## Crontab

You can use a crontab file directly instead of the `crontab` tool:
* `crontab:*:filepath`: Use a crontab file `filepath` **with a user field** filled in automatically
* `crontab:username:filepath`: Use a crontab file `filepath` **with a user field** always set to `username`
* `crontab:-:filepath`: Use a crontab file `filepath` **without a user field**
### With user field

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
[global]
  scheduler = "crontab:*:/etc/cron.d/resticprofile"
```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
---
global:
    scheduler: "crontab:*:/etc/cron.d/resticprofile"
```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
"global" = {
  "scheduler" = "crontab:*:/etc/cron.d/resticprofile"
}
```

{{% /tab %}}
{{% tab title="json" %}}

```json
{
  "global": {
    "scheduler": "crontab:*:/etc/cron.d/resticprofile"
  }
}
```

{{% /tab %}}
{{< /tabs >}}


### Without a user field

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
[global]
  scheduler = "crontab:-:/var/spool/cron/crontabs/username"
```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
---
global:
    scheduler: "crontab:-:/var/spool/cron/crontabs/username"
```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
"global" = {
  "scheduler" = "crontab:-:/var/spool/cron/crontabs/username"
}
```

{{% /tab %}}
{{% tab title="json" %}}

```json
{
  "global": {
    "scheduler": "crontab:-:/var/spool/cron/crontabs/username"
  }
}
```

{{% /tab %}}
{{< /tabs >}}

