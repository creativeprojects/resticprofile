---
title: "Examples"
date: 2022-04-24T09:44:47+01:00
weight: 5
---


Here's a simple configuration file using a Microsoft Azure backend:

## Simple TOML configuration
```toml
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

## Simple YAML configuration
```yaml
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

## More complex configuration in TOML

Here's a more complex configuration file showing profile inheritance and two backup profiles using the same repository:

```toml
[global]
# ionice is available on Linux only
ionice = false
ionice-class = 2
ionice-level = 6
# priority is using priority class on windows, and "nice" on unixes - it's acting on CPU usage only
priority = "low"
# run 'snapshots' when no command is specified when invoking resticprofile
default-command = "snapshots"
# initialize a repository if none exist at location
initialize = false
# resticprofile won't start a profile if there's less than 100MB of RAM available
min-memory = 100

# a group is a profile that will call all profiles one by one
[groups]
# when starting a backup on profile "full-backup", it will run the "root" and "src" backup profiles
full-backup = [ "root", "src" ]

# Default profile when not specified (-n or --name)
# Please note there's no default inheritance from the 'default' profile (you can use the 'inherit' flag if needed)
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

[default.env]
TMPDIR= "/tmp"

[no-cache]
inherit = "default"
no-cache = true
initialize = false

# New profile named 'root'
[root]
inherit = "default"
initialize = true
# this will add a LOCAL lockfile so you cannot run the same profile more than once at a time
# (it's totally independent of the restic locks on the repository)
lock = "/tmp/resticprofile-root.lock"
force-inactive-lock = false

# 'backup' command of profile 'root'
[root.backup]
# files with no path are relative to the configuration file
exclude-file = [ "root-excludes", "excludes" ]
exclude-caches = true
one-file-system = false
tag = [ "test", "dev" ]
source = [ "/" ]
# if scheduled, will run every day at midnight
schedule = "daily"
schedule-permission = "system"
schedule-lock-wait = "2h"
# run this after a backup to share a repository between a user and root (via sudo)
run-after = "chown -R $SUDO_USER $HOME/.cache/restic /backup"
# ignore restic warnings (otherwise the backup is considered failed when restic couldn't read some files)
no-error-on-warning = true

# retention policy for profile root
[root.retention]
before-backup = false
after-backup = true
keep-last = 3
keep-hourly = 1
keep-daily = 1
keep-weekly = 1
keep-monthly = 1
keep-yearly = 1
keep-within = "3h"
keep-tag = [ "forever" ]
compact = false
prune = false
# path can be a boolean ('true' meaning to copy source paths from 'backup') 
# or a path or list of paths to use instead. Default is `true` if not specified.
#path = []
# tag can be a boolean ('true' meaning to copy tag set from 'backup') 
# or a custom set of tags. Default is 'false', meaning that tags are NOT used.
tag = true
# host can be a boolean ('true' meaning current hostname) or a string to specify a different hostname
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
compact = false
prune = true

# check command of profile src
[src.check]
read-data = true
# if scheduled, will check the repository the first day of each month at 3am
schedule = "*-*-01 03:00"

```

## TOML configuration example for Windows

And another simple example for Windows:

```toml
[global]
restic-binary = "c:\\ProgramData\\chocolatey\\bin\\restic.exe"

# Default profile when not specified (-n or --name)
# Please note there's no default inheritance from the 'default' profile (you can use the 'inherit' flag if needed)
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
# ignore restic warnings (otherwise the backup is considered failed when restic couldn't read some files)
no-error-on-warning = true
```

## Use stdin in configuration

Simple example sending a file via stdin

```toml

[stdin]
repository = "local:/backup/restic"
password-file = "key"

[stdin.backup]
stdin = true
stdin-filename = "stdin-test"
tag = [ 'stdin' ]

```

## Special case for the `copy` command section

The copy command needs two repository (and quite likely 2 different set of keys). You can configure a `copy` section like this:

```toml
[default]
initialize = false
repository = "/backup/original"
password-file = "key"

    [default.copy]
    initialize = true
    repository = "/backup/copy"
    password-file = "other_key"
```

You will note that the secondary repository doesn't need to have a `2` behind its flags (`repository2`, `password-file2`, etc.). It's because the flags are well separated in the configuration.

Here's the same configuration in YAML format:

```yaml
default:
    initialize: false
    repository: "/backup/original"
    password-file: key
    copy:
        initialize: true
        repository: "/backup/copy"
        password-file: other_key
```
