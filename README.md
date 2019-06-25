[![Build Status](https://travis-ci.com/creativeprojects/resticprofile.svg?branch=master)](https://travis-ci.com/creativeprojects/resticprofile)

# resticprofile
Configuration profiles manager for [restic backup](https://restic.net/)

**resticprofile** is the missing link between a configuration file and restic backup. Creating a configuration file for restic has been [discussed before](https://github.com/restic/restic/issues/16), but seems to be a very low priority right now.

The configuration file is [TOML](https://github.com/toml-lang/toml) format:

* You no longer need to remember command parameters and environment variables
* You can create multiple profiles inside a configuration file
* A profile can inherit the options from another profile
* You can run the forget command before or after a backup (in a section called *retention*)
* You can check a repository before or after a backup
* You can create groups of profiles that will run sequentially
* You can run shell commands before or after a backup
* Allows to start the restic process using _nice_ (not available on Windows) and/or _ionice_ (only available on Linux)

## Requirements

**resticprofile** needs python 3 (tested with version 3.5 minimum) installed on your machine

It's been actively tested on macOs X and Linux, and regularly tested on Windows.

**This is at _beta_ stage. Please don't use it in production yet. Even though I'm using it on my servers, I cannot guarantee all combinations of configuration are going to work properly for you.**

## Install

The simpliest way to install resticprofile for now is via **pip**:
```
python3 -m pip install --user --upgrade resticprofile
```

## Configuration examples

Here's a simple configuration file using a Microsoft Azure backend:

```ini
[default]
repository = "azure:restic:/"
password-file = "key"

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

Here's a more complex configuration file showing profile inheritance and two backup profiles using the same repository:

```ini
[global]
# ionice is available on Linux only
ionice = false
ionice-class = 2
ionice-level = 6
# nice is available on all unixes (macOs X included)
nice = 10
# run 'snapshots' when no command is specified when invoking resticprofile
default-command = "snapshots"
# initialize a repository if none exist at location
initialize = false

# a group is a profile that will call all profiles one by one
[groups]
# when starting a backup on profile "full-backup", it will run the "root" and "src" backup profiles
full-backup = [ "root", "src" ]

# Default profile when not specified (-n or --name)
# Please note there's no default inheritance from the 'default' profile (you can use the 'inherit' flag if needed)
[default]
repository = "/backup"
password-file = "key"
initialize = false

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

# 'backup' command of profile 'root'
[root.backup]
exclude-file = [ "root-excludes", "excludes" ]
exclude-caches = true
one-file-system = false
tag = [ "test", "dev" ]
source = [ "." ]

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
# if path is NOT specified, it will be copied from the 'backup' source
# path = []
# the tags are NOT copied from the 'backup' command
tag = [ "test", "dev" ]
# host can be a boolean ('true' meaning current hostname) or a string to specify a different hostname
host = true

# New profile named 'src'
[src]
inherit = "default"
initialize = true

# 'backup' command of profile 'src'
[src.backup]
run-before = [ "echo Starting!", "ls -al ./src" ]
run-after = "echo All Done!"
exclude = [ '/**/.git' ]
exclude-caches = true
one-file-system = false
tag = [ "test", "dev" ]
source = [ "./src" ]
check-before = true

# retention policy for profile src
[src.retention]
before-backup = false
after-backup = true
keep-within = "30d"
compact = false
prune = true

```

And another simple example for Windows:

```ini
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

```

## Using resticprofile

Here are a few examples how to run resticprofile (using the main example configuration file)

See all snapshots:

```
python -m resticprofile
```

Backup root & src profiles (using full-backup group)

```
python -m resticprofile --name "full-backup" backup
```

Mount the default profile (_default_) in /mnt/restic:

```
python -m resticprofile mount /mnt/restic
```

Display quick help

```
python -m resticprofile --help

Usage:
 resticprofile
   [-c|--config <configuration_file>]
   [-h|--help]
   [-n|--name <profile_name>]
   [-q|--quiet]
   [-v|--verbose]
   [restic command] [additional parameters to pass to restic]

```
