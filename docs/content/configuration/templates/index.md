---
title: "Templates"
date: 2022-05-16T20:04:35+01:00
weight: 26
---



Templates are a great way to compose configuration profiles.

Please keep in mind that `yaml` files are sensitive to the number of spaces. Also if you declare a block already declared, it overrides the previous declaration (instead of merging them).

For that matter, configuration templates are probably more useful if you use the `toml` or `hcl` configuration format.

Here's a simple example

```
{{ define "hello" }}
hello = "world"
{{ end }}
```

To use the content of this template anywhere in your configuration, simply call it:

```
{{ template "hello" . }}
```

Note the **dot** after the name: it's used to pass the variables to the template. Without it, all your variables (like `.Profile.Name`) would display `<no value>`.

Here's a working example:

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
#
# This is an example of configuration using nested templates
#

# nested template declarations
# this template declaration won't appear here in the configuration file
# it will only appear when called by { { template "backup_root" . } }
{{ define "backup_root" }}
    exclude = [ "{{ .Profile.Name }}-backup.log" ]
    exclude-file = [
        "{{ .ConfigDir }}/root-excludes",
        "{{ .ConfigDir }}/excludes"
    ]
    exclude-caches = true
    tag = [ "root" ]
    source = [ "/" ]
{{ end }}

version = "1"

[global]
  priority = "low"
  ionice = true
  ionice-class = 2
  ionice-level = 6

[base]
  status-file = "{{ .Env.HOME }}/status.json"

    [base.snapshots]
      host = true

    [base.retention]
      host = true
      after-backup = true
      keep-within = "30d"

#########################################################

[nas]
  inherit = "base"
  repository = "rest:http://{{ .Env.BACKUP_REST_USER }}:{{ .Env.BACKUP_REST_PASSWORD }}@nas:8000/root"
  password-file = "nas-key"

# root

[nas-root]
  inherit = "nas"

    [nas-root.backup]
      # get the content of "backup_root" defined at the top
      {{ template "backup_root" . }}
      schedule = "01:47"
      schedule-permission = "system"
      schedule-log = "{{ .Profile.Name }}-backup.log"

#########################################################

[azure]
  inherit = "base"
  repository = "azure:restic:/"
  password-file = "azure-key"
  lock = "/tmp/resticprofile-azure.lock"

    [azure.backup]
      schedule-permission = "system"
      schedule-log = "{{ .Profile.Name }}-backup.log"

# root

[azure-root]
  inherit = "azure"

    [azure-root.backup]
      # get the content of "backup_root" defined at the top
      {{ template "backup_root" . }}
      schedule = "03:58"

# mysql

[azure-mysql]
  inherit = "azure"

    [azure-mysql.backup]
      tag = [ "mysql" ]
      run-before = [
          "rm -f /tmp/mysqldumpall.sql",
          "mysqldump -u{{ .Env.MYSQL_BACKUP_USER }} -p{{ .Env.MYSQL_BACKUP_PASSWORD }} --all-databases > /tmp/mysqldumpall.sql"
      ]
      source = "/tmp/mysqldumpall.sql"
      run-after = [
          "rm -f /tmp/mysqldumpall.sql"
      ]
      schedule = "03:18"

```

{{% /tab %}}
{{% tab title="yaml" %}}


```yaml
#
# This is an example of configuration using nested templates
#

# nested template declarations
# this template declaration won't appear here in the configuration file
# it will only appear when called by { { template "backup_root" . } }

{{ define "backup_root" }}
    exclude:
      - '{{ .Profile.Name }}-backup.log'
    exclude-file:
      - '{{ .ConfigDir }}/root-excludes'
      - '{{ .ConfigDir }}/excludes'
    exclude-caches: true
    tag:
      - root
    source:
      - /
{{ end }}

version: "1"

global:
  priority: low
  ionice: true
  ionice-class: 2
  ionice-level: 6

base:
  status-file: '{{ .Env.HOME }}/status.json'
  snapshots:
    host: true
  retention:
    host: true
    after-backup: true
    keep-within: 30d

nas:
  inherit: base
  repository: >-
    rest:http://{{ .Env.BACKUP_REST_USER }}:{{ .Env.BACKUP_REST_PASSWORD
    }}@nas:8000/root
  password-file: nas-key

nas-root:
  inherit: nas
  backup:
    # get the content of "backup_root" defined at the top
{{ template "backup_root" . }}
    schedule: '01:47'
    schedule-permission: system
    schedule-log: '{{ .Profile.Name }}-backup.log'

azure:
  inherit: base
  repository: 'azure:restic:/'
  password-file: azure-key
  lock: /tmp/resticprofile-azure.lock
  backup:
    schedule-permission: system
    schedule-log: '{{ .Profile.Name }}-backup.log'

