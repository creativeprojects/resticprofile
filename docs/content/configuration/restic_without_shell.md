---
title: "Restic without shell"
weight: 27
---

By default, resticprofile runs restic through a shell: `sh` or `bash` on unixes, `cmd.exe` on Windows, or the one set by the `shell` parameter in the `global` section. Configuration values are escaped for that shell, which can still interpret some of them.

Set `restic-without-shell` to `true` in the `global` section to run restic directly: each value reaches restic exactly as it is written in the configuration.

{{% notice style="warning" %}}
This parameter changes how some values are interpreted, so it is disabled by default. Check the differences below before enabling it.
{{% /notice %}}

## Differences

- **Values are passed as written**: no shell removes quotes or backslashes, or expands `*`, `{a,b}` or `~`. For example, an `exclude` value of `My\ Documents` excludes `My\ Documents`, not `My Documents`.
  - Glob patterns in `source` are still resolved by resticprofile, and `~` is still expanded in the values that take a [path]({{% relref "/configuration/path" %}}).
- **Environment variables**: `$VAR` and `${VAR}` are expanded by resticprofile on every operating system. A variable that is not defined is left as written.
  - On Windows, `cmd.exe` is not used anymore, so `%VAR%` and `!VAR!` are not expanded: use `$VAR` instead.
- **Command hooks**: `run-before`, `run-after`, `stdin-command` and the other [command hooks]({{% relref "/configuration/run_hooks" %}}) still run through the shell.
- **`legacy-arguments`** has no effect.

## Example

{{< tabs groupid="config-with-hcl" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[global]
  restic-without-shell = true
```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

global:
  restic-without-shell: true
```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
global {
    restic-without-shell = true
}
```

{{% /tab %}}
{{< /tabs >}}
