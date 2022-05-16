![Build](https://github.com/creativeprojects/resticprofile/workflows/Build/badge.svg)
[![codecov](https://codecov.io/gh/creativeprojects/resticprofile/branch/master/graph/badge.svg?token=cUozgF9j4I)](https://codecov.io/gh/creativeprojects/resticprofile)
[![Go Report Card](https://goreportcard.com/badge/github.com/creativeprojects/resticprofile)](https://goreportcard.com/report/github.com/creativeprojects/resticprofile)
[![Lines of Code](https://sonarcloud.io/api/project_badges/measure?project=creativeprojects_resticprofile&metric=ncloc)](https://sonarcloud.io/dashboard?id=creativeprojects_resticprofile)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=creativeprojects_resticprofile&metric=bugs)](https://sonarcloud.io/dashboard?id=creativeprojects_resticprofile)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=creativeprojects_resticprofile&metric=vulnerabilities)](https://sonarcloud.io/dashboard?id=creativeprojects_resticprofile)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=creativeprojects_resticprofile&metric=sqale_rating)](https://sonarcloud.io/dashboard?id=creativeprojects_resticprofile)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=creativeprojects_resticprofile&metric=reliability_rating)](https://sonarcloud.io/dashboard?id=creativeprojects_resticprofile)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=creativeprojects_resticprofile&metric=security_rating)](https://sonarcloud.io/dashboard?id=creativeprojects_resticprofile)

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
* You can run shell commands before or after running a profile: useful if you need to mount and unmount your backup disk for example
* You can run a shell command if an error occurred (at any time)
* You can send a backup stream via _stdin_
* You can start restic at a lower or higher priority (Priority Class in Windows, *nice* in all unixes) and/or _ionice_ (only available on Linux)
* It can check that you have enough memory before starting a backup. (I've had some backups that literally killed a server with swap disabled)
* You can generate cryptographically secure random keys to use as a restic key file
* You can easily schedule backups, retentions and checks (works for *systemd*, *crond*, *launchd* and *windows task scheduler*)
* You can generate a simple status file to send to some monitoring software and make sure your backups are running fine 
* You can use a template syntax in your configuration file
* You can generate scheduled tasks using *crond*
* Get backup statistics in your status file
* **[new for v0.14.0]** Automatically clear up [stale locks](#locks)
* **[new for v0.15.0]** Export a **prometheus** file after a backup, or send the report to a push gateway automatically
* **[new for v0.16.0]** Full support for the copy command (with scheduling)
* **[new for v0.16.0]** Describe your own systemd units and timers with go templates

The configuration file accepts various formats:
* [TOML](https://github.com/toml-lang/toml) : configuration file with extension _.toml_ and _.conf_ to keep compatibility with versions before 0.6.0
* [JSON](https://en.wikipedia.org/wiki/JSON) : configuration file with extension _.json_
* [YAML](https://en.wikipedia.org/wiki/YAML) : configuration file with extension _.yaml_
* [HCL](https://github.com/hashicorp/hcl): configuration file with extension _.hcl_

For the rest of the documentation, I'll be showing examples using different formats, but mostly TOML and YAML.

# Table of Contents

This README is getting too big now, so I'm in the process of moving most of its content into a proper documentation.

**This is work in progress** but you can still [check the documentation out here](https://creativeprojects.github.io/resticprofile/).

<!--ts-->
* [resticprofile](#resticprofile)
* [Table of Contents](#table-of-contents)
* [Requirements](#requirements)
* [Installation](#installation)
  * [Homebrew (all macOS, Linux on amd64)](#homebrew-all-macos-linux-on-amd64)
    * [Note on installing on Linux via Homebrew](#note-on-installing-on-linux-via-homebrew)
  * [Via a script (macOS, Linux &amp; other unixes)](#via-a-script-macos-linux--other-unixes)
  * [Installation for Windows using bash](#installation-for-windows-using-bash)
  * [Manual installation (Windows)](#manual-installation-windows)
  * [Ansible](#ansible)
  * [Installation from source](#installation-from-source)
  * [Shell completion](#shell-completion)
* [Upgrade](#upgrade)
* [Using docker image](#using-docker-image)
  * [Container host name](#container-host-name)
* [Configuration format](#configuration-format)
* [Configuration examples](#configuration-examples)
  * [Simple TOML configuration](#simple-toml-configuration)
  * [Simple YAML configuration](#simple-yaml-configuration)
  * [More complex configuration in TOML](#more-complex-configuration-in-toml)
  * [TOML configuration example for Windows](#toml-configuration-example-for-windows)
  * [Use stdin in configuration](#use-stdin-in-configuration)
  * [Special case for the copy command section](#special-case-for-the-copy-command-section)
* [Configuration paths](#configuration-paths)
  * [macOS X](#macos-x)
  * [Other unixes (Linux and BSD)](#other-unixes-linux-and-bsd)
  * [Windows](#windows)
* [Includes](#includes)
* [Path resolution in configuration](#path-resolution-in-configuration)
* [Run commands before, after success or after failure](#run-commands-before-after-success-or-after-failure)
  * [order of run\-\* during a backup](#order-of-run--during-a-backup)
* [Warnings from restic](#warnings-from-restic)
  * [no\-error\-on\-warning](#no-error-on-warning)
* [Locks](#locks)
  * [Stale locks](#stale-locks)
  * [Lock wait](#lock-wait)
  * [restic lock management](#restic-lock-management)
* [Using resticprofile](#using-resticprofile)
* [Command line reference](#command-line-reference)
* [Minimum memory required](#minimum-memory-required)
* [Version](#version)
* [Generating random keys](#generating-random-keys)
* [Scheduled backups](#scheduled-backups)
  * [retention schedule is deprecated](#retention-schedule-is-deprecated)
  * [Schedule configuration](#schedule-configuration)
    * [schedule\-permission](#schedule-permission)
    * [schedule\-lock\-mode](#schedule-lock-mode)
    * [schedule\-lock\-wait](#schedule-lock-wait)
    * [schedule\-log](#schedule-log)
    * [schedule\-priority (systemd and launchd only)](#schedule-priority-systemd-and-launchd-only)
    * [schedule](#schedule)
  * [Scheduling commands](#scheduling-commands)
    * [schedule command](#schedule-command)
    * [unschedule command](#unschedule-command)
    * [status command](#status-command)
    * [Examples of scheduling commands under Windows](#examples-of-scheduling-commands-under-windows)
    * [Examples of scheduling commands under Linux](#examples-of-scheduling-commands-under-linux)
    * [Examples of scheduling commands under macOS](#examples-of-scheduling-commands-under-macos)
  * [Changing schedule\-permission from user to system, or system to user](#changing-schedule-permission-from-user-to-system-or-system-to-user)
* [Status file for easy monitoring](#status-file-for-easy-monitoring)
  * [Extended status](#extended-status)
* [Prometheus](#prometheus)
  * [User defined labels](#user-defined-labels)
* [Variable expansion in configuration file](#variable-expansion-in-configuration-file)
  * [Pre\-defined variables](#pre-defined-variables)
  * [Hand\-made variables](#hand-made-variables)
  * [Examples](#examples)
* [Configuration templates](#configuration-templates)
* [Debugging your template and variable expansion](#debugging-your-template-and-variable-expansion)
* [Limitations of using templates](#limitations-of-using-templates)
* [Documentation on template, variable expansion and other configuration scripting](#documentation-on-template-variable-expansion-and-other-configuration-scripting)
* [Configuration file reference](#configuration-file-reference)
* [Appendix](#appendix)
* [Using resticprofile and systemd](#using-resticprofile-and-systemd)
  * [systemd calendars](#systemd-calendars)
  * [First time schedule](#first-time-schedule)
  * [How to change the default systemd unit and timer file using a template](#how-to-change-the-default-systemd-unit-and-timer-file-using-a-template)
    * [Default unit file](#default-unit-file)
    * [Default timer file](#default-timer-file)
    * [Template variables](#template-variables)
* [Using resticprofile and launchd on macOS](#using-resticprofile-and-launchd-on-macos)
  * [User agent](#user-agent)
    * [Special case of schedule\-permission=user with sudo](#special-case-of-schedule-permissionuser-with-sudo)
  * [Daemon](#daemon)
* [Contributions](#contributions)

<!--te-->



# Run commands before, after success or after failure

resticprofile has 2 places where you can run commands around restic:

- commands that will run before and after every restic command (snapshots, backup, check, forget, prune, mount, etc.). These are placed at the root of each profile.
- commands that will only run before and after a backup: these are placed in the backup section of your profiles.

Here's an example of all the external commands that you can run during the execution of a profile:

```yaml
documents:
  inherit: default
  run-before: "echo == run-before profile $PROFILE_NAME command $PROFILE_COMMAND"
  run-after: "echo == run-after profile $PROFILE_NAME command $PROFILE_COMMAND"
  run-after-fail: "echo == Error in profile $PROFILE_NAME command $PROFILE_COMMAND: $ERROR"
  run-finally: "echo == run-finally $PROFILE_NAME command $PROFILE_COMMAND"
  backup:
    run-before: "echo === run-before backup profile $PROFILE_NAME command $PROFILE_COMMAND"
    run-after: "echo === run-after backup profile $PROFILE_NAME command $PROFILE_COMMAND"
    run-finally: "echo == run-finally $PROFILE_NAME command $PROFILE_COMMAND"
    source: ~/Documents
```

`run-before`, `run-after`, `run-after-fail` and `run-finally` can be a string, or an array of strings if you need to run more than one command

A few environment variables will be set before running these commands:
- `PROFILE_NAME`
- `PROFILE_COMMAND`: backup, check, forget, etc.

Additionally, for the `run-after-fail` commands, these environment variables will also be available:
- `ERROR` containing the latest error message
- `ERROR_COMMANDLINE` containing the command line that failed
- `ERROR_EXIT_CODE` containing the exit code of the command line that failed
- `ERROR_STDERR` containing any message that the failed command sent to the standard error (stderr)

The commands of `run-finally` get the environment of `run-after-fail` when `run-before`, `run-after` or `restic` failed. 
Failures in `run-finally` are logged but do not influence environment or return code.

## order of `run-*` during a backup

The commands will be running in this order **during a backup**:
- `run-before` from the profile - if error, go to `run-after-fail`
- `run-before` from the backup section - if error, go to `run-after-fail`
- run the restic backup (with check and retention if configured) - if error, go to `run-after-fail`
- `run-after` from the backup section - if error, go to `run-after-fail`
- `run-after` from the profile - if error, go to `run-after-fail`
- If error: `run-after-fail` from the profile - if error, go to `run-finally`
- `run-finally` from the backup section - if error, log and continue with next
- `run-finally` from the profile - if error, log and continue with next

Maybe it's easier to understand with a flow diagram:

![run flow diagram](https://github.com/creativeprojects/resticprofile/raw/master/run-flow.svg)

# Warnings from restic

Until version 0.13.0, resticprofile was always considering a restic warning as an error. This will remain the **default**.
But the version 0.13.0 introduced a parameter to avoid this behaviour and consider that the backup was successful instead.

A restic warning occurs when it cannot read some files, but a snapshot was successfully created.

## no-error-on-warning

```yaml
profile:
    inherit: default
    backup:
        no-error-on-warning: true
```

# Locks

restic is already using a lock to avoid running some operations at the same time.

Since resticprofile can run several commands in a profile, it could be better to run the whole batch in a lock so nobody can interfere in the meantime.

For this to happen you can specify a lock file in each profile:

```yaml
src:
    lock: "/tmp/resticprofile-profile-src.lock"
    backup:
        check-before: true
        exclude:
        - /**/.git
        source:
        - ~/go
    retention:
        after-backup: true
        before-backup: false
        compact: false
        keep-within: 30d
        prune: true
```

For this profile, a lock will be set using the file `/tmp/resticprofile-profile-src.lock` for the duration of the profile: *check*, *backup* and *retention* (via the forget command)

**Please note restic locks and resticprofile locks are completely independent**

## Stale locks

In some cases, resticprofile as well as restic may leave a lock behind if the process died (or the machine rebooted).

For that matter, if you add the flag `force-inactive-lock` to your profile, resticprofile will detect and remove stale locks: 
* **resticprofile locks**: Check for the presence of a process with the PID indicated in the lockfile. If it can't find any, it will try to delete the lock and continue the operation (locking again, running profile and so on...)
* **restic locks**: Evaluate if a restic command failed on acquiring a lock. If the lock is older than `restic-stale-lock-age`, invoke `restic unlock` and retry the command that failed (can be disabled by setting `restic-stale-lock-age` to 0, default is 2h).

```yaml
global:
  restic-stale-lock-age: 2h

src:
    lock: "/tmp/resticprofile-profile-src.lock"
    force-inactive-lock: true
```

## Lock wait

By default, restic and resticprofile fail when a lock cannot be acquired as another process is currently holding it.

Depending on the use case (e.g. scheduled backups), it may be more appropriate to wait on another process to finish instead of failing immediately.

For that matter, if you add the commandline flag `--lock-wait` or configure schedules with `schedule-lock-wait`, resticprofile will wait on other backup processes:
* **resticprofile locks**: Retry acquiring the lockfile until it either succeeds (when the other resticprofile process released the lock) or fail as the lock-wait duration has passed without success.
* **restic locks**: Evaluate if a restic command failed on acquiring a lock. If the lock is not considered stale, retry the restic command every `restic-lock-retry-after` (default 1 minute) until it acquired the lock, or fail as the lock-wait duration has passed.

Note: The lock wait duration is cumulative. If various locks in one profile-run require lock wait, the total wait time may not exceed the duration that was specified. 

## restic lock management

resticprofile can retry restic commands that fail on acquiring a lock and can also ask restic to unlock stale locks. The behaviour is controlled by 2 settings inside the `global` section:

```yaml
global:
  # Retry a restic command that failed on acquiring a lock every minute 
  # (at least), for up to the time specified in "--lock-wait duration". 
  restic-lock-retry-after: 1m
  # Ask restic to unlock a stale lock when its age is more than 2 hours
  # and the option "force-inactive-lock" is enabled in the profile.
  restic-stale-lock-age: 2h
```

If restic lock management is not desired, it can be disabled by setting both values to 0.

# Using resticprofile

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
or
```
$ resticprofile full-backup.backup
```

Assuming the _stdin_ profile from the configuration file shown before, the command to send a mysqldump to the backup is as simple as:

```
$ mysqldump --all-databases --order-by-primary | resticprofile --name stdin backup
```
or
```
$ mysqldump --all-databases --order-by-primary | resticprofile stdin.backup
```
or when resticprofile runs "mysqldump" (can be scheduled):
```
$ resticprofile mysql.backup
```

Mount the default profile (_default_) in /mnt/restic:

```
$ resticprofile mount /mnt/restic
```

Display quick help

```
$ resticprofile --help

Usage of resticprofile:
	resticprofile [resticprofile flags] [profile name.][restic command] [restic flags]
	resticprofile [resticprofile flags] [profile name.][resticprofile command] [command specific flags]

resticprofile flags:
  -c, --config string        configuration file (default "profiles")
      --dry-run              display the restic commands instead of running them
  -f, --format string        file format of the configuration (default is to use the file extension)
  -h, --help                 display this help
      --lock-wait duration   wait up to duration to acquire a lock (syntax "1h5m30s")
  -l, --log string           logs into a file instead of the console
  -n, --name string          profile name (default "default")
      --no-ansi              disable ansi control characters (disable console colouring)
      --no-lock              skip profile lock file
      --no-prio              don't set any priority on load: used when started from a service that has already set the priority
  -q, --quiet                display only warnings and errors
      --theme string         console colouring theme (dark, light, none) (default "light")
      --trace                display even more debugging information
  -v, --verbose              display some debugging information
  -w, --wait                 wait at the end until the user presses the enter key

resticprofile own commands:
   version       display version (run in verbose mode for detailed information)
   self-update   update to latest resticprofile (use -q/--quiet flag to update without confirmation)
   profiles      display profile names from the configuration file
   show          show all the details of the current profile
   random-key    generate a cryptographically secure random key to use as a restic keyfile
   schedule      schedule jobs from a profile (use --all flag to schedule all jobs of all profiles)
   unschedule    remove scheduled jobs of a profile (use --all flag to unschedule all profiles)
   status        display the status of scheduled jobs (use --all flag for all profiles)


```

A command is either a restic command or a resticprofile own command.


# Command line reference ##

There are not many options on the command line, most of the options are in the configuration file.

* **[-h]**: Display quick help
* **[-c | --config] configuration_file**: Specify a configuration file other than the default
* **[-f | --format] configuration_format**: Specify the configuration file format: `toml`, `yaml`, `json` or `hcl`
* **[-n | --name] profile_name**: Profile section to use from the configuration file.
  You can also use `[profile_name].[command]` syntax instead, this will only work if `-n` is not set.
  Using `-n [profile_name] [command]` or `[profile_name].[command]` both select profile and command and are technically equivalent.
* **[--dry-run]**: Doesn't run the restic command but display the command line instead
* **[-q | --quiet]**: Force resticprofile and restic to be quiet (override any configuration from the profile)
* **[-v | --verbose]**: Force resticprofile and restic to be verbose (override any configuration from the profile)
* **[--no-ansi]**: Disable console colouring (to save output into a log file)
* **[--no-lock]**: Disable resticprofile locks, neither create nor fail on a lock. restic locks are unaffected by this option.
* **[--theme]**: Can be `light`, `dark` or `none`. The colours will adjust to a 
light or dark terminal (none to disable colouring)
* **[--lock-wait] duration**: Retry to acquire resticprofile and restic locks for up to the specified amount of time before failing on a lock failure. 
* **[-l | --log] log_file**: To write the logs in file instead of displaying on the console
* **[-w | --wait]**: Wait at the very end of the execution for the user to press enter. This is only useful in Windows when resticprofile is started from explorer and the console window closes automatically at the end.
* **[resticprofile OR restic command]**: Like snapshots, backup, check, prune, forget, mount, etc.
* **[additional flags]**: Any additional flags to pass to the restic command line

# Minimum memory required

restic can be memory hungry. I'm running a few servers with no swap (I know: it is _bad_) and I managed to kill some of them during a backup.
For that matter I've introduced a parameter in the `global` section called `min-memory`. The **default value is 100MB**. You can disable it by using a value of `0`.

It compares against `(total - used)` which is probably the best way to know how much memory is available (that is including the memory used for disk buffers/cache).

# Version

The `version` command displays resticprofile version. If run in vebose mode (using `--verbose` flag) additional information such as OS version or golang version or modules are displayed as well.

```
$ resticprofile --verbose version
```

# Generating random keys

resticprofile has a handy tool to generate cryptographically secure random keys encoded in base64. You can simply put this key into a file and use it as a strong key for restic

On Linux and FreeBSD, the generator uses getrandom(2) if available, /dev/urandom otherwise. On OpenBSD, the generator uses getentropy(2). On other Unix-like systems, the generator reads from /dev/urandom. On Windows systems, the generator uses the CryptGenRandom API. On Wasm, the generator uses the Web Crypto API. 
[Reference from the Go documentation](https://golang.org/pkg/crypto/rand/#pkg-variables)

```
$ resticprofile random-key
```

generates a 1024 bytes random key (converted into 1368 base64 characters) and displays it on the console

To generate a different size of key, you can specify the bytes length on the command line:

```
$ resticprofile random-key 2048
```

# Scheduled backups

resticprofile is capable of managing scheduled backups for you using:
- **launchd** on macOS X
- **Task Scheduler** on Windows
- **systemd** where available (Linux and other BSDs)
- **crond** on supported platforms (Linux and other BSDs)

On unixes (except macOS) resticprofile is using **systemd** by default. **crond** can be used instead if configured in `global` `scheduler` parameter:

```yaml
---
global:
    scheduler: crond
```



Each profile can be scheduled independently (groups are not available for scheduling yet).

These 5 profile sections are accepting a schedule configuration:
- backup
- check
- forget (version 0.11.0)
- prune (version 0.11.0)
- copy (version 0.16.0)

which mean you can schedule `backup`, `forget`, `prune`, `check` and `copy` independently (I recommend to use a local `lock` in this case).

## retention schedule is deprecated
**Important**:
starting from version 0.11.0 the schedule of the `retention` section is **deprecated**: Use the `forget` section instead.


## Schedule configuration

The schedule configuration consists of a few parameters which can be added on each profile:

```toml
[profile.backup]
schedule = "*:00,30"
schedule-permission = "system"
schedule-priority = "background"
schedule-log = "profile-backup.log"
schedule-lock-mode = "default"
schedule-lock-wait = "15m30s"
```



### schedule-permission

`schedule-permission` accepts two parameters: `user` or `system`:

* `user`: your backup will be running using your current user permissions on files. This is fine if you're only saving your documents (or any other file inside your profile). Please note on **systemd** that the schedule **will only run when your user is logged in**.

* `system`: if you need to access some system or protected files. You will need to run resticprofile with `sudo` on unixes and with elevated prompt on Windows (please note on Windows resticprofile will ask you for elevated permissions automatically if needed).

* *empty*: resticprofile will try its best guess based on how you started it (with sudo or as a normal user) and fallback to `user`

### schedule-lock-mode

Starting from version 0.14.0, `schedule-lock-mode` accepts 3 values:
- `default`: Wait on acquiring a lock for the time duration set in `schedule-lock-wait`, before failing a schedule.
   Behaves like `fail` when `schedule-lock-wait` is "0" or not specified.
- `fail`: Any lock failure causes a schedule to abort immediately. 
- `ignore`: Skip resticprofile locks. restic locks are not skipped and can abort the schedule.

### schedule-lock-wait

Sets the amount of time to wait for a resticprofile and restic lock to become available. Is only used when `schedule-lock-mode` is unset or `default`.

### schedule-log

Allow to redirect all output from resticprofile and restic to a file

### schedule-priority (systemd and launchd only)

Starting from version 0.11.0, `schedule-priority` accepts two values:
- `background`: the process shouldn't be noticeable when working on the machine at the same time (this is the default)
- `standard`: the process should get the same priority as any other process on the machine (but it won't run faster if you're not using the machine at the same time)

`schedule-priority` is not available for windows task scheduler, nor crond

### schedule

The `schedule` parameter accepts many forms of input from the [systemd calendar event](https://www.freedesktop.org/software/systemd/man/systemd.time.html#Calendar%20Events) type. This is by far the easiest to use: **It is the same format used to schedule on macOS and Windows**.

The most general form is:
```
weekdays year-month-day hour:minute:second
```

- use `*` to mean any
- use `,` to separate multiple entries
- use `..` for a range

**limitations**:
- the divider (`/`), the `~` and timezones are not (yet?) supported on macOS and Windows.
- the `year` and `second` fields have no effect on macOS. They do have limited availability on Windows (they don't make much sense anyway).

Here are a few examples (taken from the systemd documentation):

```
On the left is the user input, on the right is the full format understood by the system

  Sat,Thu,Mon..Wed,Sat..Sun → Mon..Thu,Sat,Sun *-*-* 00:00:00
      Mon,Sun 12-*-* 2,1:23 → Mon,Sun 2012-*-* 01,02:23:00
                    Wed *-1 → Wed *-*-01 00:00:00
           Wed..Wed,Wed *-1 → Wed *-*-01 00:00:00
                 Wed, 17:48 → Wed *-*-* 17:48:00
Wed..Sat,Tue 12-10-15 1:2:3 → Tue..Sat 2012-10-15 01:02:03
                *-*-7 0:0:0 → *-*-07 00:00:00
                      10-15 → *-10-15 00:00:00
        monday *-12-* 17:00 → Mon *-12-* 17:00:00
     Mon,Fri *-*-3,1,2 *:30 → Mon,Fri *-*-01,02,03 *:30:00
       12,14,13,12:20,10,30 → *-*-* 12,13,14:10,20,30:00
            12..14:10,20,30 → *-*-* 12..14:10,20,30:00
                03-05 08:05 → *-03-05 08:05:00
                      05:40 → *-*-* 05:40:00
        Sat,Sun 12-05 08:05 → Sat,Sun *-12-05 08:05:00
              Sat,Sun 08:05 → Sat,Sun *-*-* 08:05:00
           2003-03-05 05:40 → 2003-03-05 05:40:00
             2003-02..04-05 → 2003-02..04-05 00:00:00
                 2003-03-05 → 2003-03-05 00:00:00
                      03-05 → *-03-05 00:00:00
                     hourly → *-*-* *:00:00
                      daily → *-*-* 00:00:00
                    monthly → *-*-01 00:00:00
                     weekly → Mon *-*-* 00:00:00
                     yearly → *-01-01 00:00:00
                   annually → *-01-01 00:00:00
```

The `schedule` can be a string or an array of string (to allow for multiple schedules)

Here's an example of a YAML configuration:

```yaml
default:
    repository: "d:\\backup"
    password-file: key

self:
    inherit: default
    retention:
      after-backup: true
      keep-within: 14d
    backup:
        source: "."
        schedule:
        - "Mon..Fri *:00,15,30,45" # every 15 minutes on weekdays
        - "Sat,Sun 0,12:00"        # twice a day on week-ends
        schedule-permission: user
        schedule-lock-wait: 10m
    prune:
        schedule: "sun 3:30"
        schedule-permission: user
        schedule-lock-wait: 1h
```

## Scheduling commands

resticprofile accepts these internal commands:
- schedule
- unschedule
- status

All internal commands either operate on the profile selected by `--name`, on the profiles selected by a group, or on all profiles when the flag `--all` is passed.

Examples:
```
resticprofile --name profile schedule 
resticprofile --name group schedule 
resticprofile schedule --all 
```

Please note, schedules are always independent of each other no matter whether they have been created with `--all`, by group or from a single profile.

### schedule command

Install all the schedules defined on the selected profile or profiles.

Please note on systemd, we need to `start` the timer once to enable it. Otherwise it will only be enabled on the next reboot. If you **dont' want** to start (and enable) it now, pass the `--no-start` flag to the command line.

Also if you use the `--all` flag to schedule all your profiles at once, make sure you use only the `user` mode or `system` mode. A combination of both would not schedule the tasks properly:
- if the user is not privileged, only the `user` tasks will be scheduled
- if the user **is** privileged, **all schedule will end-up as a `system` schedule**

### unschedule command

Remove all the schedules defined on the selected profile or profiles.

### status command

Print the status on all the installed schedules of the selected profile or profiles. 

The display of the `status` command will be OS dependant. Please see the examples below on which output you can expect from it.

### Examples of scheduling commands under Windows

If you create a task with `user` permission under Windows, you will need to enter your password to validate the task. It's a requirement of the task scheduler. I'm inviting you to review the code to make sure I'm not emailing your password to myself. Seriously you shouldn't trust anyone.

Example of the `schedule` command under Windows (with git bash):

```
$ resticprofile -c examples/windows.yaml -n self schedule

Analyzing backup schedule 1/2
=================================
  Original form: Mon..Fri *:00,15,30,45
Normalized form: Mon..Fri *-*-* *:00,15,30,45:00
    Next elapse: Wed Jul 22 21:30:00 BST 2020
       (in UTC): Wed Jul 22 20:30:00 UTC 2020
       From now: 1m52s left

Analyzing backup schedule 2/2
=================================
  Original form: Sat,Sun 0,12:00
Normalized form: Sat,Sun *-*-* 00,12:00:00
    Next elapse: Sat Jul 25 00:00:00 BST 2020
       (in UTC): Fri Jul 24 23:00:00 UTC 2020
       From now: 50h31m52s left

Creating task for user Creative Projects
Task Scheduler requires your Windows password to validate the task: 

2020/07/22 21:28:15 scheduled job self/backup created

Analyzing retention schedule 1/1
=================================
  Original form: sun 3:30
Normalized form: Sun *-*-* 03:30:00
    Next elapse: Sun Jul 26 03:30:00 BST 2020
       (in UTC): Sun Jul 26 02:30:00 UTC 2020
       From now: 78h1m44s left

2020/07/22 21:28:22 scheduled job self/retention created
```

To see the status of the triggers, you can use the `status` command:

```
$ resticprofile -c examples/windows.yaml -n self status

Analyzing backup schedule 1/2
=================================
  Original form: Mon..Fri *:00,15,30,45
Normalized form: Mon..Fri *-*-* *:00,15,30,45:00
    Next elapse: Wed Jul 22 21:30:00 BST 2020
       (in UTC): Wed Jul 22 20:30:00 UTC 2020
       From now: 14s left

Analyzing backup schedule 2/2
=================================
  Original form: Sat,Sun 0,12:*
Normalized form: Sat,Sun *-*-* 00,12:*:00
    Next elapse: Sat Jul 25 00:00:00 BST 2020
       (in UTC): Fri Jul 24 23:00:00 UTC 2020
       From now: 50h29m46s left

           Task: \resticprofile backup\self backup
           User: Creative Projects
    Working Dir: D:\Source\resticprofile
           Exec: D:\Source\resticprofile\resticprofile.exe --no-ansi --config examples/windows.yaml --name self backup
        Enabled: true
          State: ready
    Missed runs: 0
  Last Run Time: 2020-07-22 21:30:00 +0000 UTC
    Last Result: 0
  Next Run Time: 2020-07-22 21:45:00 +0000 UTC

Analyzing retention schedule 1/1
=================================
  Original form: sun 3:30
Normalized form: Sun *-*-* 03:30:00
    Next elapse: Sun Jul 26 03:30:00 BST 2020
       (in UTC): Sun Jul 26 02:30:00 UTC 2020
       From now: 77h59m46s left

           Task: \resticprofile backup\self retention
           User: Creative Projects
    Working Dir: D:\Source\resticprofile
           Exec: D:\Source\resticprofile\resticprofile.exe --no-ansi --config examples/windows.yaml --name self forget
        Enabled: true
          State: ready
    Missed runs: 0
  Last Run Time: 1999-11-30 00:00:00 +0000 UTC
    Last Result: 267011
  Next Run Time: 2020-07-26 03:30:00 +0000 UTC

```

To remove the schedule, use the `unschedule` command:

```
$ resticprofile -c examples/windows.yaml -n self unschedule
2020/07/22 21:34:51 scheduled job self/backup removed
2020/07/22 21:34:51 scheduled job self/retention removed
```

### Examples of scheduling commands under Linux

With this example of configuration for Linux:

```yaml
default:
    password-file: key
    repository: /tmp/backup

test1:
    inherit: default
    backup:
        source: ./
        schedule: "*:00,15,30,45"
        schedule-permission: user
        schedule-lock-wait: 15m
    check:
        schedule: "*-*-1"
        schedule-permission: user
        schedule-lock-wait: 15m

```

```
$ resticprofile -c examples/linux.yaml -n test1 schedule

Analyzing backup schedule 1/1
=================================
  Original form: *:00,15,30,45
Normalized form: *-*-* *:00,15,30,45:00
    Next elapse: Thu 2020-07-23 17:15:00 BST
       (in UTC): Thu 2020-07-23 16:15:00 UTC
       From now: 6min left

2020/07/23 17:08:51 writing /home/user/.config/systemd/user/resticprofile-backup@profile-test1.service
2020/07/23 17:08:51 writing /home/user/.config/systemd/user/resticprofile-backup@profile-test1.timer
Created symlink /home/user/.config/systemd/user/timers.target.wants/resticprofile-backup@profile-test1.timer → /home/user/.config/systemd/user/resticprofile-backup@profile-test1.timer.
2020/07/23 17:08:51 scheduled job test1/backup created

Analyzing check schedule 1/1
=================================
  Original form: *-*-1
Normalized form: *-*-01 00:00:00
    Next elapse: Sat 2020-08-01 00:00:00 BST
       (in UTC): Fri 2020-07-31 23:00:00 UTC
       From now: 1 weeks 1 days left

2020/07/23 17:08:51 writing /home/user/.config/systemd/user/resticprofile-check@profile-test1.service
2020/07/23 17:08:51 writing /home/user/.config/systemd/user/resticprofile-check@profile-test1.timer
Created symlink /home/user/.config/systemd/user/timers.target.wants/resticprofile-check@profile-test1.timer → /home/user/.config/systemd/user/resticprofile-check@profile-test1.timer.
2020/07/23 17:08:51 scheduled job test1/check created
```

The `status` command shows a combination of `journalctl` displaying errors (only) in the last month and `systemctl status`:

```
$ resticprofile -c examples/linux.yaml -n test1 status

Analyzing backup schedule 1/1
=================================
  Original form: *:00,15,30,45
Normalized form: *-*-* *:00,15,30,45:00
    Next elapse: Tue 2020-07-28 15:15:00 BST
       (in UTC): Tue 2020-07-28 14:15:00 UTC
       From now: 4min 44s left

Recent log (>= warning in the last month)
==========================================
-- Logs begin at Wed 2020-06-17 11:09:19 BST, end at Tue 2020-07-28 15:10:10 BST. --
Jul 27 20:48:01 Desktop76 systemd[2986]: Failed to start resticprofile backup for profile test1 in examples/linux.yaml.
Jul 27 21:00:55 Desktop76 systemd[2986]: Failed to start resticprofile backup for profile test1 in examples/linux.yaml.
Jul 27 21:15:34 Desktop76 systemd[2986]: Failed to start resticprofile backup for profile test1 in examples/linux.yaml.

Systemd timer status
=====================
● resticprofile-backup@profile-test1.timer - backup timer for profile test1 in examples/linux.yaml
   Loaded: loaded (/home/user/.config/systemd/user/resticprofile-backup@profile-test1.timer; enabled; vendor preset: enabled)
   Active: active (waiting) since Tue 2020-07-28 15:10:06 BST; 8s ago
  Trigger: Tue 2020-07-28 15:15:00 BST; 4min 44s left

Jul 28 15:10:06 Desktop76 systemd[2951]: Started backup timer for profile test1 in examples/linux.yaml.


Analyzing check schedule 1/1
=================================
  Original form: *-*-1
Normalized form: *-*-01 00:00:00
    Next elapse: Sat 2020-08-01 00:00:00 BST
       (in UTC): Fri 2020-07-31 23:00:00 UTC
       From now: 3 days left

Recent log (>= warning in the last month)
==========================================
-- Logs begin at Wed 2020-06-17 11:09:19 BST, end at Tue 2020-07-28 15:10:10 BST. --
Jul 27 19:39:59 Desktop76 systemd[2986]: Failed to start resticprofile check for profile test1 in examples/linux.yaml.

Systemd timer status
=====================
● resticprofile-check@profile-test1.timer - check timer for profile test1 in examples/linux.yaml
   Loaded: loaded (/home/user/.config/systemd/user/resticprofile-check@profile-test1.timer; enabled; vendor preset: enabled)
   Active: active (waiting) since Tue 2020-07-28 15:10:07 BST; 7s ago
  Trigger: Sat 2020-08-01 00:00:00 BST; 3 days left

Jul 28 15:10:07 Desktop76 systemd[2951]: Started check timer for profile test1 in examples/linux.yaml.


```

And `unschedule`:

```
$ resticprofile -c examples/linux.yaml -n test1 unschedule
Removed /home/user/.config/systemd/user/timers.target.wants/resticprofile-backup@profile-test1.timer.
2020/07/23 17:13:42 scheduled job test1/backup removed
Removed /home/user/.config/systemd/user/timers.target.wants/resticprofile-check@profile-test1.timer.
2020/07/23 17:13:42 scheduled job test1/check removed
```

### Examples of scheduling commands under macOS

macOS has a very tight protection system when running scheduled tasks (also called agents).

Under macOS, resticprofile is asking if you want to start a profile right now so you can give the access needed to the task, which consists on a few popup windows (you can disable this behavior by adding the flag `--no-start` after the schedule command).

Here's an example of scheduling a backup to Azure (which needs network access):

```
% resticprofile -v -c examples/private/azure.yaml -n self schedule

Analyzing backup schedule 1/1
=================================
  Original form: *:0,15,30,45:00
Normalized form: *-*-* *:00,15,30,45:00
    Next elapse: Tue Jul 28 23:00:00 BST 2020
       (in UTC): Tue Jul 28 22:00:00 UTC 2020
       From now: 2m34s left


By default, a macOS agent access is restricted. If you leave it to start in the background it's likely to fail.
You have to start it manually the first time to accept the requests for access:

% launchctl start local.resticprofile.self.backup

Do you want to start it now? (Y/n):
2020/07/28 22:57:26 scheduled job self/backup created
```

Right after you started the profile, you should get some popup asking you to grant access to various files/folders/network.

If you backup your files to an external repository on a network, you should get this popup window:

!["resticprofile" would like to access files on a network volume](https://github.com/creativeprojects/resticprofile/raw/master/network_volume.png)

**Note:**
If you prefer not being asked, you can add the `--no-start` flag like so:

```
% resticprofile -v -c examples/private/azure.yaml -n self schedule --no-start
```

## Changing schedule-permission from user to system, or system to user

If you need to change the permission of a schedule, **please be sure to `unschedule` the profile before**.

This order is important:

- `unschedule` the job first. resticprofile does **not** keep track of how your profile **was** installed, so you have to remove the schedule first
- now you can change your permission (`user` to `system`, or `system` to `user`)
- `schedule` your updated profile

# Status file for easy monitoring

If you need to escalate the result of your backup to a monitoring system, you can definitely use the `run-after` and `run-after-fail` scripting.

But sometimes we just need something simple that a monitoring system can regularly check. For that matter, resticprofile can generate a simple JSON file with the details of the latest backup/forget/check command. For example I have a Zabbix agent [checking this file](https://github.com/creativeprojects/resticprofile/tree/master/contrib/zabbix) once a day, and so you can hook up any monitoring system that can interpret a JSON file.

In your profile, you simply need to add a new parameter, which is the location of your status file

```toml
[profile]
status-file = "backup-status.json"
```

Here's an example of a generated file, where you can see that the last `check` failed, whereas the last `backup` succeeded:

```json
{
  "profiles": {
    "self": {
      "backup": {
        "success": true,
        "time": "2021-03-24T16:36:56.831077Z",
        "error": "",
        "stderr": "",
        "duration": 16,
        "files_new": 215,
        "files_changed": 0,
        "files_unmodified": 0,
        "dirs_new": 58,
        "dirs_changed": 0,
        "dirs_unmodified": 0,
        "files_total": 215,
        "bytes_added": 296536447,
        "bytes_total": 362952485
      },
      "check": {
        "success": false,
        "time": "2021-03-24T15:23:40.270689Z",
        "error": "exit status 1",
        "stderr": "unable to create lock in backend: repository is already locked exclusively by PID 18534 on dingo by cloud_user (UID 501, GID 20)\nlock was created at 2021-03-24 15:23:29 (10.42277s ago)\nstorage ID 1bf636d2\nthe `unlock` command can be used to remove stale locks\n",
        "duration": 1
      }
    }
  }
}
```

## Extended status

Note: In the backup section above you can see some fields like `files_new`, `files_total`, etc. This information is only available when resticprofile's output is either *not* sent to the terminal (e.g. redirected) or when you add the flag `extended-status` to your backup configuration.
This is a technical limitation to ensure restic displays terminal output correctly. 

`extended-status` or stdout redirection is **not needed** for these fields:
- success
- time
- error
- stderr
- duration

`extended-status` is **not set by default because it hides any output from restic**

```yaml
profile:
    inherit: default
    status-file: /home/backup/status.json
    backup:
        extended-status: true
        source: /go
        exclude:
          - "/**/.git/"

```

# Prometheus

resticprofile can generate a prometheus file, or send the report to a push gateway. For now, only a `backup` command will generate a report.
Here's a configuration example with both options to generate a file and send to a push gateway:

```yaml
root:
    inherit: default
    prometheus-save-to-file: "root.prom"
    prometheus-push: "http://localhost:9091/"
    backup:
        extended-status: true
        no-error-on-warning: true
        source:
          - /
```

Please note you need to set `extended-status` to `true` if you want all the available metrics. See [Extended status](#extended-status) for more information.

Here's an example of the generated prometheus file:

```
# HELP resticprofile_backup_added_bytes Total number of bytes added to the repository.
# TYPE resticprofile_backup_added_bytes gauge
resticprofile_backup_added_bytes{profile="root"} 35746
# HELP resticprofile_backup_dir_changed Number of directories with changes.
# TYPE resticprofile_backup_dir_changed gauge
resticprofile_backup_dir_changed{profile="root"} 9
# HELP resticprofile_backup_dir_new Number of new directories added to the backup.
# TYPE resticprofile_backup_dir_new gauge
resticprofile_backup_dir_new{profile="root"} 0
# HELP resticprofile_backup_dir_unmodified Number of directories unmodified since last backup.
# TYPE resticprofile_backup_dir_unmodified gauge
resticprofile_backup_dir_unmodified{profile="root"} 314
# HELP resticprofile_backup_duration_seconds The backup duration (in seconds).
# TYPE resticprofile_backup_duration_seconds gauge
resticprofile_backup_duration_seconds{profile="root"} 0.946567354
# HELP resticprofile_backup_files_changed Number of files with changes.
# TYPE resticprofile_backup_files_changed gauge
resticprofile_backup_files_changed{profile="root"} 3
# HELP resticprofile_backup_files_new Number of new files added to the backup.
# TYPE resticprofile_backup_files_new gauge
resticprofile_backup_files_new{profile="root"} 0
# HELP resticprofile_backup_files_processed Total number of files scanned by the backup for changes.
# TYPE resticprofile_backup_files_processed gauge
resticprofile_backup_files_processed{profile="root"} 3925
# HELP resticprofile_backup_files_unmodified Number of files unmodified since last backup.
# TYPE resticprofile_backup_files_unmodified gauge
resticprofile_backup_files_unmodified{profile="root"} 3922
# HELP resticprofile_backup_processed_bytes Total number of bytes scanned for changes.
# TYPE resticprofile_backup_processed_bytes gauge
resticprofile_backup_processed_bytes{profile="root"} 3.8524672e+07
# HELP resticprofile_backup_status Backup status: 0=fail, 1=warning, 2=success.
# TYPE resticprofile_backup_status gauge
resticprofile_backup_status{profile="root"} 1
# HELP resticprofile_build_info resticprofile build information.
# TYPE resticprofile_build_info gauge
resticprofile_build_info{goversion="go1.16.6",version="0.15.0-dev"} 1

```

## User defined labels

You can add your own prometheus labels. Please note they will be applied to **all** the metrics.
Here's an example:

```yaml
root:
    inherit: default
    prometheus-save-to-file: "root.prom"
    prometheus-push: "http://localhost:9091/"
    prometheus-labels:
      - host: {{ .Hostname }}
    backup:
        extended-status: true
        no-error-on-warning: true
        source:
          - /
```

which will add the `host` label to all your metrics.


# Variable expansion in configuration file

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

# Configuration file reference

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
* **restic-lock-retry-after**: duration
* **restic-stale-lock-age**: duration
* **min-memory**: integer (MB)
* **scheduler**: string (`crond` is the only non-default value)
* **systemd-unit-template**: string (file containing a go template to generate systemd unit file)
* **systemd-timer-template**: string (file containing a go template to generate systemd timer file)

`[profile]`

`profile` is the name of your profile

Flags used by resticprofile only

* ****inherit****: string
* **description**: string
* **initialize**: true / false
* **lock**: string: specify a local lockfile
* **force-inactive-lock**: true / false
* **run-before**: string OR list of strings
* **run-after**: string OR list of strings
* **run-after-fail**: string OR list of strings
* **status-file**: string
* **prometheus-save-to-file**: string
* **prometheus-push**: string

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
* **repository-file**: string
* **tls-client-cert**: string
* **verbose**: true / false OR integer

`[profile.backup]`

Flags used by resticprofile only

* **run-before**: string OR list of strings
* **run-after**: string OR list of strings
* **check-before**: true / false
* **check-after**: true / false
* **schedule**: string OR list of strings
* **schedule-permission**: string (`user` or `system`)
* **schedule-lock-mode**: string (`default`, `fail` or `ignore`) 
* **schedule-lock-wait**: duration 
* **schedule-log**: string
* **stdin-command**: string OR list of strings
* **extended-status**: true / false
* **no-error-on-warning**: true / false

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
* **schedule**: string OR list of strings
* **schedule-permission**: string (`user` or `system`)
* **schedule-lock-mode**: string (`default`, `fail` or `ignore`)
* **schedule-lock-wait**: duration
* **schedule-log**: string

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
* **tag**: true / false, string OR list of strings
* **path**: true / false, string OR list of strings
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
* **path**: true / false, string OR list of strings
* **tag**: true / false, string OR list of strings

`[profile.forget]`

Flags used by resticprofile only

* **schedule**: string OR list of strings
* **schedule-permission**: string (`user` or `system`)
* **schedule-lock-mode**: string (`default`, `fail` or `ignore`)
* **schedule-lock-wait**: duration
* **schedule-log**: string

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
* **tag**: true / false, string OR list of strings
* **path**: true / false, string OR list of strings
* **compact**: true / false
* **group-by**: string
* **dry-run**: true / false
* **prune**: true / false

`[profile.check]`

Flags used by resticprofile only

* **schedule**: string OR list of strings
* **schedule-permission**: string (`user` or `system`)
* **schedule-lock-mode**: string (`default`, `fail` or `ignore`)
* **schedule-lock-wait**: duration
* **schedule-log**: string

Flags passed to the restic command line

* **check-unused**: true / false
* **read-data**: true / false
* **read-data-subset**: string
* **with-cache**: true / false

`[profile.prune]`

Flags used by resticprofile only

* **schedule**: string OR list of strings
* **schedule-permission**: string (`user` or `system`)
* **schedule-lock-mode**: string (`default`, `fail` or `ignore`)
* **schedule-lock-wait**: duration
* **schedule-log**: string

`[profile.mount]`

Flags passed to the restic command line

* **allow-other**: true / false
* **allow-root**: true / false
* **host**: true / false OR string
* **no-default-permissions**: true / false
* **owner-root**: true / false
* **path**: true / false, string OR list of strings
* **snapshot-template**: string
* **tag**: true / false, string OR list of strings

`[profile.copy]`

Flags used by resticprofile only

* **initialize**: true / false
* **schedule**: string OR list of strings
* **schedule-permission**: string (`user` or `system`)
* **schedule-lock-mode**: string (`default`, `fail` or `ignore`)
* **schedule-lock-wait**: duration
* **schedule-log**: string

Flags passed to the restic command line

* **key-hint**: string **(will be passed as 'key-hint2')**
* **password-command**: command **(will be passed as 'password-command2')**
* **password-file**: string **(will be passed as 'password-file2')**
* **host**: true / false OR string
* **path**: true / false, string OR list of strings
* **repository**: repository **(will be passed as 'repo2')**
* **repository-file**: string **(will be passed as 'repository-file2')**
* **tag**: true / false, string OR list of strings

`[profile.dump]`

Flags passed to the restic command line

* **host**: true / false OR string
* **path**: true / false, string OR list of strings
* **tag**: true / false, string OR list of strings

`[profile.find]`

Flags passed to the restic command line

* **host**: true / false OR string
* **path**: true / false, string OR list of strings
* **tag**: true / false, string OR list of strings

`[profile.ls]`

Flags passed to the restic command line

* **host**: true / false OR string
* **path**: true / false, string OR list of strings
* **tag**: true / false, string OR list of strings

`[profile.restore]`

Flags passed to the restic command line

* **host**: true / false OR string
* **path**: true / false, string OR list of strings
* **tag**: true / false, string OR list of strings

`[profile.stats]`

Flags passed to the restic command line

* **host**: true / false OR string
* **path**: true / false, string OR list of strings
* **tag**: true / false, string OR list of strings

`[profile.tag]`

Flags passed to the restic command line

* **host**: true / false OR string
* **path**: true / false, string OR list of strings
* **tag**: true / false, string OR list of strings


# Using resticprofile and systemd

systemd is a common service manager in use by many Linux distributions.
resticprofile has the ability to create systemd timer and service files.
systemd can be used in place of cron to schedule backups.

User systemd units are created under the user's systemd profile (~/.config/systemd/user).

System units are created in /etc/systemd/system

## systemd calendars

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

## First time schedule

When you schedule a profile with the `schedule` command, under the hood resticprofile will
- create the unit file (of type `notify`)
- create the timer file
- run `systemctl daemon-reload` (only if `schedule-permission` is set to `system`)
- run `systemctl enable`
- run `systemctl start`

## How to change the default systemd unit and timer file using a template

By default, an opinionated systemd unit and timer are automatically generated by resticprofile.

Since version 0.16.0, you now can describe your own templates if you need to add things in it (typically like sending an email on failure).

The format used is a [go template](https://pkg.go.dev/text/template) and you need specify your own unit and/or timer file in the global section of the configuration (it will apply to all your profiles):

```yaml
---
global:
    systemd-unit-template: service.tmpl
    systemd-timer-template: timer.tmpl
```
Here are the defaults if you don't specify your own (which I recommend to use as a starting point for your own templates)

### Default unit file

```
[Unit]
Description={{ .JobDescription }}

[Service]
Type=notify
WorkingDirectory={{ .WorkingDirectory }}
ExecStart={{ .CommandLine }}
{{ if .Nice }}Nice={{ .Nice }}{{ end }}
{{ range .Environment -}}
Environment="{{ . }}"
{{ end -}}
```

### Default timer file

```
[Unit]
Description={{ .TimerDescription }}

[Timer]
{{ range .OnCalendar -}}
OnCalendar={{ . }}
{{ end -}}
Unit={{ .SystemdProfile }}
Persistent=true

[Install]
WantedBy=timers.target
```

### Template variables

These are available for both the unit and timer templates:

* JobDescription   *string*
* TimerDescription *string*
* WorkingDirectory *string*
* CommandLine      *string*
* OnCalendar       *array of strings*
* SystemdProfile   *string*
* Nice             *integer*
* Environment      *array of strings*

# Using resticprofile and launchd on macOS

`launchd` is the service manager on macOS. resticprofile can schedule a profile via a _user agent_ or a _daemon_ in launchd.

## User agent

A user agent is generated when you set `schedule-permission` to `user`.

It consists of a `plist` file in the folder `~/Library/LaunchAgents`:

A user agent **mostly** runs with the privileges of the user. But if you backup some specific files, like your contacts or your calendar for example, you will need to give more permissions to resticprofile **and** restic.

For this to happen, you need to start the agent or daemon from a console window first (resticprofile will ask if you want to do so)

If your profile is a backup profile called `remote`, the command to run manually is:

```
% launchctl start local.resticprofile.remote.backup
```

Once you grant the permission, the background agents/daemon will be able to run normally.

There's some information in this thread: https://github.com/restic/restic/issues/2051

*TODO: I'm going to try to compile a comprehensive how-to guide from all the information from the thread. Stay tuned!*

### Special case of schedule-permission=user with sudo

Please note if you schedule a user agent while running resticprofile with sudo: the user agent will be registered to the root user, and not your initial user context. It means you can only see it (`status`) and remove it (`unschedule`) via sudo.

## Daemon

A launchd daemon is generated when you set `schedule-permission` to `system`. 

It consists of a `plist` file in the folder `/Library/LaunchDaemons`. You have to run resticprofile with sudo to `schedule`, check the  `status` and `unschedule` the profile.

# Contributions

Please share your resticprofile recipes, fancy configuration files, or tips and tricks.
I have created a [contributions section](https://github.com/creativeprojects/resticprofile/tree/master/contrib) for that matter.
