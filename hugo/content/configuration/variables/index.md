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

This is obviously not a real world example, but it shows many of the possibilities you can do with variable expansion.

To check the generated configuration, you can use the resticprofile `show` command:

```
% resticprofile -c examples/template.yaml -n src show

global:
    default-command:  snapshots
    restic-binary:    restic
    min-memory:       100


src:
    backup:
        check-before:    true
        run-before:      echo Hello CP
                         echo current dir: /Users/CP/go/src/resticprofile
                         echo config dir: /Users/CP/go/src/resticprofile/examples
                         echo profile started at 04 Nov 20 21:56 GMT
        run-after:       echo All Done!
        source:          /Users/CP/go/src
        tag:             src
                         dev
        exclude:         /**/.git
        exclude-caches:  true

    retention:
        after-backup:  true
        keep-within:   30d
        prune:         true
        tag:           src
                       dev

    repository:     /backup/Wednesday
    password-file:  /Users/CP/go/src/resticprofile/examples/src-key
    initialize:     true
    lock:           /Users/CP/resticprofile-profile-src.lock
    snapshots:
        tag:  src
              dev
```

As you can see, the `src` profile inherited from the `generic` profile. The tags `{{ .Profile.Name }}` got replaced by the name of the current profile `src`. Now you can reuse the same generic configuration in another profile.

Here's another example of an HCL configuration on Linux where I use a variable `$mountpoint` set to a USB drive mount point:

```hcl
global {
    priority = "low"
    ionice = true
    ionice-class = 2
    ionice-level = 6
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

# Configuration templates

Templates are a great way to compose configuration profiles.

Please keep in mind that `yaml` files are sensitive to the number of spaces. Also if you declare a block already declared, it overrides the previous declaration (instead of merging them).

For that matter, configuration template is probably more useful if you use the `toml` or `hcl` configuration format.

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

```toml
#
# This is an example of TOML configuration using nested templates
#

# nested template declarations
# this template declaration won't appear here in the configuration file
# it will only appear when called by {{ template "backup_root" . }}
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

# Debugging your template and variable expansion

If for some reason you don't understand why resticprofile is not loading your configuration file, you can display the generated configuration after executing the template (and replacing the variables and everything) using the `--trace` flag.

# Limitations of using templates

There's something to be aware of when dealing with templates: at the time the template is compiled, it has no knowledge of what the end configuration should look like: it has no knowledge of profiles for example. Here is a **non-working** example of what I mean:

```toml
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

```
% resticprofile -c examples/parse-error.toml -n src show
2020/11/06 21:39:48 cannot load configuration file: cannot parse toml configuration: While parsing config: (35, 6): duplicated tables
exit status 1
```

Run the command again, this time asking a display of the compiled version of the configuration:

```
% go run . -c examples/parse-error.toml -n src --trace show
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

# Documentation on template, variable expansion and other configuration scripting

There are a lot more you can do with configuration templates. If you're brave enough, [you can read the full documentation of the Go templates](https://golang.org/pkg/text/template/).

For a more end-user kind of documentation, you can also read [hugo documentation on templates](https://gohugo.io/templates/introduction/) which is using the same Go implementation, but don't talk much about the developer side of it.
Please note there are some functions only made available by hugo though.
