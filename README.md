[![Build Status](https://travis-ci.com/creativeprojects/resticprofile.svg?branch=master)](https://travis-ci.com/creativeprojects/resticprofile)
[![Go Report Card](https://goreportcard.com/badge/github.com/creativeprojects/resticprofile)](https://goreportcard.com/report/github.com/creativeprojects/resticprofile)

# resticprofile
Configuration profiles manager for [restic backup](https://restic.net/)

**resticprofile** is the missing link between a configuration file and restic backup. Creating a configuration file for restic has been [discussed before](https://github.com/restic/restic/issues/16), but seems to be a very low priority right now.

With resticprofile:

* You no longer need to remember command parameters and environment variables
* You can create multiple profiles inside one configuration file
* A profile can inherit all the options from another profile
* You can run the forget command before or after a backup (in a section called *retention*)
* You can check a repository before or after a backup
* You can create groups of profiles that will run sequentially
* You can run shell commands before or after a backup
* You can also run shell commands before or after running a profile (any command): useful if you need to mount your backup disk
* You can send a backup stream via _stdin_
* You can start restic at a lower or higher priority (Priority Class in Windows, *nice* in all unixes) and/or _ionice_ (only available on Linux)

The configuration file accepts various formats:
* [TOML](https://github.com/toml-lang/toml) : configuration file with extension _.toml_ and _.conf_ to keep compatibility with versions before 0.6.0
* [JSON](https://en.wikipedia.org/wiki/JSON) : configuration file with extension _.json_
* [YAML](https://en.wikipedia.org/wiki/YAML) : configuration file with extension _.yaml_

For the rest of the documentation, I'll be showing examples using the TOML file configuration format (because it was the only one supported before version 0.6.0) but you can pick your favourite: they all work with resticprofile :-)

## Requirements

Since version 0.6.0, **resticprofile** **no longer needs** python installed on your machine. It is distributed as an executable (same as restic).

It's been actively tested on macOS X and Linux, and regularly tested on Windows.

**This is at _beta_ stage. Please don't use it in production yet. Even though I'm using it on my servers, I cannot guarantee all combinations of configuration are going to work properly for you.**

## Installation

Here's a simple script to download the binary automatically for you. It works on mac OS X, FreeBSD, OpenBSD and Linux:

```
$ curl -sfL https://raw.githubusercontent.com/creativeprojects/resticprofile/master/install.sh | sh
```

It should copy resticprofile in a `bin` directory under your current directory.

If you need more control, you can save the shell script and run it manually:

```
$ curl -LO https://raw.githubusercontent.com/creativeprojects/resticprofile/master/install.sh
$ chmod +x install.sh
$ sudo ./install.sh -b /usr/local/bin
```

It will install resticprofile in `/usr/local/bin/`


### Manual installation (Windows)

- Download the package corresponding to your system and CPU from the [release page](https://github.com/creativeprojects/resticprofile/releases)
- Once downloaded you need to open the archive and copy the binary file `resticprofile` (or `resticprofile.exe`) in your PATH.

## Upgrade

Once installed, you can easily upgrade resticprofile to the latest release using this flag:

```
$ resticprofile --self-update
```

Starting with version 0.7.0 you can also upgrade resticprofile using this command:

```
$ resticprofile self-update
```

## Using docker image ##

You can run resticprofile inside a docker container. It is probably the easiest way to install resticprofile (and restic at the same time) and keep it updated.

**But** be aware that you will need to mount your backup source (and destination if it's local) as a docker volume.
Depending on your operating system, the backup might be **slower**. Volumes mounted on a mac OS host are well known for being quite slow.

By default, the resticprofile container starts at `/resticprofile`. So you can feed a configuration this way:

```
$ docker run -it --rm -v $PWD/examples:/resticprofile creativeprojects/resticprofile
```

You can list your profiles:
```
$ docker run -it --rm -v $PWD/examples:/resticprofile creativeprojects/resticprofile profiles
```

### Please note:

Each time a container is started, it gets assigned a new random name. You should probably force a hostname to your container...

```
$ docker run -it --rm -v $PWD:/resticprofile -h my-machine creativeprojects/resticprofile -n profile backup
```

... or in your configuration:

```ini
[profile]
host = "my-machine"
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

There's **one exception**: the flag `--repo` is named `repository` in the configuration

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
# priority is using priority class on windows, and "nice" on unixes - it's acting on CPU usage only
priority = "low"
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
# you can use a relative path, it will be relative to the configuration file
repository = "/backup"
password-file = "key"
initialize = false
# will run these scripts before and after each command (including 'backup')
run-before = "mount /backup"
run-after = "umount /backup"

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

# 'backup' command of profile 'root'
[root.backup]
# files with no path are relative to the configuration file
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
exclude = [ '/**/.git' ]
exclude-caches = true
one-file-system = false
tag = [ "test", "dev" ]
source = [ "./src" ]
check-before = true
# will only run these scripts before and after a backup
run-before = [ "echo Starting!", "ls -al ./src" ]
run-after = "echo All Done!"

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

See all snapshots of your `[default]` profile:

```
$ resticprofile
```

See all available profiles in your configuration file (and the restic commands where some flags are defined):

```
$ resticprofile profiles

Profiles available:
  stdin:     (backup)
  default:   (env)
  root:      (retention, backup)
  src:       (retention, backup)
  linux:     (retention, backup, snapshots, env)
  no-cache:  (n/a)

Groups available:
  full-backup:  root, src

```

Backup root & src profiles (using _full-backup_ group shown earlier)

```
$ resticprofile --name "full-backup" backup
```

Assuming the _stdin_ profile from the configuration file shown before, the command to send a mysqldump to the backup is as simple as:

```
$ mysqldump --all-databases | resticprofile --name stdin backup
```

Mount the default profile (_default_) in /mnt/restic:

```
$ resticprofile mount /mnt/restic
```

Display quick help

```
$ resticprofile --help

Usage of resticprofile:
	resticprofile [resticprofile flags] [command] [restic flags]

resticprofile flags:
  -c, --config string   configuration file (default "profiles.conf")
  -h, --help            display this help
  -n, --name string     profile name (default "default")
      --no-ansi         disable ansi control characters (disable console colouring)
  -q, --quiet           display only warnings and errors
      --theme string    console colouring theme (dark, light, none) (default "light")
  -v, --verbose         display all debugging information

resticprofile own commands:
   profiles       display profile names from the configuration file
   self-update    update resticprofile to latest version (does not update restic)
   systemd-unit   create a user systemd timer

```

A command is a restic command **except** for one command recognized by resticprofile only: `profiles`


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

`global` is a fixed name

None of these flags are passed on the restic command line

* **ionice**: true / false
* **ionice-class**: integer
* **ionice-level**: integer
* **nice**: true / false OR integer
* **priority**: string = `Idle`, `Background`, `Low`, `Normal`, `High`, `Highest`
* **default-command**: string
* **initialize**: true / false
* **restic-binary**: string

`[profile]`

`profile` is the name of your profile

Flags used by resticprofile only

* ****inherit****: string
* **initialize**: true / false
* **lock**: string: specify a local lockfile
* **run-before**: string OR list of strings
* **run-after**: string OR list of strings

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
* **host**: true / false OR string
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

## Appendix

As an example, here's a similar configuration file in YAML:

```yaml
global:
    default-command: version
    initialize: false
    priority: low

groups:
    full-backup:
    - root
    - src

default:
    env:
        tmp: /tmp
    initialize: false
    password-file: key
    repository: /backup

documents:
    backup:
        source: ~/Documents
    initialize: false
    repository: ~/backup
    snapshots:
        tag:
        - documents

root:
    backup:
        exclude-caches: true
        exclude-file:
        - root-excludes
        - excludes
        one-file-system: false
        source:
        - .
        tag:
        - test
        - dev
    inherit: default
    initialize: true
    retention:
        after-backup: true
        before-backup: false
        compact: false
        host: true
        keep-daily: 1
        keep-hourly: 1
        keep-last: 3
        keep-monthly: 1
        keep-tag:
        - forever
        keep-weekly: 1
        keep-within: 3h
        keep-yearly: 1
        prune: false
        tag:
        - test
        - dev

self:
    backup:
        source: ./
    initialize: false
    repository: ../backup
    snapshots:
        tag:
        - self

src:
    lock: "/tmp/resticprofile-profile-src.lock"
    backup:
        check-before: true
        exclude:
        - /**/.git
        exclude-caches: true
        one-file-system: false
        run-after: echo All Done!
        run-before:
        - echo Starting!
        - ls -al ~/go
        source:
        - ~/go
        tag:
        - test
        - dev
    inherit: default
    initialize: true
    retention:
        after-backup: true
        before-backup: false
        compact: false
        keep-within: 30d
        prune: true
    snapshots:
        tag:
        - test
        - dev
        
stdin:
    backup:
        stdin: true
        stdin-filename: stdin-test
        tag:
        - stdin
    inherit: default
    snapshots:
        tag:
        - stdin

```

## Using resticprofile and systemd

systemd is a common service manager in use by many Linux distributions.
resticprofile has the ability to autocreate systemd timer and service files.
systemd can be used in place of cron to schedule backups.

All systemd units are created under the user's systemd profile (~/.config/systemd/user).

TODO: create system profiles

### systemd calendars

resticprofile uses systemd
[OnCalendar](https://www.freedesktop.org/software/systemd/man/systemd.time.html#Calendar%20Events)
format to schedule events.

Testing systemd calendars can be done with the systemd-analyze application.
systemd-analyze will display when the next trigger will happen:

```
$ systemd-analyze calendar 'daily'
  Original form: daily
Normalized form: *-*-* 00:00:00
    Next elapse: Sat 2020-04-18 00:00:00 CDT
       (in UTC): Sat 2020-04-18 05:00:00 UTC
       From now: 10h left
```

### Configuring a systemd profile

Running the following command will create a timer and systemd unit for the
'configs' profile name within resticprofile. 

```
$ resticprofile -n configs systemd-unit daily
2020/04/17 13:34:07 resticprofile 0.6.0 compiled with go1.14.2
2020/04/17 13:34:07 Writing /home/<user>/.config/systemd/user/resticprofile-backup@configs.service
2020/04/17 13:34:07 Writing /home/<user>/.config/systemd/user/resticprofile-backup@configs.timer
```

The service can be tested or run once with:

```
$ systemctl --user start resticprofile-backup@configs.service
```

Or, starting the timer will enable the schedule:
```
$ systemctl --user start resticprofile-backup@configs.timer
```

To persist the timer across reboots, replace `start` with enable:

```
$ systemctl --user enable resticprofile-backup@configs.timer
```

