# resticprofile
Configuration profiles for restic backup

This is not production ready, it's only a small script I'm making for my own backups.

Try [restic](https://restic.net/) and you'll understand why we need a configuration profile wrapper :)

Here's a configuration file example:

```ini
# This is the global configuration used by all profiles
[global]
ionice = false # Linux only
default-command = "snapshots" # when no command is specified when invoking resticprofile

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

```