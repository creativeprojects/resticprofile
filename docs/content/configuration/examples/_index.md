---
title: "Examples"
date: 2022-04-24T09:44:47+01:00
weight: 5
---

## Simple configuration using Azure storage

Here's a simple configuration file using a Microsoft Azure backend. You will notice that the `env` section lets you define environment variables:

{{< tabs groupid="config-with-hcl" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[default]
  repository = "azure:restic:/"
  password-file = "key"
  option = "azure.connections=3"

  [default.env]
    AZURE_ACCOUNT_NAME = "my_storage_account"
    AZURE_ACCOUNT_KEY = "my_super_secret_key"

  [default.backup]
    exclude-file = "excludes"
    exclude-caches = true
    one-file-system = true
    tag = [ "root" ]
    source = [ "/", "/var" ]
```
{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

default:
  repository: "azure:restic:/"
  password-file: "key"
  option: "azure.connections=3"

  env:
    AZURE_ACCOUNT_NAME: "my_storage_account"
    AZURE_ACCOUNT_KEY: "my_super_secret_key"

  backup:
    exclude-file: "excludes"
    exclude-caches: true
    one-file-system: true
    tag:
      - "root"
    source:
      - "/"
      - "/var"
```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
default {
    repository = "azure:restic:/"
    password-file = "key"
    options = "azure.connections=3"

    env {
      AZURE_ACCOUNT_NAME = "my_storage_account"
      AZURE_ACCOUNT_KEY = "my_super_secret_key"
    }

    backup = {
        exclude-file = "excludes"
        exclude-caches = true
        one-file-system = true
        tag = [ "root" ]
        source = [ "/", "/var" ]
    }
}
```

{{% /tab %}}
{{< /tabs >}}

## Configuration with inheritance

Here's a more complex configuration file showing profile inheritance and two backup profiles using the same repository:

{{< tabs groupid="config-with-hcl" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[global]
  # ionice is available on Linux only
  ionice = false
  ionice-class = 2
  ionice-level = 6
  # priority is using priority class on windows, and "nice" on unixes
  priority = "low"
  # run 'snapshots' when no command is specified when invoking resticprofile
  default-command = "snapshots"
  # initialize a repository if none exist at location
  initialize = false
  # resticprofile won't start a profile if there's less than 100MB of RAM available
  min-memory = 100

# a group is a profile that will call all profiles one by one
[groups]
  # when starting a backup on profile "full-backup",
  # it will run the "root" and "src" backup profiles
  full-backup = [ "root", "src" ]

# Default profile when not specified on the command line (-n or --name)
# There's no default inheritance from the 'default' profile,
# you can use the 'inherit' flag if needed
[default]
  # you can use a relative path, it will be relative to the configuration file
  repository = "/backup"
  password-file = "key"
  initialize = false
  # will run these scripts before and after each command (including 'backup')
  run-before = "mount /backup"
  run-after = "umount /backup"
  # if a restic command fails, the run-after won't be running
  # add this parameter to run the script in case of a failure
  run-after-fail = "umount /backup"

  # add environment variables
  [default.env]
    TMPDIR= "/tmp"

# New profile named 'root'
[root]
  inherit = "default"
  initialize = true
  # LOCAL lockfile so you cannot run the same profile more than once at a time
  # (it's totally independent of the restic locks on the repository)
  lock = "/tmp/resticprofile-root.lock"

  # 'backup' command of profile 'root'
  [root.backup]
    # files with no path are relative to the configuration file
    exclude-file = [ "root-excludes", "excludes" ]
    exclude-caches = true
    one-file-system = false
    tag = [ "test", "dev" ]
    source = [ "/" ]
    # ignore restic warnings when files cannot be read
    no-error-on-warning = true
    # run every day at midnight
    schedule = "daily"
    schedule-permission = "system"
    schedule-lock-wait = "2h"

  # retention policy for profile root
  # retention is a special section that run the "forget" command
  # before or after a backup
  [root.retention]
    before-backup = false
    after-backup = true
    keep-hourly = 1
    keep-daily = 1
    keep-weekly = 1
    keep-monthly = 1
    keep-within = "3h"
    keep-tag = [ "forever" ]
    prune = false
    # tag can be a boolean ('true' meaning to copy tag set from 'backup') 
    # or a custom set of tags.
    # Default is 'false', meaning that tags are NOT used.
    tag = true
    # host can be a boolean ('true' meaning current hostname)
    # or a string to specify a different hostname
    host = true

# New profile named 'src'
[src]
  inherit = "default"
  initialize = true

  # 'backup' command of profile 'src'
  [src.backup]
    exclude = [ '/**/.git' ]
    exclude-caches = true
    one-file-system = false
    tag = [ "test", "dev" ]
    source = [ "./src" ]
    check-before = true
    # will only run these scripts before and after a backup
    run-before = [ "echo Starting!", "ls -al ./src" ]
    run-after = "sync"
    # if scheduled, will run every 30 minutes
    schedule = "*:0,30"
    schedule-permission = "user"
    schedule-lock-wait = "10m"

    # retention policy for profile src
    [src.retention]
    before-backup = false
    after-backup = true
    keep-within = "30d"
    prune = true

  # check command of profile src
  [src.check]
    read-data = true
    # if scheduled, will check the repository the first day of each month at 3am
    schedule = "*-*-01 03:00"

```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

global:
  # run 'snapshots' when no command is specified when invoking resticprofile
  default-command: snapshots
  # initialize a repository if none exist at location
  initialize: false
  # priority is using priority class on windows, and "nice" on unixes
  priority: low
  # resticprofile won't start a profile if there's less than 100MB of RAM available
  min-memory: 100

# a group is a profile that will call all profiles one by one
groups:
  # when starting a backup on profile "full-backup",
  # it will run the "root" and "src" backup profiles
  full-backup:
    - root
    - src

# Default profile when not specified on the command line (-n or --name)
# There's no default inheritance from the 'default' profile,
# you can use the 'inherit' flag if needed
default:
  # add environment variables
  env:
    TMPDIR: /tmp
  password-file: key
  # you can use a relative path, it will be relative to the configuration file
  repository: /backup
  # will run these scripts before and after each command (including 'backup')
  run-before: mount /backup
  run-after: umount /backup
  # if a restic command fails, the run-after won't be running
  # add this parameter to run the script in case of a failure
  run-after-fail: umount /backup

# New profile named 'root'
root:
  inherit: default
  initialize: true
  # LOCAL lockfile so you cannot run the same profile more than once at a time
  # (it's totally independent of the restic locks on the repository)
  lock: /tmp/resticprofile-root.lock

  backup:
    exclude-caches: true
    # files with no path are relative to the configuration file
    exclude-file:
      - root-excludes
      - excludes
    one-file-system: false
    source:
      - /
    tag:
      - test
      - dev
    # ignore restic warnings when files cannot be read
    no-error-on-warning: true
    # run every day at midnight
    schedule: daily
    schedule-permission: system
    schedule-lock-wait: 2h

  # retention policy for profile root
  # retention is a special section that run the "forget" command
  # before or after a backup
  retention:
    before-backup: false
    after-backup: true
    keep-daily: 1
    keep-hourly: 1
    keep-weekly: 1
    keep-monthly: 1
    keep-within: 3h
    keep-tag:
      - forever
    prune: false
    # tag can be a boolean ('true' meaning to copy tag set from 'backup') 
    # or a custom set of tags.
    # Default is 'false', meaning that tags are NOT used.
    tag: true
    # host can be a boolean ('true' meaning current hostname)
    # or a string to specify a different hostname
    host: true

# New profile named 'src'
src:
  lock: "/tmp/resticprofile-profile-src.lock"
  force-inactive-lock: false
  inherit: default
  initialize: true

  # 'backup' command of profile 'src'
  backup:
    check-before: true
    exclude:
      - /**/.git
    exclude-caches: true
    one-file-system: false
    # will only run these scripts before and after a backup
    run-before:
      - echo Starting!
      - ls -al ~/go
    run-after: echo All Done!
    source:
      - ~/go
    tag:
      - test
      - dev
    # run every 30 minutes
    schedule: "*:0,30"
    schedule-permission: user
    schedule-lock-wait: 10m

  # retention policy for profile src
  retention:
    before-backup: false
    after-backup: true
    keep-within: 30d
    prune: true

  # check command of profile src
  check:
    read-data: true
    # check the repository the first day of each month at 3am
    schedule: "*-*-01 03:00"

```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
global {
    priority = "low"
    ionice = true
    ionice-class = 2
    ionice-level = 6
    # don't start if the memory available is < 1000MB
    min-memory = 1000
}

groups {
    all = ["src", "self"]
}

default {
    repository = "/tmp/backup"
    password-file = "key"
    run-before = "echo Profile started!"
    run-after = "echo Profile finished!"
    run-after-fail = "echo An error occurred!"
}


src {
    inherit = "default"
    initialize = true
    lock = "/tmp/backup/resticprofile-profile-src.lock"
    force-inactive-lock = false

    snapshots = {
        tag = [ "test", "dev" ]
    }

    backup = {
        run-before = [ "echo Starting!", "ls -al ~/go/src" ]
        run-after = "echo All Done!"
        exclude = [ "/**/.git" ]
        exclude-caches = true
        tag = [ "test", "dev" ]
        source = [ "~/go/src" ]
        check-before = true
    }

    retention = {
        before-backup = false
        after-backup = true
        keep-last = 3
        compact = false
        prune = true
    }

    check = {
        check-unused = true
        with-cache = false
    }
}

self {
    inherit = "default"
    initialize = false

    snapshots = {
        tag = [ "self" ]
    }

    backup = {
        source = "./"
        tag = [ "self" ]
    }
}

```

{{% /tab %}}
{{< /tabs >}}

## Configuration example for Windows


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

## Use stdin in configuration

Simple example sending a file via stdin

{{< tabs groupid="config-with-hcl" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[stdin]
  repository = "local:/backup/restic"
  password-file = "key"

  [stdin.backup]
    stdin = true
    stdin-filename = "stdin-test"
    tag = [ 'stdin' ]
  
[mysql]
  inherit = "stdin"

  [mysql.backup]
    stdin-command = [ 'mysqldump --all-databases --order-by-primary' ]
    stdin-filename = "dump.sql"
    tag = [ 'mysql' ]

```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

stdin:
  repository: "local:/backup/restic"
  password-file: key
  backup:
    stdin: true
    stdin-filename: stdin-test
    tag:
      - stdin

mysql:
  inherit: stdin
  backup:
    stdin-command: "mysqldump --all-databases --order-by-primary"
    stdin-filename: "dump.sql"
    tag:
      - mysql

```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
# sending stream through stdin
stdin {
    repository = "local:/backup/restic"
    password-file = "key"

    backup = {
        stdin = true
        stdin-filename = "stdin-test"
        tag = [ "stdin" ]
    }
}

mysql {
  inherit = "stdin"

  backup = {
    stdin-command = [ "mysqldump --all-databases --order-by-primary" ]
    stdin-filename = "dump.sql"
    tag = [ "mysql" ]
  }
}
```

{{% /tab %}}
{{< /tabs >}}

