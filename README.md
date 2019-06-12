[![Build Status](https://travis-ci.com/creativeprojects/resticprofile.svg?branch=master)](https://travis-ci.com/creativeprojects/resticprofile)

# resticprofile
Configuration profiles for restic backup

This is not production ready, it's only a small script I'm making for my own backups.

Try [restic](https://restic.net/) and you'll understand why we need a configuration profile wrapper :)

Here's a configuration file example:

```ini
# Global configuration section
[global]
ionice = false # Linux only
ionice-class = 2
ionice-level = 6
nice = 10 # All unix-like
default-command = "snapshots" # when no command is specified when invoking resticprofile
initialize = false # initialize a repository if none exist at location

# Default profile when not specified (-n or --name)
[default]
repository = "/backup/default"
no-cache = true
initialize = false

# New profile named 'root'
[root]
repository = "/backup/root"
password-file = "key"
initialize = true

# 'backup' command of profile 'root'

[root.backup]
exclude-file = [ "root-excludes", "excludes" ]
exclude-caches = true
one-file-system = false
tag = [ "test", "dev" ]
backup = [ "." ]
prune-before = false
prune-after = true

# Environment variables of profile 'root'
[root.env]
EXAMPLE="some value"

# New profile named 'src'
[src]
repository = "/backup/sources"
password-file = "key"
initialize = true

# 'backup' command of profile 'src'
[src.backup]
exclude-file = [ "excludes" ]
tag = [ "src" ]
backup = [ "src" ]

# profile that represents a group: when called all the profiles will be running
[full-backup]
group = [ "root", "src" ]
prune-before = true
prune-after = false

```