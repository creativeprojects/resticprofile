---
title: Command Line Reference
weight: 5
---
## Version

The `version` command displays resticprofile version. If run in verbose mode (using `--verbose` flag) additional information such as OS version, golang version and modules are displayed as well.

```shell
resticprofile --verbose version
```

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
