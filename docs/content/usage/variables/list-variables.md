---
title: "More on Variables"
---

{{< toc >}}

## Hand-made variables
Here's an example of a configuration on Linux where I use a variable `$mountpoint` set to a USB drive mount point:

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[global]
  priority = "low"

{{ $mountpoint := "/mnt/external" }}

[default]
  repository = "local:{{ $mountpoint }}/backup"
  password-file = "key"
  run-before = "mount {{ $mountpoint }}"
  run-after = "umount {{ $mountpoint }}"
  run-after-fail = "umount {{ $mountpoint }}"

  [default.backup]
    exclude-caches = true
    source = [ "/etc", "/var/lib/libvirt" ]
    check-after = true
```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

global:
  priority: low

{{ $mountpoint := "/mnt/external" }}

default:
  repository: 'local:{{ $mountpoint }}/backup'
  password-file: key
  run-before: 'mount {{ $mountpoint }}'
  run-after: 'umount {{ $mountpoint }}'
  run-after-fail: 'umount {{ $mountpoint }}'

  backup:
    exclude-caches: true
    source:
      - /etc
      - /var/lib/libvirt
    check-after: true
```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
global {
    priority = "low"
}

{{ $mountpoint := "/mnt/external" }}

default {
    repository = "local:{{ $mountpoint }}/backup"
    password-file = "key"
    run-before = "mount {{ $mountpoint }}"
    run-after = "umount {{ $mountpoint }}"
    run-after-fail = "umount {{ $mountpoint }}"

    backup {
        exclude-caches = true
        source = [ "/etc", "/var/lib/libvirt" ]
        check-after = true
    }
}

```

{{% /tab %}}
{{% tab title="json" %}}

```json
{{ $mountpoint := "/mnt/external" }}
{
  "version": "1",
  "global": {
    "priority": "low"
  },
  "default": {
    "repository": "local:{{ $mountpoint }}/backup",
    "password-file": "key",
    "run-before": "mount {{ $mountpoint }}",
    "run-after": "umount {{ $mountpoint }}",
    "run-after-fail": "umount {{ $mountpoint }}",
    "backup": {
      "exclude-caches": true,
      "source": [
        "/etc",
        "/var/lib/libvirt"
      ],
      "check-after": true
    }
  }
}
```

{{% /tab %}}
{{< /tabs >}}

## Runtime variable expansion

Backup current dir (`$PWD`) but prevent backup of `$HOME` where the repository is located:

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[default]
  repository = "local:${HOME}/backup"
  password-file = "${HOME}/backup.key"

  [default.backup]
    source = "$PWD"
    exclude = ["$HOME/**", ".*", "~*"]

```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

default:
  repository: 'local:${HOME}/backup'
  password-file: '${HOME}/backup.key'

  backup:
    source: '$PWD'
    exclude: ['$HOME/**', '.*', '~*']

```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
default {
    repository = "local:${HOME}/backup"
    password-file = "${HOME}/backup.key"

    backup {
        source = [ "$PWD" ]
        exclude = [ "$HOME/**", ".*", "~*" ]
    }
}
```

{{% /tab %}}
{{% tab title="json" %}}

```json
{
  "default": {
    "repository": "local:${HOME}/backup",
    "password-file": "${HOME}/backup.key",
    "backup": {
      "source": [ "$PWD" ],
      "exclude": [ "$HOME/**", ".*", "~*" ]
    }
  }
}
```

{{% /tab %}}
{{< /tabs >}}

{{% notice style="tip" %}}
Use `$$` to escape a single `$` in configuration values that support variable expansion. E.g. on Windows you might want to exclude `$RECYCLE.BIN`. Specify it as: `exclude = ["$$RECYCLE.BIN"]`.
{{% /notice %}}
