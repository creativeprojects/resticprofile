---
title: "Inheritance"
date: 2022-05-02T20:00:00+02:00
weight: 16
---

{{% notice style="tip" %}}
You can use `resticprofile [<profile-name>.]show` to see the effect inheritance has on a profile
{{% /notice %}}

## Profile Inheritance

Profiles can inherit from a parent profile. This allows to define the general behavior and common configuration in a base profile while **derived** profiles only define what is specific, e.g. what needs to be included in the backup or which command [hooks]({{< ref "/configuration/run_hooks" >}}) (e.g. `run-before`, `run-after` & `run-finally`) must be started.

When assuming profile "*backup-homes*" inherits from profile "*base*", then the effective configuration of "*backup-homes*" is built by merging the profile configuration of "*backup-homes*" into "*base*". 

Profile configuration merging follows the same logic as [configuration merging]({{< ref "/configuration/include/#configuration-merging" >}}) in includes: 
* What defines in the parent profile is replaced by definitions from the derived profile
* Configuration structure is merged, configuration properties are replaced
* A profile declares that it inherits from a parent by setting the property `inherit` to the name of the parent profile
* There is no default inheritance. If `inherit` is not set, no inheritance applies


{{< tabs groupid="profile-inheritance-example" >}}
{{% tab title="Profile 'base' (yaml)" %}}

```yaml
version: "1"

base:
  initialize: true
  repository: local:/backup/my-repo
  password-file: my-repo.key
  
  retention:
    after-backup: true
    keep-last: 2
    keep-hourly: 1
    keep-daily: 1
    keep-weekly: 1
    
  backup:
    exclude:
      - "*."
      - "*~"
      - "/backup/*"
    source:
      - /

```

{{% /tab %}}
{{% tab title="Profile 'backup-homes' (yaml)" %}}

<!-- checkdoc-ignore -->
```yaml
version: "1"

backup-homes:
  inherit: base
  
  

  retention:
    
    
    keep-hourly: false
    keep-daily: 30
    keep-weekly: 26

  backup:
    
    
    
    
    source:
      - /home/
```

{{% /tab %}}
{{% tab title="... after applying 'inherit'" %}}

```yaml
version: "1"

backup-homes:
  initialize: true
  repository: local:/backup/my-repo
  password-file: my-repo.key

  retention:
    after-backup: true
    keep-last: 2
    keep-hourly: false
    keep-daily: 30
    keep-weekly: 26

  backup:
    exclude:
      - "*."
      - "*~"
      - "/backup/*"
    source:
      - /home/
```

{{% /tab %}}
{{% /tabs %}}


{{% notice style="note" %}}

Configurations prior to **version 2**, treat lists as if they were configuration structure. Instead of replacing the parent with the derived list entirely, a derived list is **merged** into the parent list using **list-index** as key.

This differs from how includes handle lists and may lead to unexpected results. In configuration file format **version 2** the behavior was changed to match that of [includes]({{< ref "/configuration/include/#configuration-merging" >}}) and extended with a deterministic way of pre- & appending to list properties.

{{% /notice %}}

## Inheritance of List Properties

Starting with configuration format **version 2**, lists are no longer considered configuration structure and are replaced in derived profiles in the same way as inheritance behaves for any non-list properties. For example, when the parent and child profile define the same list property like `run-before` or `source`, the declaration of the child property replaces the declaration of the parent property entirely.

For **version 2**, when the parent defined `source = ['/my-files1', '/my-files2']` and the child `source = ['/my-other-files']`, then only `/my-other-files` will really make it into the backup.

In contrast to this, configurations in **version 1** partially merge lists on the list index. E.g. when the parent profile defines 2 items and the child only one, then the first entry in parent is replaced with the single child item and the second parent item is derived into the child profile.

For **version 1**, when the parent defined `source = ['/my-files1', '/my-files2']` and the child `source = ['/my-other-files']`, then `/my-other-files` **and** `/my-files2` will make it into the backup.

### Prepend & Append to List Properties

{{% notice style="warning" title="Config format version 2" %}}
**Feature preview**, may change without notice
{{% /notice %}}

Inheritance in configuration format **version 2** can prepend and append to parent list properties. This feature replaces list merging of version 1.

Assuming the parent profile declares the list property `<list-property>`:
* `<list-property>...` or `<list-property>__APPEND` appends to the list property
* `...<list-property>` or `<list-property>__PREPEND` prepends to the list property

{{< tabs groupid="config-with-inheritance-list-append" >}}
{{% tab title="yaml" %}}

```yaml
version: 2

profiles:
  
  default:
    backup:
      exclude:
        - '.*'
        - '~*'

  derived-profile:
    inherit: default
    backup:
      exclude...: '.git'
      source: '/myrepo'
```

