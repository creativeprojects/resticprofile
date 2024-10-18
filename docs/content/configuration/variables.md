---
title: "Variables"
date: 2022-05-16T20:04:35+01:00
weight: 26
---

## Variable expansion in configuration file

You might want to reuse the same configuration (or bits of it) on different environments. One way of doing it is to create a generic configuration where specific bits can be replaced by a variable.

### There are two kinds of variables:
- **template variables**: These variables are fixed once the full configuration file is loaded: [includes]({{% relref "configuration/include" %}}) are loaded, and [inheritance]({{% relref "/configuration/inheritance" %}}) is resolved. These variables are replaced by their value **before** the configuration is parsed.
- **runtime variables**: These variables are replaced by their value **after** the configuration is parsed. In other words: these variables are replaced by their value just before the command is executed.

## Template variables
### Pre-defined variables

The syntax for using a pre-defined variable is:

```
{{ .VariableName }}
```

The list of pre-defined variables is:

| Variable          | Type                                             | Description                                                      |
|-------------------|--------------------------------------------------|------------------------------------------------------------------|
| **.Profile.Name** | string                                           | Profile name                                                     |
| **.Now**          | [time.Time](https://golang.org/pkg/time/) object | Now object: see explanation bellow                               |
| **.StartupDir**   | string                                           | Current directory at the time resticprofile was started          |
| **.CurrentDir**   | string                                           | Current directory at the time a profile is executed              |
| **.ConfigDir**    | string                                           | Directory where the configuration was loaded from                |
| **.TempDir**      | string                                           | OS temporary directory (might not exist)                         |
| **.BinaryDir**    | string                                           | Directory where resticprofile was started from (since `v0.18.0`) |
| **.OS**           | string                                           | GOOS name: "windows", "linux", "darwin", etc. (since `v0.21.0`)  |
| **.Arch**         | string                                           | GOARCH name: "386", "amd64", "arm64", etc. (since `v0.21.0`)     |
| **.Hostname**     | string                                           | Host name                                                        |
| **.Env.{NAME}**   | string                                           | Environment variable `${NAME}`                                   |

Environment variables are accessible using `.Env.` followed by the (upper case) name of the environment variable.

Example: `{{ .Env.HOME }}` will be replaced by your home directory (on unixes). The equivalent on Windows would be `{{ .Env.USERPROFILE }}`.

Default and fallback values for an empty or unset variable can be declared with `{{ ... | or ... }}`.
For example `{{ .Env.HOME | or .Env.USERPROFILE | or "/fallback-homedir" }}` will try to resolve `$HOME`, if empty try to resolve `$USERPROFILE` 
and finally default to `/fallback-homedir` if none of the env variables are defined.

The variables `.OS` and `.Arch` are filled with the target platform that `resticprofile` was compiled for (see 
[releases](https://github.com/creativeprojects/resticprofile/releases) for more information on existing precompiled platform binaries). 

For variables that are objects, you can call all public fields or methods on it.
For example, for the variable `.Now` ([time.Time](https://golang.org/pkg/time/)) you can use:

- `(.Now.AddDate years months days)`
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

Time can be formatted with `.Now.Format layout`, for example `{{ .Now.Format "2006-01-02T15:04:05Z07:00" }}` formats the current time as RFC3339 timestamp. 
Check [time.Time#constants](https://pkg.go.dev/time#pkg-constants) for more layout examples.

The variable `.Now` also allows to derive a relative `Time`. For example `{{ (.Now.AddDate 0 -6 -14).Format "2006-01-02" }}` formats a date that 
is 6 months and 14 days before now.


#### Example

You can use a combination of inheritance and variables in the resticprofile configuration file like so:

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
version = "1"

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
  
  [src.check]
    # Weekday is an integer from 0 to 6 (starting from Sunday)
    # Nice trick to add 1 to an integer: https://stackoverflow.com/a/72465098
    read-data-subset = "{{ len (printf "a%*s" .Now.Weekday "") }}/7"

```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
---
version: "1"

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

    check:
        # Weekday is an integer from 0 to 6 (starting from Sunday)
        # Nice trick to add 1 to an integer: https://stackoverflow.com/a/72465098
        read-data-subset: "{{ len (printf "a%*s" .Now.Weekday "") }}/7"

```

{{% /tab %}}
{{% tab title="hcl" %}}

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

  "check" = {
    # Weekday is an integer from 0 to 6 (starting from Sunday)
    # Nice trick to add 1 to an integer: https://stackoverflow.com/a/72465098
    "read-data-subset" = "{{ len (printf "a%*s" .Now.Weekday "") }}/7"
  }
}
```

{{% /tab %}}
{{% tab title="json" %}}

```json
{
  "version": "1",
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
    },
    "check": {
      "read-data-subset": "{{ len (printf "a%*s" .Now.Weekday "") }}/7"
    }
  }
}
```

{{% /tab %}}
{{< /tabs >}}

This is obviously not a real world example, but it shows many of the possibilities you can do with variable expansion.

To check the generated configuration, you can use the resticprofile `show` command:

```shell
% resticprofile -c examples/template.yaml -n src show

global:
    default-command:          snapshots
    restic-lock-retry-after:  1m0s
    restic-stale-lock-age:    2h0m0s
    min-memory:               100
    send-timeout:             30s

profile src:
    repository:     /backup/Monday
    password-file:  /Users/CP/go/src/resticprofile/examples/src-key
    initialize:     true
    lock:           /Users/CP/resticprofile-profile-src.lock

    backup:
        check-before:    true
        run-before:      echo Hello CP
                         echo current dir: /Users/CP/go/src/resticprofile
                         echo config dir: /Users/CP/go/src/resticprofile/examples
                         echo profile started at 05 Sep 22 17:39 BST
        run-after:       echo All Done!
        source:          /Users/CP/go/src
        exclude:         /**/.git
        exclude-caches:  true
        tag:             src
                         dev

    retention:
        after-backup:  true
        keep-within:   30d
        path:          /Users/CP/go/src
        prune:         true
        tag:           src
                       dev

    check:
        read-data-subset:  2/7

    snapshots:
        tag:  src
              dev
```

As you can see, the `src` profile inherited from the `generic` profile. The tags `{{ .Profile.Name }}` got replaced by the name of the current profile `src`.
Now you can reuse the same generic configuration in another profile.

You might have noticed the `read-data-subset` in the `check` section which will read a seventh of the data every day, meaning the whole repository data will be checked over a week. You can find [more information about this trick](https://stackoverflow.com/a/72465098).

### Hand-made variables

You can also define variables yourself. Hand-made variables starts with a `$` ([PHP](https://en.wikipedia.org/wiki/PHP) anyone?) and get declared and assigned with the `:=` operator ([Pascal](https://en.wikipedia.org/wiki/Pascal_(programming_language)) anyone?).

{{% notice style="info" %}}
You can only use double quotes `"` to declare the string, single quotes `'` are not allowed. You can also use backticks to declare the string.
{{% /notice %}}

Here's an example:

```yaml
# declare and assign a value to the variable
{{ $name := "something" }}

profile:
  # put the content of the variable here
  tag: "{{ $name }}"
```
{{% notice style="note" %}}
Variables are only valid in the file they are declared in. They cannot be shared in files loaded via `include`.
{{% /notice %}}

Variables can be redefined using the `=` operator. The new value will be used from the point of redefinition to the end of the file.

```yaml
# declare and assign a value to the variable
{{ $name := "something" }}

# reassign a new value to the variable
{{ $name = "something else" }}

```

#### Windows path inside a variable

Windows path are using backslashes `\` and are interpreted as escape characters in the configuration file. To use a Windows path inside a variable, you have a few options:
- you can escape the backslashes with another backslash.
- you can use forward slashes `/` instead of backslashes. Windows is able to use forward slashes in paths.
- you can use the backtick to declare the string instead of a double quote.

For example:
```yaml
# double backslash
{{ $path := "C:\\Users\\CP\\Documents" }}
# forward slash
{{ $path := "C:/Users/CP/Documents" }}
# backticks
{{ $path := `C:\Users\CP\Documents` }}
```

#### Example

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

Variable expansion as described in the previous section using the `{{ .Var }}` syntax refers to [template variables]({{% relref "/configuration/templates" %}}) that are expanded prior to parsing the configuration file. 
This means they must be used carefully to create correct config markup, but they are also very flexible.

There is also unix style variable expansion using the `${variable}` or `$variable` syntax on configuration **values** that expand after the config file was parsed. Values that take a file path or path expression and a few others support this expansion. 

If not specified differently, these variables resolve to the corresponding environment variable or to an empty value if no such environment variable exists. Exceptions are [mixins]({{% relref "/configuration/inheritance#mixins" %}}) where `$variable` style is used for parametrisation and the profile [config flag]({{% relref "reference/profile" %}}) `prometheus-push-job`.

### Example

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
