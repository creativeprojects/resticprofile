---
title: "Variables"
date: 2022-05-16T20:04:35+01:00
weight: 25
---


## Variable expansion in configuration file

You might want to reuse the same configuration (or bits of it) on different environments. One way of doing it is to create a generic configuration where specific bits will be replaced by a variable.

## Pre-defined variables

The syntax for using a pre-defined variable is:
```
{{ .VariableName }}
```


The list of pre-defined variables is:
- **.Profile.Name** (string)
- **.Now** ([time.Time](https://golang.org/pkg/time/) object)
- **.CurrentDir** (string)
- **.ConfigDir** (string)
- **.Hostname** (string)
- **.Env.{NAME}** (string)

Environment variables are accessible using `.Env.` followed by the name of the environment variable.

Example: `{{ .Env.HOME }}` will be replaced by your home directory (on unixes). The equivalent on Windows would be `{{ .Env.USERPROFILE }}`.

For variables that are objects, you can call all public field or method on it.
For example, for the variable `.Now` you can use:
- `.Now.Day`
- `.Now.Format layout`
- `.Now.Hour`
- `.Now.Minute`
- `.Now.Month`
- `.Now.Second`
- `.Now.UTC`
- `.Now.Unix`
- `.Now.Weekday`
- `.Now.Year`
- `.Now.YearDay`


## Hand-made variables

But you can also define variables yourself. Hand-made variables starts with a `$` ([PHP](https://en.wikipedia.org/wiki/PHP) anyone?) and get declared and assigned with the `:=` operator ([Pascal](https://en.wikipedia.org/wiki/Pascal_(programming_language)) anyone?). Here's an example:

```yaml
# declare and assign a value to the variable
{{ $name := "something" }}

# put the content of the variable here
tag: "{{ $name }}"
```

## Examples

You can use a combination of inheritance and variables in the resticprofile configuration file like so:

{{< tabs groupId="config-with-json" >}}
{{% tab name="toml" %}}

```toml
[generic]
  password-file = "{{ .ConfigDir }}/{{ .Profile.Name }}-key"
  repository = "/backup/{{ .Now.Weekday }}"
  lock = "$HOME/resticprofile-profile-{{ .Profile.Name }}.lock"
  initialize = true

  [generic.backup]
    check-before = true
    exclude = [ "/**/.git" ]
    exclude-caches = true
    one-file-system = false
    run-after = "echo All Done!"
    run-before = [
        "echo Hello {{ .Env.LOGNAME }}",
        "echo current dir: {{ .CurrentDir }}",
        "echo config dir: {{ .ConfigDir }}",
        "echo profile started at {{ .Now.Format "02 Jan 06 15:04 MST" }}"
    ]
    tag = [ "{{ .Profile.Name }}", "dev" ]

  [generic.retention]
    after-backup = true
    before-backup = false
    compact = false
    keep-within = "30d"
    prune = true
    tag = [ "{{ .Profile.Name }}", "dev" ]

  [generic.snapshots]
    tag = [ "{{ .Profile.Name }}", "dev" ]

[src]
  inherit = "generic"

  [src.backup]
    source = [ "{{ .Env.HOME }}/go/src" ]

```

{{% /tab %}}
{{% tab name="yaml" %}}

```yaml
---
generic:
    password-file: "{{ .ConfigDir }}/{{ .Profile.Name }}-key"
    repository: "/backup/{{ .Now.Weekday }}"
    lock: "$HOME/resticprofile-profile-{{ .Profile.Name }}.lock"
    initialize: true

    backup:
        check-before: true
        exclude:
        - /**/.git
        exclude-caches: true
        one-file-system: false
        run-after: echo All Done!
        run-before:
          - "echo Hello {{ .Env.LOGNAME }}"
          - "echo current dir: {{ .CurrentDir }}"
          - "echo config dir: {{ .ConfigDir }}"
          - "echo profile started at {{ .Now.Format "02 Jan 06 15:04 MST" }}"
        tag:
          - "{{ .Profile.Name }}"
          - dev

    retention:
        after-backup: true
        before-backup: false
        compact: false
        keep-within: 30d
        prune: true
        tag:
          - "{{ .Profile.Name }}"
          - dev

    snapshots:
        tag:
          - "{{ .Profile.Name }}"
          - dev

src:
    inherit: generic
    backup:
        source:
          - "{{ .Env.HOME }}/go/src"

```

{{% /tab %}}
{{% tab name="hcl" %}}

```hcl
"generic" = {
  "password-file" = "{{ .ConfigDir }}/{{ .Profile.Name }}-key"
  "repository" = "/backup/{{ .Now.Weekday }}"
  "lock" = "$HOME/resticprofile-profile-{{ .Profile.Name }}.lock"
  "initialize" = true

  "backup" = {
    "check-before" = true
    "exclude" = ["/**/.git"]
    "exclude-caches" = true
    "one-file-system" = false
    "run-after" = "echo All Done!"
    "run-before" = ["echo Hello {{ .Env.LOGNAME }}", "echo current dir: {{ .CurrentDir }}", "echo config dir: {{ .ConfigDir }}", "echo profile started at {{ .Now.Format "02 Jan 06 15:04 MST" }}"]
    "tag" = ["{{ .Profile.Name }}", "dev"]
  }

  "retention" = {
    "after-backup" = true
    "before-backup" = false
    "compact" = false
    "keep-within" = "30d"
    "prune" = true
    "tag" = ["{{ .Profile.Name }}", "dev"]
  }

  "snapshots" = {
    "tag" = ["{{ .Profile.Name }}", "dev"]
  }
}

"src" = {
  "inherit" = "generic"

  "backup" = {
    "source" = ["{{ .Env.HOME }}/go/src"]
  }
}
```

{{% /tab %}}
{{% tab name="json" %}}

```json
{
  "generic": {
    "password-file": "{{ .ConfigDir }}/{{ .Profile.Name }}-key",
    "repository": "/backup/{{ .Now.Weekday }}",
    "lock": "$HOME/resticprofile-profile-{{ .Profile.Name }}.lock",
    "initialize": true,
    "backup": {
      "check-before": true,
      "exclude": [
        "/**/.git"
      ],
      "exclude-caches": true,
      "one-file-system": false,
      "run-after": "echo All Done!",
      "run-before": [
        "echo Hello {{ .Env.LOGNAME }}",
        "echo current dir: {{ .CurrentDir }}",
        "echo config dir: {{ .ConfigDir }}",
        "echo profile started at {{ .Now.Format "02 Jan 06 15:04 MST" }}"
      ],
      "tag": [
        "{{ .Profile.Name }}",
        "dev"
      ]
    },
    "retention": {
      "after-backup": true,
      "before-backup": false,
      "compact": false,
      "keep-within": "30d",
      "prune": true,
      "tag": [
        "{{ .Profile.Name }}",
        "dev"
      ]
    },
    "snapshots": {
      "tag": [
        "{{ .Profile.Name }}",
        "dev"
      ]
    }
  },
  "src": {
    "inherit": "generic",
    "backup": {
      "source": [
        "{{ .Env.HOME }}/go/src"
      ]
    }
  }
}
```

{{% /tab %}}
{{% /tabs %}}

This is obviously not a real world example, but it shows many of the possibilities you can do with variable expansion.

To check the generated configuration, you can use the resticprofile `show` command:

```
% resticprofile -c examples/template.yaml -n src show

global:
    default-command:          snapshots
    restic-lock-retry-after:  1m0s
    restic-stale-lock-age:    2h0m0s
    min-memory:               100

profile src:
    repository:     /backup/Wednesday
    password-file:  /Users/CP/go/src/resticprofile/examples/src-key
    initialize:     true
    lock:           /Users/CP/resticprofile-profile-src.lock

    backup:
        check-before:    true
        run-before:      echo Hello CP
                         echo current dir: /Users/CP/go/src/resticprofile
                         echo config dir: /Users/CP/go/src/resticprofile/examples
                         echo profile started at 18 May 22 18:32 BST
        run-after:       echo All Done!
        source:          /Users/CP/go/src
        exclude:         /**/.git
        tag:             src
                         dev
        exclude-caches:  true

    retention:
        after-backup:  true
        prune:         true
        tag:           src
                       dev
        keep-within:   30d
        path:          /Users/CP/go/src

    snapshots:
        tag:  src
              dev
```

As you can see, the `src` profile inherited from the `generic` profile. The tags `{{ .Profile.Name }}` got replaced by the name of the current profile `src`. Now you can reuse the same generic configuration in another profile.

Here's another example of a configuration on Linux where I use a variable `$mountpoint` set to a USB drive mount point:

{{< tabs groupId="config-with-json" >}}
{{% tab name="toml" %}}

```toml
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
{{% tab name="yaml" %}}

```yaml
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
{{% tab name="hcl" %}}


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
{{% tab name="json" %}}

```json
{{ $mountpoint := "/mnt/external" }}
{
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
{{% /tabs %}}