azure-root:
  inherit: azure
  backup:
    # get the content of "backup_root" defined at the top
{{ template "backup_root" . }}
    schedule: '03:58'

azure-mysql:
  inherit: azure
  backup:
    tag:
      - mysql
    run-before:
      - rm -f /tmp/mysqldumpall.sql
      - >-
        mysqldump -u{{ .Env.MYSQL_BACKUP_USER }} -p{{ .Env.MYSQL_BACKUP_PASSWORD
        }} --all-databases > /tmp/mysqldumpall.sql
    source: /tmp/mysqldumpall.sql
    run-after:
      - rm -f /tmp/mysqldumpall.sql
    schedule: '03:18'
```


{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
#
# This is an example of configuration using nested templates
#

# nested template declarations
# this template declaration won't appear here in the configuration file
# it will only appear when called by { { template "backup_root" . } }

{{ define "backup_root" }}
  "exclude" = ["{{ .Profile.Name }}-backup.log"]
  "exclude-file" = ["{{ .ConfigDir }}/root-excludes", "{{ .ConfigDir }}/excludes"]
  "exclude-caches" = true
  "tag" = ["root"]
  "source" = ["/"]
{{end}}

"global" = {
  "priority" = "low"
  "ionice" = true
  "ionice-class" = 2
  "ionice-level" = 6
}

"base" = {
  "status-file" = "{{ .Env.HOME }}/status.json"

  "snapshots" = {
    "host" = true
  }

  "retention" = {
    "host" = true
    "after-backup" = true
    "keep-within" = "30d"
  }
}

"nas" = {
  "inherit" = "base"
  "repository" = "rest:http://{{ .Env.BACKUP_REST_USER }}:{{ .Env.BACKUP_REST_PASSWORD }}@nas:8000/root"
  "password-file" = "nas-key"
}

"nas-root" = {
  "inherit" = "nas"

  "backup" = {
    # get the content of "backup_root" defined at the top
    {{ template "backup_root" . }}
    "schedule" = "01:47"
    "schedule-permission" = "system"
    "schedule-log" = "{{ .Profile.Name }}-backup.log"
  }
}

"azure" = {
  "inherit" = "base"
  "repository" = "azure:restic:/"
  "password-file" = "azure-key"
  "lock" = "/tmp/resticprofile-azure.lock"

  "backup" = {
    "schedule-permission" = "system"
    "schedule-log" = "{{ .Profile.Name }}-backup.log"
  }
}

"azure-root" = {
  "inherit" = "azure"

  "backup" = {
    # get the content of "backup_root" defined at the top
    {{ template "backup_root" . }}
    "schedule" = "03:58"
  }
}

"azure-mysql" = {
  "inherit" = "azure"

  "backup" = {
    "tag" = ["mysql"]
    "run-before" = ["rm -f /tmp/mysqldumpall.sql", "mysqldump -u{{ .Env.MYSQL_BACKUP_USER }} -p{{ .Env.MYSQL_BACKUP_PASSWORD }} --all-databases > /tmp/mysqldumpall.sql"]
    "source" = "/tmp/mysqldumpall.sql"
    "run-after" = ["rm -f /tmp/mysqldumpall.sql"]
    "schedule" = "03:18"
  }
}
```

{{% /tab %}}
{{% tab title="json" %}}

```json
{{ define "backup_root" }}
    "exclude": [
      "{{ .Profile.Name }}-backup.log"
    ],
    "exclude-file": [
      "{{ .ConfigDir }}/root-excludes",
      "{{ .ConfigDir }}/excludes"
    ],
    "exclude-caches": true,
    "tag": [
      "root"
    ],
    "source": [
      "/"
    ],
{{ end }}
{
  "version": "1",
  "global": {
    "priority": "low",
    "ionice": true,
    "ionice-class": 2,
    "ionice-level": 6
  },
  "base": {
    "status-file": "{{ .Env.HOME }}/status.json",
    "snapshots": {
      "host": true
    },
    "retention": {
      "host": true,
      "after-backup": true,
      "keep-within": "30d"
    }
  },
  "nas": {
    "inherit": "base",
    "repository": "rest:http://{{ .Env.BACKUP_REST_USER }}:{{ .Env.BACKUP_REST_PASSWORD }}@nas:8000/root",
    "password-file": "nas-key"
  },
  "nas-root": {
    "inherit": "nas",
    "backup": {
      {{ template "backup_root" . }}
      "schedule": "01:47",
      "schedule-permission": "system",
      "schedule-log": "{{ .Profile.Name }}-backup.log"
    }
  },
  "azure": {
    "inherit": "base",
    "repository": "azure:restic:/",
    "password-file": "azure-key",
    "lock": "/tmp/resticprofile-azure.lock",
    "backup": {
      "schedule-permission": "system",
      "schedule-log": "{{ .Profile.Name }}-backup.log"
    }
  },
  "azure-root": {
    "inherit": "azure",
    "backup": {
      {{ template "backup_root" . }}
      "schedule": "03:58"
    }
  },
  "azure-mysql": {
    "inherit": "azure",
    "backup": {
      "tag": [
        "mysql"
      ],
      "run-before": [
        "rm -f /tmp/mysqldumpall.sql",
        "mysqldump -u{{ .Env.MYSQL_BACKUP_USER }} -p{{ .Env.MYSQL_BACKUP_PASSWORD }} --all-databases > /tmp/mysqldumpall.sql"
      ],
      "source": "/tmp/mysqldumpall.sql",
      "run-after": [
        "rm -f /tmp/mysqldumpall.sql"
      ],
      "schedule": "03:18"
    }
  }
}
```

