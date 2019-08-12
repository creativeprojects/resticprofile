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

**resticprofile** needs python >=3.5 installed on your machine.

It's been actively tested on macOS X and Linux, and regularly tested on Windows.

**This is at _beta_ stage. Please don't use it in production yet. Even though I'm using it on my servers, I cannot guarantee all combinations of configuration are going to work properly for you.**

## Install

**resticprofile** is published on [Python Package Index](https://pypi.org/project/resticprofile/).
The easiest way to install resticprofile is using **pip**:
```
python3 -m pip install --user --upgrade resticprofile
```

## Configuration format

* A configuration is a set of _profiles_.
* Each profile is in its own `[section]`.
* Inside each profile, you can specify different flags for each command.
* A command definition is `[section.command]`.

All the restic flags can be defined in a section. For most of them you just need to remove the two dashes in front.

To set the flag `--password-file password.txt` you need to add a line like
```
password-file = "password.txt"
```

There's one **exception**: the flag `--repo` is named `repository` in the configuration

Let's say you normally use this command:

```
restic --repo "local:/backup" --password-file "password.txt" --verbose backup /home
```

For resticprofile to generate this command automatically for you, here's the configuration file:

```ini
[default]
repository = "local:/backup"
password-file = "password.txt"

[default.backup]
verbose = true
source = [ "/home" ]
```

You may have noticed the `source` flag is accepting an array of values (inside brakets)

Now, assuming this configuration file is named `profiles.conf` in the current folder, you can simply run

```
resticprofile backup
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

Simple example sending a file via stdin

```ini

[stdin]
repository = "local:/backup/restic"
password-file = "key"

[stdin.backup]
stdin = true
stdin-filename = "stdin-test"
tag = [ 'stdin' ]

```

## Using resticprofile

Here are a few examples how to run resticprofile (using the main example configuration file)

See all snapshots (assuming your default python is python3):

```
python -m resticprofile
```

Or if the folder `~/.local/bin/` is in your PATH, simply

```
resticprofile
```

Backup root & src profiles (using _full-backup_ group shown earlier)

```
python -m resticprofile --name "full-backup" backup
```

Assuming the _stdin_ profile from the configuration file shown before, the command to send a mysqldump to the backup is as simple as:

```
mysqldump --all-databases | python3 -m resticprofile -n stdin backup
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
   [--no-ansi]
   [-c|--config <configuration_file>]
   [-h|--help]
   [-n|--name <profile_name>]
   [-q|--quiet]
   [-v|--verbose]
   [restic command] [additional parameters to pass to restic]

```

## Command line reference ##

There are not many options on the command line, most of the options are in the configuration file.

* **[-h]**: Display quick help
* **[-c | --config] configuration_file**: Specify a configuration file other than the default
* **[-n | --name] profile_name**: Profile section to use from the configuration file
* **[-q | --quiet]**: Force resticprofile and restic to be quiet (override any configuration from the profile)
* **[-v | --verbose]**: Force resticprofile and restic to be verbose (override any configuration from the profile)
* **[--no-ansi]**: Disable console colouring (to save output into a log file)
* **[restic command]**: Like snapshots, backup, check, prune, forget, mount, etc.
* **[additional flags]**: Any additional flags to pass to the restic command line

## Configuration file reference

`[global]`

None of these flags are passed on the restic command line

* **ionice**: true / false
* **ionice-class**: integer
* **ionice-level**: integer
* **nice**: true / false OR integer
* **default-command**: string
* **initialize**: true / false
* **restic-binary**: string

`[profile]`

Flags used by resticprofile only

* ****inherit****: string
* **initialize**: true / false

Flags passed to the restic command line

* **cacert**: string
* **cache-dir**: string
* **cleanup-cache**: true / false
* **json**: true / false
* **key-hint**: string
* **limit-download**: integer
* **limit-upload**: integer
* **no-cache**: true / false
* **no-lock**: true / false
* **option**: string OR list of strings
* **password-command**: string
* **password-file**: string
* **quiet**: true / false
* **repository**: string **(will be passed as 'repo' to the command line)**
* **tls-client-cert**: string
* **verbose**: true / false OR integer

`[profile.backup]`

Flags used by resticprofile only

* **run-before**: string OR list of strings
* **run-after**: string OR list of strings
* **check-before**: true / false
* **check-after**: true / false

Flags passed to the restic command line

* **exclude**: string OR list of strings
* **exclude-caches**: true / false
* **exclude-file**: string OR list of strings
* **exclude-if-present**: string OR list of strings
* **files-from**: string OR list of strings
* **force**: true / false
* **host**: string
* **iexclude**: string OR list of strings
* **ignore-inode**: true / false
* **one-file-system**: true / false
* **parent**: string
* **stdin**: true / false
* **stdin-filename**: string
* **tag**: string OR list of strings
* **time**: string
* **with-atime**: true / false
* **source**: string OR list of strings

`[profile.retention]`

Flags used by resticprofile only

* **before-backup**: true / false
* **after-backup**: true / false

Flags passed to the restic command line

* **keep-last**: integer
* **keep-hourly**: integer
* **keep-daily**: integer
* **keep-weekly**: integer
* **keep-monthly**: integer
* **keep-yearly**: integer
* **keep-within**: string
* **keep-tag**: string OR list of strings
* **host**: true / false OR string
* **tag**: string OR list of strings
* **path**: string OR list of strings
* **compact**: true / false
* **group-by**: string
* **dry-run**: true / false
* **prune**: true / false

`[profile.snapshots]`

Flags passed to the restic command line

* **compact**: true / false
* **group-by**: string
* **host**: true / false OR string
* **last**: true / false
* **path**: string OR list of strings
* **tag**: string OR list of strings

`[profile.forget]`

Flags passed to the restic command line

* **keep-last**: integer
* **keep-hourly**: integer
* **keep-daily**: integer
* **keep-weekly**: integer
* **keep-monthly**: integer
* **keep-yearly**: integer
* **keep-within**: string
* **keep-tag**: string OR list of strings
* **host**: true / false OR string
* **tag**: string OR list of strings
* **path**: string OR list of strings
* **compact**: true / false
* **group-by**: string
* **dry-run**: true / false
* **prune**: true / false

`[profile.check]`

Flags passed to the restic command line

* **check-unused**: true / false
* **read-data**: true / false
* **read-data-subset**: string
* **with-cache**: true / false

`[profile.mount]`

Flags passed to the restic command line

* **allow-other**: true / false
* **allow-root**: true / false
* **host**: true / false OR string
* **no-default-permissions**: true / false
* **owner-root**: true / false
* **path**: string OR list of strings
* **snapshot-template**: string
* **tag**: string OR list of strings
