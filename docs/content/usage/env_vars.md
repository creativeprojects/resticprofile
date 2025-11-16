---
title: Environment Variables
weight: 7
---

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
