---
title: "Include"
date: 2022-05-02T20:00:00+02:00
weight: 15
---

The configuration may be split into multiple files by adding `includes = "glob-pattern"` to the main configuration file. 
E.g. the following `profiles.conf` loads configurations from `conf.d` and `profiles.d`:

{{< tabs groupId="config-with-json" >}}
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
{{% tab name="json" %}}

```json
{
  "includes": [
    "conf.d/*.conf",
    "profiles.d/*.yaml",
    "profiles.d/*.toml"
  ],
  "global": {
    "initialize": true
  }
}
```

{{% /tab %}}
{{% /tabs %}}


Included configuration files may use any supported format and settings are merged so that multiple files can extend the same profiles.
The HCL format is special in that it cannot be mixed with other formats.

Included files cannot include nested files. Specifying `includes` inside an included file has no effect.

Within included files, the current [configuration path](../#path-resolution-in-configuration) is not changed. Path resolution remains relative to the path of the main configuration file.

## Configuration Merging

Configuration files are loaded and applied in a fixed order:

1. The main configuration file is loaded first
2. `includes` are iterated in declaration order:
   * Every item may be a single file path or glob expression
   * Glob expressions are resolved and iterated in alphabetical order
   * All paths are resolved relative to [configuration path](../#path-resolution-in-configuration)

Configuration files are loaded in the following order when assuming `/etc/resticprofile/profiles.conf` with `includes = ["first.conf", "conf.d/*.conf", "last.conf"]`:
```
/etc/resticprofile/profiles.conf
/etc/resticprofile/first.conf
/etc/resticprofile/conf.d/00_a.conf
/etc/resticprofile/conf.d/01_a.conf
/etc/resticprofile/conf.d/01_b.conf
/etc/resticprofile/last.conf
```

Configuration **merging** follows the logic:

* Configuration properties are replaced
* Configuration structure (tree) is merged
* What includes later overrides what defines earlier
* Lists of values or objects (tree structure) are considered properties not config structure and will be replaced


{{< tabs groupId="include-merging-example" >}}
{{% tab name="Final configuration" %}}

```yaml
includes:
  - first.yaml
  - second.yaml

default:
  initialize: true
  backup:
     exclude:
        - .*
     source:
        - /etc
        - /opt
```

{{% /tab %}}
{{% tab name="profiles.yaml" %}}

```yaml
includes:
  - first.yaml
  - second.yaml

default:
   
   backup:
      source:
         - /usr


        
```

{{% /tab %}}
{{% tab name="first.yaml" %}}

```yaml
        



default:
   initialize: false
   backup:
      source:
         - /etc
         - /opt

        
```

{{% /tab %}}
{{% tab name="second.yaml" %}}

```yaml
        



default:
   initialize: true
   backup:
      exclude:
         - .*


        
```

{{% /tab %}}
{{% /tabs %}}


{{% notice style="note" %}}

`resticprofile` prior to v0.18.0 had a slightly different behavior when merging configuration properties of a different type (e.g. number <-> text or list <-> single value). In such cases the existing value was not overridden by an included file, breaking the rule "what includes later overrides what defines earlier".

{{% /notice %}}

