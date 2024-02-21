---
title: "Priority"
weight: 35
---

By default, restic is running with the default priority. It means it will get equal share of the resources with other processes.

You can lower the priority of restic to avoid slowing down other processes. This is especially useful when you run restic on a production server.

## Nice

You can use these values for the `priority` parameter:

| String value | "nice" equivalent on unixes |
|-------|-------------------|
| Idle       | 19 |
| Background | 15 |
| Low        | 10 |
| Normal     | 0 |
| High       | -10 |
| Highest    | -20 |

## IO Nice

This setting is only available on Linux. It allows you to set the IO priority of restic.
More information about ionice "class" and "level" can be found [here](https://linux.die.net/man/1/ionice).

## Examples

{{< tabs groupid="config-with-hcl" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[global]
  # priority is using priority class on windows, and "nice" on unixes
  priority = "low"
  # ionice is available on Linux only
  ionice = true
  ionice-class = 2
  ionice-level = 6
```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

global:
  # priority is using priority class on windows, and "nice" on unixes
  priority: low
  # ionice is available on Linux only
  ionice: true
  ionice-class: 2
  ionice-level: 6
```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
global {
    # priority is using priority class on windows, and "nice" on unixes
    priority = "low"
    # ionice is available on Linux only
    ionice = true
    ionice-class = 2
    ionice-level = 6
}
```

{{% /tab %}}
{{< /tabs >}}

## Warnings

In some cases, resticprofile will not be able to set the priority of restic.

A warning message like this will be displayed:

```
cannot set process group priority, restic will run with the default priority: operation not permitted
```

This usually means:
- resticprofile is running inside docker
- you are using a tight security linux distribution which is launching every process inside a new container
- resticprofile is running in WSL
