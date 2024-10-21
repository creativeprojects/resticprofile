---
archetype: "chapter"
pre: "<b>3. </b>"
title: "Using resticprofile"
weight: 3
# tags: ["v0.17.0"]
---

Here are a few examples how to run resticprofile (using the main example configuration file)

See all snapshots of your `default` profile:

```shell
resticprofile
```

See all available profiles in your configuration file (and the restic commands where some flags are defined):

```shell
resticprofile profiles

Profiles available (name, sections, description):
  root:           (backup, copy, forget, retention)
  self:           (backup, check, copy, forget, retention)
  src:            (backup, copy, retention, snapshots)

Groups available (name, profiles, description):
  full-backup:  [root, src]

```

Backup root & src profiles (using _full-backup_ group shown earlier)

```shell
resticprofile --name "full-backup" backup
```
or using the syntax introduced in v0.17.0:

```shell
resticprofile full-backup.backup
```

Assuming the _stdin_ profile from the configuration file shown before, the command to send a mysqldump to the backup is as simple as:

```shell
mysqldump --all-databases --order-by-primary | resticprofile --name stdin backup
```
or using the syntax introduced in v0.17.0:

```shell
mysqldump --all-databases --order-by-primary | resticprofile stdin.backup
```

Mount the default profile (_default_) in /mnt/restic:

```shell
resticprofile mount /mnt/restic
```

Display quick help

```shell
resticprofile --help
```

A command is either a restic command or a resticprofile own command.


## Command line reference

There are not many options on the command line, most of the options are in the configuration file.

* **[-h | --help]**: Display quick help
* **[-c | --config] configuration_file**: Specify a configuration file other than the default ("profiles")
* **[-f | --format] configuration_format**: Specify the configuration file format: `toml`, `yaml`, `json` or `hcl`
* **[-n | --name] profile_name**: Profile section to use from the configuration file.
  You can also use `[profile_name].[command]` syntax instead, this will only work if `-n` is not set.
  Using `-n [profile_name] [command]` or `[profile_name].[command]` both select profile and command and are technically equivalent.
* **[--dry-run]**: Doesn't run the restic commands but display the command lines instead
* **[-q | --quiet]**: Force resticprofile and restic to be quiet (override any configuration from the profile)
* **[-v | --verbose]**: Force resticprofile and restic to be verbose (override any configuration from the profile)
* **[--trace]**: Display even more debugging information
* **[--no-ansi]**: Disable console colouring (to save output into a log file)
* **[--stderr]**: Send console output from resticprofile to stderr (is enabled for commands `cat` and `dump`)
* **[--no-lock]**: Disable resticprofile locks, neither create nor fail on a lock. restic locks are unaffected by this option.
* **[--theme]**: Can be `light`, `dark` or `none`. The colours will adjust to a 
light or dark terminal (none to disable colouring)
* **[--lock-wait] duration**: Retry to acquire resticprofile and restic locks for up to the specified amount of time before failing on a lock failure. 
* **[-l | --log] file path or url**: To write the logs to a file or a syslog server instead of displaying on the console. 
The format of the syslog server url is `syslog-tcp://192.168.0.1:514`, `syslog://udp-server:514` or `syslog:`.
For custom log forwarding, the prefix `temp:` can be used (e.g. `temp:/t/msg.log`) to create unique log output that can be fed 
into a command or http hook by referencing it with `"{{ tempFile "msg.log" }}"` in the configuration file.
* **[--command-output]**: Sets how to redirect command output when a log target is specified. Can be `auto`, `log`, `console` or `all`.
* **[-w | --wait]**: Wait at the very end of the execution for the user to press enter. 
This is only useful in Windows when resticprofile is started from explorer and the console window closes automatically at the end.
* **[--ignore-on-battery]**: Don't start the profile when the computer is running on battery. You can specify a value to ignore only when the % charge left is less or equal than the value.
* **[resticprofile OR restic command]**: Like snapshots, backup, check, prune, forget, mount, etc.
* **[additional flags]**: Any additional flags to pass to the restic command line

## Environment variables

Most flags for resticprofile can be set using environment variables. If both are specified, command line flags take the precedence.

| Flag                  | Environment variable              | Built-In default |
|-----------------------|-----------------------------------|------------------|
| `--quiet`             | `RESTICPROFILE_QUIET`             | `false`          |
| `--verbose`           | `RESTICPROFILE_VERBOSE`           | `false`          |
| `--trace`             | `RESTICPROFILE_TRACE`             | `false`          |
| `--config`            | `RESTICPROFILE_CONFIG`            | `"profiles"`     |
| `--format`            | `RESTICPROFILE_FORMAT`            | `""`             |
| `--name`              | `RESTICPROFILE_NAME`              | `"default"`      |
| `--log`               | `RESTICPROFILE_LOG`               | `""`             |
| `--command-output`    | `RESTICPROFILE_COMMAND_OUTPUT`    | `"auto"`         |
| `--dry-run`           | `RESTICPROFILE_DRY_RUN`           | `false`          |
| `--no-lock`           | `RESTICPROFILE_NO_LOCK`           | `false`          |
| `--lock-wait`         | `RESTICPROFILE_LOCK_WAIT`         | `0`              |
| `--stderr`            | `RESTICPROFILE_STDERR`            | `false`          |
| `--no-ansi`           | `RESTICPROFILE_NO_ANSI`           | `false`          |
| `--theme`             | `RESTICPROFILE_THEME`             | `"light"`        |
| `--no-priority`       | `RESTICPROFILE_NO_PRIORITY`       | `false`          |
| `--wait`              | `RESTICPROFILE_WAIT`              | `false`          |
| `--ignore-on-battery` | `RESTICPROFILE_IGNORE_ON_BATTERY` | `0`              |

### Other environment variables

| Environment Variable            | Default | Purpose                                                                              |
|---------------------------------|---------|--------------------------------------------------------------------------------------|
| `RESTICPROFILE_PWSH_NO_AUTOENV` | _empty_ | Disables powershell script pre-processing that converts unset `$VAR` into `$Env:VAR` |

### Environment variables set by resticprofile

| Environment Variable        | Example                        | When                                |
|-----------------------------|--------------------------------|-------------------------------------|
| `RESTICPROFILE_SCHEDULE_ID` | `profiles.yaml:backup@profile` | Set when running scheduled commands |
