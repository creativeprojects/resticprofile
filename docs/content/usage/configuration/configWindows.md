---
title: "Example: Windows"
weight: 30
---

{{< tabs groupid="config-with-hcl" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[global]
  restic-binary = "c:\\ProgramData\\chocolatey\\bin\\restic.exe"

# Default profile when not specified on the command line
# There's no default inheritance from the 'default' profile,
# but you can use the 'inherit' flag if needed
[default]
  repository = "local:r:/"
  password-file = "key"
  initialize = false

# New profile named 'test'
[test]
  inherit = "default"
  initialize = true

  # 'backup' command of profile 'test'
  [test.backup]
    tag = [ "windows" ]
    source = [ "c:\\" ]
    check-after = true
    run-before = "dir /l"
    run-after = "echo All Done!"
    # ignore restic warnings
    # without it the backup is considered failed when restic can't read some files
    no-error-on-warning = true
```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

global:
  restic-binary: c:\ProgramData\chocolatey\bin\restic.exe

# Default profile when not specified on the command line
# There's no default inheritance from the 'default' profile,
# but you can use the 'inherit' flag if needed
default:
  repository: local:r:/
  password-file: key
  initialize: false

# New profile named 'test'
test:
  inherit: default
  initialize: true

  # 'backup' command of profile 'test'
  backup:
    tag:
      - windows
    source:
      - c:\
    check-after: true
    run-before: dir /l
    run-after: echo All Done!
    # ignore restic warnings
    # without it the backup is considered failed when restic can't read some files
    no-error-on-warning: true
```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
global {
  restic-binary = "c:\\ProgramData\\chocolatey\\bin\\restic.exe"
}

default {
  repository = "local:r:/"
  password-file = "key"
  initialize = false
}

test {
  inherit = "default"
  initialize = true

  backup = {
    tag = [ "windows" ]
    source = [ "c:\\" ]
    check-after = true
    run-before = "dir /l"
    run-after = "echo All Done!"
    no-error-on-warning = true
  }
}

```

{{% /tab %}}
{{< /tabs >}}
