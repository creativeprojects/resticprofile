---
title: "Inheritance"
date: 2022-05-02T20:00:00+02:00
weight: 16
---

{{% notice style="tip" %}}
You can use `resticprofile [<profile-name>.]show` to see the effect inheritance on a profile
{{% /notice %}}

## Profile Inheritance

Profiles can inherit from a parent allowing to configure a base profile that defines the general behavior and holds common configurations while **derived** profiles only define what is specific, e.g. what needs to be backed up or which additional commands (`run-before`, `run-after` & `run-finally`) must be run.

When assuming profile "backup-homes" inherits from profile "base", the profile configuration of "backup-homes" is merged into "base" to build the effective configuration of "backup-homes". 

Profile configuration merging follows the same logic as [Configuration Merging](../include/#configuration-merging) in includes: 
* What defines in the parent profile is replaced by definitions in the derived profile
* Configuration structure is merged, configuration properties are replaced
* A profile declares that it inherits from a parent by setting the property `inherit` to the name of the parent profile.


{{< tabs groupId="profile-inheritance-example" >}}
{{% tab name="Profile \"base\" (yaml)" %}}

```yaml
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
{{% tab name="Profile \"backup-homes\" (yaml)" %}}

```yaml
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
{{% tab name="... after applying inherit" %}}

```yaml
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

This differs from how includes handle lists and may lead to unexpected results. In configuration file format **version 2** the behavior was changed to match that of [includes](../include/#configuration-merging). See [Mixins](#mixins) for a deterministic way of pre/appending to list properties instead.

{{% /notice %}}


## Mixins

Starting with configuration file format **version 2**, mixins offer an easy way to share pieces of configurations between profiles without forcing a hierarchy of inheritance. Mixins can be used at every level within the profile configuration, support parametrisation (`vars`) and can prepend or append to list properties in addition to replacing them.

Mixins are declared in section `mixins` as a named objects. The contents of these objects are merged into the profile configuration wherever a `use` property references (uses) the mixin. 
Configuration merging is following the same logic as used in [inheritance](#profile-inheritance) and [includes](../include/#configuration-merging). When `use` references multiple mixins, the mixins apply in the order they are referenced and can override each other (mixins referenced later override what earlier mixins defined).

Configuration values inside a mixin may use variables following the syntax `${variable}` or `$variable`. Default values can be defined inside the mixin and `use` can set alternate values before merging the mixin.

{{< tabs groupId="config-with-mixins" >}}
{{% tab name="yaml" %}}

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
{{% tab name="yaml (with vars)" %}}

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
{{% tab name="toml" %}}

```toml
version = 2

[mixins.name-of-mixin]
config-key = "config-value"

[profiles.profile]
# set config-key to config-value in "profile"
use = "name-of-mixin"
```

{{% /tab %}}
{{% tab name="toml (with vars)" %}}

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
[[profiles.profile-non-default.use]]
name = "name-of-mixin"
WHAT = "Mixin"
```

{{% /tab %}}
{{% /tabs %}}

#### Named Mixin Declaration

Every mixin object below the `mixins` section has the following structure (all properties are optional):

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

| Property                    | Purpose                                                               |
|-----------------------------|-----------------------------------------------------------------------|
| `name`                      | Name of the mixin to use and merge in place of the `use` property     |
| `vars`: `<variable-name>`   | Set `<variable-name>` before using the mixin                          |
| `<variable-name>`           | Set `<variable-name>` (except `name` & `vars`) before using the mixin |

Mixins are applied to the configuration after processing all [includes](../include/) but prior to [profile inheritance](#profile-inheritance) which means the `use` properties are not inherited but the result of applying `use` is inherited instead. What is defined by a mixin in a parent profile can still be overridden by a definition in a derived profile, but derived profiles can not change which mixins apply to their parent.

#### Mixin Example

{{< tabs groupId="config-with-mixins-examples" >}}
{{% tab name="yaml" %}}

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
{{% tab name="toml" %}}

```toml
# file format version 2
version = 2

# mixin declarations
[mixins]
  [mixins.alternate-repository]
  repository = "local:/backup/alternate"
  password-file = "alternate-repo.key"
  
  [mixins.retain-last]
    [mixins.retain-last.default-vars]
    LAST = 30
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
    [[profiles.select-all-and-retain-last-60.use]]
    name = "alternate-repository"
      
    [[profiles.select-all-and-retain-last-60.use]]
    name = "retain-last"
    LAST = 60

    [profiles.select-all-and-retain-last-60.backup]
    source = "/"
```

{{% /tab %}}
{{% /tabs %}}


## Common Flags

Profiles in resticprofile configure commandline options (flags) for running restic commands. While a profile has several predefined common properties (`repository`, `password-file`, ...), any arbitrary common flags can be set directly inside the profile and will be inherited by all command sections of the profile.

An arbitrary flag like `insecure-tls` that is not part of the profile config reference but valid for every restic command can be set at profile level and will be converted to a restic flag. See the following example:

{{< tabs groupId="config-with-common-flags-in-profile" >}}
{{% tab name="yaml" %}}

```yaml
default:
  repository: rest:https://backup-host/my-repo
  insecure-tls: true
  backup:
    source: / 
```

{{% /tab %}}
{{% tab name="toml" %}}

```toml
[default]
repository = "rest:https://backup-host/my-repo"
insecure-tls = true
[default.backup]
source = "/"
```

{{% /tab %}}
{{% /tabs %}}

The config above results in the following restic commandline:

* `resticprofile backup` executes \
  `restic backup --repo rest:https://backup-host/my-repo --insecure-tls /`
* `resticprofile prune` executes \
  `restic prune --repo rest:https://backup-host/my-repo --insecure-tls`


{{% notice style="tip" %}}
The option `--dry-run` prints restic commands to console or log file. To see what flags are effectively used with each *restic command* involved in *backup*, use `resticprofile --dry-run [<profile-name>.]backup`.
{{% /notice %}}