{{% /tab %}}
{{% tab title="toml" %}}

```toml
version = 2

[profiles.default.backup]
exclude = ['.*', '~*']

[profiles.derived-profile]
inherit = 'default'

[profiles.derived-profile.backup]
exclude__APPEND = '.git'
source = '/myrepo'
```

{{% /tab %}}
{{% /tabs %}}

In the examples above, the final value of `exclude` in `derived-profile` is `['.*', '~*', '.git']`.

## Mixins

{{% notice style="warning" title="Config format version 2" %}}
**Feature preview**, may change without notice
{{% /notice %}}

Mixins offer an easy way to share pieces of configuration between profiles without forcing a hierarchy of inheritance. Mixins can be used at every level within the profile configuration, support parametrisation (`vars`) and similar to hierarchic inheritance, they can prepend or append to list properties in addition to setting or replacing properties.

Mixins are declared in section `mixins` as named objects. The contents of these objects are merged into the profile configuration wherever a `use` property references (uses) the mixin. 
Configuration merging is following the same logic as used in [inheritance](#profile-inheritance) and [includes]({{< ref "/configuration/include/#configuration-merging" >}}). When `use` references multiple mixins, the mixins apply in the order they are referenced and can override each other (mixins referenced later override what earlier mixins defined).

Configuration values inside a mixin may be parametrized with variables following the syntax `${variable}` or `$variable`. Defaults for variables can be defined inside the mixin with `default-vars` and `use` can specify variables before merging the mixin. In difference to configuration [variables]({{< ref "/configuration/variables" >}}) that expand prior to parsing, mixin variables expand when the mixin is merged and for this reason the syntax differs.

Unlike configuration [variables]({{< ref "/configuration/variables" >}}) and [templates]({{< ref "/configuration/templates" >}}), mixins create parsed configuration structure not config markup that requires parsing. This allows mixins to be defined in one supported config format (`yaml`, `toml`, `json`) while being used in any other supported format when the configuration is split into multiple [includes]({{< ref "/configuration/include/#configuration-merging" >}}).

{{< tabs groupid="config-with-mixins" >}}
{{% tab title="yaml" %}}

```yaml
version: 2

mixins:
  name-of-mixin: 
    config-key: config-value

profiles:
  profile:
    # set config-key to config-value in "profile"
    use: name-of-mixin
```

{{% /tab %}}
{{% tab title="yaml (with vars)" %}}

```yaml
version: 2

mixins:
  name-of-mixin: 
    default-vars:
      WHAT: World
    parametrized-config-key: Hello $WHAT

profiles:

  profile:
    # set parametrized-config-key to "Hello World" in "profile"
    use: name-of-mixin

  profile-non-default:
    # set parametrized-config-key to "Hello Mixin" in "profile-non-default"
    use: 
      - name: name-of-mixin
        WHAT: "Mixin"
```

{{% /tab %}}
{{% tab title="toml" %}}

```toml
version = 2

[mixins.name-of-mixin]
config-key = "config-value"

[profiles.profile]
# set config-key to config-value in "profile"
use = "name-of-mixin"
```

{{% /tab %}}
{{% tab title="toml (with vars)" %}}

```toml
version = 2

[mixins.name-of-mixin]
parametrized-config-key = "Hello $WHAT"
[mixins.name-of-mixin.default-vars]
WHAT = "World"

[profiles.profile]
# set parametrized-config-key to "Hello World" in "profile"
use = "name-of-mixin"

[profiles.profile-non-default]
# set parametrized-config-key to "Hello Mixin" in "profile-non-default"
use = { name = "name-of-mixin", WHAT = "Mixin" }
```

{{% /tab %}}
{{% /tabs %}}

#### Named Mixin Declaration

Every named mixin object below the `mixins` section has the following structure (all properties are optional):

| Property                                       | Purpose                                           |
|------------------------------------------------|---------------------------------------------------|
| `default-vars`: `<variable-name>`              | Default value for variable `$<variable-name>`     |
| `<config-key>`                                 | Set `<config-key>` when the mixin is used         |
| `<config-key>`: `<sub-key>`                    | Set `<sub-key>` below `<config-key>`              |
| `<config-key>...` or `<config-key>__APPEND`    | Change `<config-key>` to a list and append to it  |
| `...<config-key>`  or  `<config-key>__PREPEND` | Change `<config-key>` to a list and prepend to it |

#### Mixin Usage

The `use` property can be placed at any depth inside the profile configuration and is referencing a single mixin, a list of mixin names or a list of names and use-objects.

Every use object within the `use` list has the following structure:

| Property                       | Purpose                                                           |
|--------------------------------|-------------------------------------------------------------------|
| `name`                         | Name of the mixin to use and merge in place of the `use` property |
| `vars`: `<variable-name>`      | Set mixin variable `$<variable-name>`                             |
| `<variable-name>`              | Set mixin variable `$<variable-name>` (short syntax)              |

Mixins are applied to the configuration after processing all [includes]({{< ref "/configuration/include/" >}}) but prior to [profile inheritance](#profile-inheritance) which means the `use` properties are not inherited but the result of applying `use` is inherited instead. What is defined by a mixin in a parent profile can still be overridden by a definition in a derived profile, but derived profiles can not change which mixins apply to their parent.

List properties that have been inherited from a parent can be modified (append/prepend) and replaced by a mixin.

#### Mixin Example

{{< tabs groupid="config-with-mixins-examples" >}}
{{% tab title="yaml" %}}

```yaml
# file format version 2
version: 2

# mixin declarations
mixins:
  alternate-repository: 
    repository: local:/backup/alternate
    password-file: alternate-repo.key

  retain-last:
    default-vars:
      LAST: 30
    retention:
      keep-last: $LAST
      keep-hourly: false
      keep-daily: false
      keep-weekly: false

  exclude-backup:
    exclude...:
      - "/backup/*"
      - "*.bak*"

  exclude-hidden:
    exclude...:
      - "*."
      - "*~"

# profile declarations
profiles:
  select-some-and-retain-last-30:
    use: 
      - alternate-repository
      - retain-last

    backup:
      use: 
        - exclude-backup
        - exclude-hidden
      exclude: /tmp
      source: /

  select-all-and-retain-last-60:
    use:
      - alternate-repository
      - name: retain-last
        LAST: 60
    backup:
      source: /
```

{{% /tab %}}
{{% tab title="toml" %}}

```toml
# file format version 2
version = 2

# mixin declarations
[mixins]
  [mixins.alternate-repository]
  repository = "local:/backup/alternate"
  password-file = "alternate-repo.key"
  
  [mixins.retain-last]
    default-vars = { LAST = 30 }
    [mixins.retain-last.retention]
    keep-last = "$LAST"
    keep-hourly = false
    keep-daily = false
    keep-weekly = false
  
  [mixins.exclude-backup]
  exclude__APPEND = [
      "/backup/*",
      "*.bak*",
  ] 
  
  [mixins.exclude-hidden]
  exclude__APPEND = [
      "*.", 
      "*~",
  ]

# profile declarations
[profiles]
  [profiles.select-some-and-retain-last-30]
  use = ["alternate-repository", "retain-last"]
  
    [profiles.some-keep-last-30.backup]
    use = ["exclude-backup", "exclude-hidden"]
    exclude = "/tmp"
    source = "/"
  
  
  [profiles.select-all-and-retain-last-60]
  use = [
      "alternate-repository",
      { name = "retain-last", LAST = 60 },
  ]
    [profiles.select-all-and-retain-last-60.backup]
    source = "/"
```

{{% /tab %}}
{{% /tabs %}}


## Common Flags

Profiles in resticprofile configure commandline options (flags) for restic commands. While a profile has several predefined common properties (`repository`, `password-file`, ...), any arbitrary common flags can be set directly inside the profile and will be inherited by all command sections of the profile. 

Resticprofile applies a filter (see `global.restic-arguments-filter`) to decide which flags are supported in which restic commands and automatically removes unsupported flags when building commandline options.

For example, a flag like `insecure-tls` can be set at profile level and will be used whenever restic is started with this profile. Most supported flags can be set in this way at profile level, see [reference]({{< ref "/configuration/reference" >}}) for details.

{{< tabs groupid="config-with-common-flags-in-profile" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[default]
repository = "rest:https://backup-host/my-repo"
insecure-tls = true
[default.backup]
source = "/"
```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

default:
  repository: rest:https://backup-host/my-repo
  insecure-tls: true
  backup:
    source: / 
```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
default {
    repository = "rest:https://backup-host/my-repo"
    insecure-tls = true
    backup {
        source = "/"
    }
}
```

{{% /tab %}}
{{% tab title="json" %}}

```json
{
  "version": "1",
  "default": {
    "repository": "rest:https://backup-host/my-repo",
    "insecure-tls": true,
    "backup": {
      "source": "/"
    }
  }
}
```

{{% /tab %}}
{{% /tabs %}}

Resulting in the following restic commandline:

```
resticprofile --dry-run backup
...
dry-run: /usr/local/bin/restic backup --insecure-tls --repo rest:https://backup-host/my-repo /


resticprofile --dry-run prune
...
dry-run: /usr/local/bin/restic prune --insecure-tls --repo rest:https://backup-host/my-repo
```


{{% notice style="tip" %}}
The option `--dry-run` prints restic commands to console or log file. To see what flags are effectively used with each *restic command* involved in *backup*, use `resticprofile --dry-run [<profile-name>.]backup`.
{{% /notice %}}