{{% /tab %}}
{{% /tabs %}}


## Debugging your template and variable expansion

If for some reason you don't understand why resticprofile is not loading your configuration file, you can display the generated configuration after executing the template (and replacing the variables and everything) using the `--trace` flag. We will see it in action in a moment.

## Limitations of using templates

There's something to be aware of when dealing with templates: at the time the template is compiled, it has no knowledge of what the end configuration should look like: it has no knowledge of profiles for example. Here is a **non-working** example of what I mean:

<!-- checkdoc-ignore -->
```toml
version = "1"

{{ define "retention" }}
  [{{ .Profile.Name }}.retention]
    after-backup = true
    before-backup = false
    compact = false
    keep-within = "30d"
    prune = true
{{ end }}

[src]
  password-file = "{{ .ConfigDir }}/{{ .Profile.Name }}-key"
  repository = "/backup/{{ .Now.Weekday }}"
  lock = "$HOME/resticprofile-profile-{{ .Profile.Name }}.lock"
  initialize = true

    [src.backup]
      source = "{{ .Env.HOME }}/go/src"
      check-before = true
      exclude = ["/**/.git"]
      exclude-caches = true
      tag = ["{{ .Profile.Name }}", "dev"]

    {{ template "retention" . }}

    [src.snapshots]
      tag = ["{{ .Profile.Name }}", "dev"]

[other]
  password-file = "{{ .ConfigDir }}/{{ .Profile.Name }}-key"
  repository = "/backup/{{ .Now.Weekday }}"
  lock = "$HOME/resticprofile-profile-{{ .Profile.Name }}.lock"
  initialize = true

    {{ template "retention" . }}

```

Here we define a template `retention` that we use twice.
When you ask for a configuration of a profile, either `src` or `other` the template will change all occurrences of `{ .Profile.Name }` to the name of the profile, no matter where it is inside the file.

```shell
% resticprofile -c examples/parse-error.toml -n src show
2020/11/06 21:39:48 cannot load configuration file: cannot parse toml configuration: While parsing config: (35, 6): duplicated tables
exit status 1
```

Run the command again, this time asking a display of the compiled version of the configuration:

```shell
% resticprofile -c examples/parse-error.toml -n src --trace show
2020/11/06 21:48:20 resticprofile 0.10.0-dev compiled with go1.15.3
2020/11/06 21:48:20 Resulting configuration for profile 'default':
====================
  1:
  2:
  3: [src]
  4: password-file = "/Users/CP/go/src/resticprofile/examples/default-key"
  5: repository = "/backup/Friday"
  6: lock = "$HOME/resticprofile-profile-default.lock"
  7: initialize = true
  8:
  9:     [src.backup]
 10:     source = "/Users/CP/go/src"
 11:     check-before = true
 12:     exclude = ["/**/.git"]
 13:     exclude-caches = true
 14:     tag = ["default", "dev"]
 15:
 16:
 17:     [default.retention]
 18:     after-backup = true
 19:     before-backup = false
 20:     compact = false
 21:     keep-within = "30d"
 22:     prune = true
 23:
 24:
 25:     [src.snapshots]
 26:     tag = ["default", "dev"]
 27:
 28: [other]
 29: password-file = "/Users/CP/go/src/resticprofile/examples/default-key"
 30: repository = "/backup/Friday"
 31: lock = "$HOME/resticprofile-profile-default.lock"
 32: initialize = true
 33:
 34:
 35:     [default.retention]
 36:     after-backup = true
 37:     before-backup = false
 38:     compact = false
 39:     keep-within = "30d"
 40:     prune = true
 41:
 42:
====================
2020/11/06 21:48:20 cannot load configuration file: cannot parse toml configuration: While parsing config: (35, 6): duplicated tables
exit status 1
 ```

 As you can see in lines 17 and 35, there are 2 sections of the same name. They could be both called `[src.retention]`, but actually the reason why they're both called `[default.retention]` is that resticprofile is doing a first pass to load the `[global]` section using a profile name of `default`.

 The fix for this configuration is very simple though, just remove the section name from the template:

