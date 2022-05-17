---
title: "Include"
date: 2022-04-25T21:14:16+01:00
weight: 15
---

The configuration may be split into multiple files by adding `includes = "glob-pattern"` to the main configuration file. 
E.g. the following `profiles.conf` loads configurations from `conf.d` and `profiles.d`:

{{< tabs groupId="config-with-hcl" >}}
{{% tab name="toml" %}}

```toml
# Includes
includes = ["conf.d/*.conf", "profiles.d/*.yaml", "profiles.d/*.toml"]

# Defaults
[global]
initialize = true
```


{{% /tab %}}
{{% tab name="yaml" %}}

```yaml
        
includes:
  - "conf.d/*.conf"
  - "profiles.d/*.yaml"
  - "profiles.d/*.toml"

global:
  initialize: true

```

{{% /tab %}}
{{% tab name="hcl" %}}

```hcl

includes = ["conf.d/*.conf", "profiles.d/*.yaml", "profiles.d/*.toml"]

global {
  initialize = true
}
```

{{% /tab %}}
{{% /tabs %}}


Included configuration files may use any supported format and settings are merged so that multiple files can extend the same profiles.
The HCL format is special in that it cannot be mixed with other formats.

Included files cannot include nested files and do not change the current configuration path.