```toml
{{ define "retention" }}
    after-backup = true
    before-backup = false
    compact = false
    keep-within = "30d"
    prune = true
{{ end }}

version = "1"

[src]
  password-file = "{{ .ConfigDir }}/{{ .Profile.Name }}-key"
  repository = "/backup/{{ .Now.Weekday }}"
  lock = "$HOME/resticprofile-profile-{{ .Profile.Name }}.lock"
  initialize = true

    [src.backup]
      source = "{{ .Env.HOME }}/go/src"
      check-before = true
      exclude = ["/**/.git"]
      exclude-caches = true
      tag = ["{{ .Profile.Name }}", "dev"]

    [src.retention]
      {{ template "retention" . }}

    [src.snapshots]
      tag = ["{{ .Profile.Name }}", "dev"]

[other]
  password-file = "{{ .ConfigDir }}/{{ .Profile.Name }}-key"
  repository = "/backup/{{ .Now.Weekday }}"
  lock = "$HOME/resticprofile-profile-{{ .Profile.Name }}.lock"
  initialize = true

    [other.retention]
      {{ template "retention" . }}
```

And now you no longer end up with duplicated sections.

## Documentation on template, variable expansion and other configuration scripting

There are a lot more you can do with configuration templates. If you're brave enough, [you can read the full documentation of the Go templates](https://golang.org/pkg/text/template/).

For a more end-user kind of documentation, you can also read [hugo documentation on templates](https://gohugo.io/templates/introduction/) which is using the same Go implementation, but don't talk much about the developer side of it.
Please note there are some functions only made available by hugo though, resticprofile adds its own set of functions.

## Template functions

resticprofile supports the following set of own functions in all templates:

* `{{ "some string" | contains "some" }}` => `true`
* `{{ "some string" | matches "^.+str.+$" }}` => `true`
* `{{ "some old string" | replace "old" "new" }}` => `some new string`
* `{{ "some old string" | replaceR "(old)" "$1 and new" }}` => `some old and new string`
* `{{ "some old string" | regex "(old)" "$1 and new" }}` => `some old and new string`
  (`regex` is an alias to `replaceR`) 
* `{{ "ABC" | lower }}` => `abc`
* `{{ "abc" | upper }}` => `ABC`
* `{{ "  A " | trim }}` => `A`
* `{{ "--A-" | trimPrefix "--" }}` => `A-`
* `{{ "--A-" | trimSuffix "-" }}` => `--A`
* `{{ range $v := "A,B,C" | split "," }} ({{ $v }}) {{ end }}` => ` (A)  (B)  (C) `
* `{{ "A,B,C" | split "," | join ";" }}` => `A;B;C`
* `{{ "A, B, C" | splitR "\\s*,\\s*" | join ";" }}` => `A;B;C`
* `{{ range $v := list "A" "B" "C" }} ({{ $v }}) {{ end }}` => ` (A)  (B)  (C) `
* `{{ with map "k1" "v1" "k2" "v2" }} {{ .k1 }}-{{ .k2 }} {{ end }}` => ` v1-v2 `
* `{{ with list "A" "B" "C" "D" | map }} {{ ._0 }}-{{ ._1 }}-{{ ._3 }} {{ end }}` => ` A-B-D `
* `{{ with list "A" "B" "C" "D" | map "key" }} {{ .key | join "-" }} {{ end }}` => ` A-B-C-D `
* `{{ tempDir }}` => `/tmp/resticprofile.../t` - unique OS specific existing temporary directory
* `{{ tempFile "filename" }}` => `/tmp/resticprofile.../t/filename` - unique OS specific existing temporary file

All `{{ temp* }}` functions guarantee that returned temporary directories and files are existing & writable. 
When resticprofile ends, temporary directories and files are removed.

The following functions can be used to encode data (e.g. when dealing with arbitrary input):

* `{{ "a & b\n" | js }}` => `a \u0026 b\u000A` - JSON string equivalent of the input (*builtin*)
* `{{ "a & b\n" | html }}` => `a &amp; b\n` - HTML text escaped input (*builtin*)
* `{{ "a & b\n" | urlquery }}` => `a+%26+b%0A` - URL query escaped input (*builtin*)
* `{{ "plain" | base64 }}` => `cGxhaW4=` - Base64 encoded input
* `{{ "plain" | hex }}` => `706c61696e` - Hexadecimal encoded input

{{% notice hint %}}
Encode with `js` when creating **strings** in *YAML*, *TOML* or *JSON* configuration files, e.g.: `"{{ .Env.MyVar | js }}"`. 
This ensures the markup remains correct and can be parsed regardless of the input data.
{{% /notice %}}


Please refer to the [official documentation](https://golang.org/pkg/text/template/) for the general syntax and set of default functions provided in go text templates. 
