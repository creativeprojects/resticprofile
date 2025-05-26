---
archetype: chapter
pre: <b>8. </b>
title: Release Notes
weight: 8
---

# v0.31.0 (2025-05-26)

## 🍸 Spring Bank Holiday release 🌞 

This release includes several long-awaited improvements.

### Improvements:  
- Add support for the new Restic flag `exclude-cloud-files` introduced in version 0.18.0 (Windows only).  
- Include missing `run-*` and `send-*` targets in all commands (previously missing in `check` and `copy` commands).  
- Send Restic version as a new Prometheus metric.  
- Support `.yml` files when searching for configuration files (previously only `.yaml` was considered).  

### Bug fixes:  
- Ignore missing drop-in files to prevent creating an empty systemd schedule.  


## Changelog
* [Add missing run-* and send-* to all command sections (#504)](https://github.com/creativeprojects/resticprofile/commit/f7566acc7d87e8f099a3e7ef5b535611c75a04f9)
* [add support for windows only flag exclude-cloud-files (#500)](https://github.com/creativeprojects/resticprofile/commit/88fcedbd7c4b72eee9b0dcfe23236e8d992ea4a6)
* [add support for yml extension when searching for configuration file (#506)](https://github.com/creativeprojects/resticprofile/commit/81a6e45f50e3c2196ff4e62e85bfb61e0e5816a0)
* [add version of restic to the Prometheus metrics (#495)](https://github.com/creativeprojects/resticprofile/commit/43b6c849331aa4b0e1126df57d556e5ba642f1fc)
* [fix: remove duplicated field](https://github.com/creativeprojects/resticprofile/commit/763a2cdf83c9cd644de4cd275bbaab8c5c4f62b5)
* [ignore missing drop-in files to avoid creating an empty configuration (#502)](https://github.com/creativeprojects/resticprofile/commit/c3add93ec2debf6e430c9588dc9096ed3c82d936)
* [replace deprecated sonar scan action (#503)](https://github.com/creativeprojects/resticprofile/commit/8fd3ddbdfdd0798bdd410909c93dfeaf854dc539)



# v0.30.1 (2025-05-05)

## 🐛 Bug fix release 🐞 

This releases fixes a regression bug when scheduling `system` tasks with `systemd`: in some cases the HOME directory was replaced by the home directory of the user who ran the `sudo` command

Bug report: https://github.com/creativeprojects/resticprofile/issues/493


## Changelog
* [fix root home regression bug in 0.30.0 (#497)](https://github.com/creativeprojects/resticprofile/commit/ceff435459d5a1327af6e5e1f449957b5a8166a3)



# v0.30.0 (2025-04-04)

## 🌱 Spring release 🌸

This release introduces features and fixes for profile scheduling.

### ⚠️ Breaking change

Until this release, the `user` scheduling permission was broken. With `systemd` or the default macOS scheduler, the permission functioned as `user_logged_on`, running the profile only when the user was logged in.

This issue is now fixed for new schedules.

To update existing schedules, run `unschedule` and then `schedule` again.

For `systemd`, resticprofile requires root privileges (via sudo) to schedule with `user` permission, as it now uses a system unit running as the user.

### Scheduler

Scheduling has been significantly improved with the ability to read existing schedules. The `status` and `unschedule` commands now detect any resticprofile schedules from the selected configuration file, even if the profile is no longer present.

Due to recent error reports, Windows Task Scheduler integration has been completely rewritten to use the `schtasks` CLI tool.

### Other

- Fixed issue with non-existent battery reported on recent Mac desktop hardware.  
- Added support for Restic 0.18.  
- Included pre-built binary for Windows on ARM64.  
- Upload Resticprofile container to GitHub Container Registry.  


## Changelog
* [Correct configuration search paths on Windows (#478)](https://github.com/creativeprojects/resticprofile/commit/f14094b98819e6fc6ebd001547ede9b2d58bd2ce)
* [Create FUNDING.yml](https://github.com/creativeprojects/resticprofile/commit/1b05c7a6932b188d0250dd0e948889296c976d26)
* [Documentation versioning (#464)](https://github.com/creativeprojects/resticprofile/commit/b6b8285e6226478aceb9c112eec2cd0a093ac5d4)
* [Run tests on read only source dir (#461)](https://github.com/creativeprojects/resticprofile/commit/edc129250c542c29d8e999aeca557b0cfc24b529)
* [Schedule user logged in (#470)](https://github.com/creativeprojects/resticprofile/commit/ae073aeb8c1b82f38bd61cbb8c0ca5e1f82cdf9e)
* [Scheduling improvements (#425)](https://github.com/creativeprojects/resticprofile/commit/8d8a0dac43b9023d951272ea95c0703b79207fd5)
* [Set schedule permissions (#476)](https://github.com/creativeprojects/resticprofile/commit/e490c429bf167f7e73e952692fe527f01f6c28c8)
* [Upload to Github Container Registry (#479)](https://github.com/creativeprojects/resticprofile/commit/07df1419c8dae8effe8e1bb46d64a7e8ea912f2c)
* [Use schtasks.exe CLI to schedule jobs on Windows (#459)](https://github.com/creativeprojects/resticprofile/commit/41039712ee76a675ad60eb9e72a6c3a6e314d4b8)
* [add microarchitecture to version command (#466)](https://github.com/creativeprojects/resticprofile/commit/ae38db96f2419b9c3cf3c5f95a48474e5db01ed9)
* [add other restic exit codes](https://github.com/creativeprojects/resticprofile/commit/87cfd5db613cc5b1a1a2ab71afddd6a08989e041)
* [chore: add pre-built binary for windows arm64](https://github.com/creativeprojects/resticprofile/commit/ac3b810f2f735642d0a0cea8ec46c585b80904bc)
* [chore: add support for restic 0.18](https://github.com/creativeprojects/resticprofile/commit/edbea57667a422ccce0237b06a8072f305dc6ee6)
* [don't include systemd dependencies on macos & windows](https://github.com/creativeprojects/resticprofile/commit/c5a5dfa4c547318f02f661f243ed6323cc8dffc4)
* [fix for macOs ghost battery (#468)](https://github.com/creativeprojects/resticprofile/commit/af9ab7f4e8f63f7b84e1939d97b3277e20e9000d)
* [remove devcontainer config](https://github.com/creativeprojects/resticprofile/commit/6c9b3cdd7b38bbf05f721c5905ffeb25b9ee0087)
* [replace experimental maps package with standard library (#467)](https://github.com/creativeprojects/resticprofile/commit/b867b01f00d7dfa0b06c9e079880c103d9af614d)



# v0.29.1 (2025-02-06)

## ❄️  Small maintenance release ☃️ 

Not much going on in this maintenance release while I'm still working on a big refactoring of the scheduling for the next release.

* logrotate package added to the Docker image
* upgrade dependencies to fix security vulnerabilities

## Changelog
* [add logrotate to docker image #440](https://github.com/creativeprojects/resticprofile/commit/e2297a6f4f0f03a1733177db1dc950c6e05ddd88)
* [upgrade package - CVE-2024-45337](https://github.com/creativeprojects/resticprofile/commit/37c9aef3032005a88ed006059432a2baddc97ab6)
* [upgrade packages](https://github.com/creativeprojects/resticprofile/commit/467b82ee1eade497c9d23c0dfe69a889143dc807)



# v0.29.0 (2024-10-28)

## 🧙🏻‍♀️  Halloween edition 🎃 

- Finally a long standing bug was fixed in this version: use proper `nice` and `ionice` values on `systemd` scheduled tasks.
- Also the last chunk of work for the configuration `v2`: we can now schedule groups!
- Improvement on the JSON schema: a single URL can be used for `v1` and `v2` configuration
- Another long standing bug on Windows: escape configuration file path on scheduled tasks

## Changelog
* [Escape config file name in schedule parameters (#420)](https://github.com/creativeprojects/resticprofile/commit/fb6c7a2fbb0d3a955a0cdb53e87c58a32a532022)
* [Fix scheduling arguments (#423)](https://github.com/creativeprojects/resticprofile/commit/1d3ec32e293c9d3c4918df7ea7bd6a62a3eb2168)
* [Global json schema (auto versioning) (#412)](https://github.com/creativeprojects/resticprofile/commit/62c5c6472f4ef5462bbafc02d6e75d7684a40abf)
* [Group schedule (#418)](https://github.com/creativeprojects/resticprofile/commit/2a34ea77e5d4c8367436e2ebb2c7cf82973f39a5)
* [Setup systemd priority (#409)](https://github.com/creativeprojects/resticprofile/commit/830d0fd4f8313a0de7a04ffd48d2b44cc797929a)
* [doc: upgrade hugo theme](https://github.com/creativeprojects/resticprofile/commit/62f311814f0f1e034942f3bf6a95dff3590597e3)
* [fix deprecated option in goreleaser config](https://github.com/creativeprojects/resticprofile/commit/da03de768a30c8d6656be5cce39d35a9e42e0817)



# v0.28.1 (2024-10-02)

## 🍂 Autumn bug fixing 🍁  

- fix JSON schema for TOML files that stopped working some time ago (when using `Even Better TOML` extension on VSCode)
- [configuration v2] fix regression bug where profile groups were stopping after an error when the `continue-on-error` flag was set

## Changelog
* [Fix for continue-on-error broken in 0.26.0 (#407)](https://github.com/creativeprojects/resticprofile/commit/57d1439ea33e6da9a5e19010cf2fc95b951ea27e)
* [Linters (#397)](https://github.com/creativeprojects/resticprofile/commit/ccccfa2d0a56432af84dda9f3833e760e30d6a48)
* [add base url on json schema (#408)](https://github.com/creativeprojects/resticprofile/commit/249fd41b3d8e1af034594c9617bc3145a6a3e29a)
* [add misspell linter](https://github.com/creativeprojects/resticprofile/commit/b254503506c76700eb74868677facaed250a7730)
* [fix bullet points from restic man (#398)](https://github.com/creativeprojects/resticprofile/commit/7dcb41bf248da13a58d551838cca7eb2412eaaf2)
* [fix overflow integer conversion lint warning from gosec (#399)](https://github.com/creativeprojects/resticprofile/commit/4181d9e809539ede5b2d73aeb682fa9324e493db)
* [generate JSON schema for restic 0.17](https://github.com/creativeprojects/resticprofile/commit/22317937234ab678c899c15ac7b1900dc7fe4b54)
* [prepare next release](https://github.com/creativeprojects/resticprofile/commit/0ad1b68504f41fc6e7ca11297683ec4e2dc02ff8)
* [use goreleaser v2](https://github.com/creativeprojects/resticprofile/commit/3612b519aa5c0e41ae302ce5fda136db288e0837)



# v0.28.0 (2024-08-17)

## 🌞  Sunny Summer Edition 🌻 

Two big things in this release:
* support for all the new commands and flags of `restic 0.17`
* experimental support for environment variables in configuration flags

### Example of using environment variables in configuration:

```yaml
check-repo-profile:
    inherit: default

    run-before:
        - "echo DOW=`date +\"%u\"` >> {{ env }}"

    check:
        read-data-subset: "$DOW/7"

```

## Changelog
* [Add contributed docs for protected configuration on Windows (#394)](https://github.com/creativeprojects/resticprofile/commit/3b42e4a)
* [Restic 0.17 (#396)](https://github.com/creativeprojects/resticprofile/commit/1e80d02)
* [build with go 1.23](https://github.com/creativeprojects/resticprofile/commit/eb8c721)
* [convert priority value to lowercase](https://github.com/creativeprojects/resticprofile/commit/1df3711)
* [fix panic when sending on closed handler](https://github.com/creativeprojects/resticprofile/commit/d74c521)
* [refactoring of arguments generation (#387)](https://github.com/creativeprojects/resticprofile/commit/49add79)
* [upgrade packages](https://github.com/creativeprojects/resticprofile/commit/cff446b)



# v0.27.1 (2024-07-08)

## 🌦️ Rainy Summer Edition ☔ 

Fix of a regression bug preventing resticprofile from sending logs to a temporary session log file (prefixed with `temp:/t/`)
Thanks @iluvatyr for the quick bug report 👍🏻 

## Changelog
* [fix creation of mock binaries during unit tests (#375)](https://github.com/creativeprojects/resticprofile/commit/a615444)
* [fix documentation for release](https://github.com/creativeprojects/resticprofile/commit/c9f87ec)
* [fix regression with temporary log file (#386)](https://github.com/creativeprojects/resticprofile/commit/eb13003)
* [prepare next release](https://github.com/creativeprojects/resticprofile/commit/abf35a7)
* [refactor unit tests on package lock (#374)](https://github.com/creativeprojects/resticprofile/commit/854f6c9)
* [remove flaky test](https://github.com/creativeprojects/resticprofile/commit/ffffabf)
* [remove goarm linked to the internal variable of runtime go1.23 will apparently forbid the use of linkname to the standard library](https://github.com/creativeprojects/resticprofile/commit/5cfdca2)



# v0.27.0 (2024-06-27)

## 🌻 Summer release 🌞

Summer is here in the Northern Hemisphere! A new release is here too! 🎉

A lot of goodies in this release:

### new run-schedule command

It's more a _behind the scene_ feature: you no longer need to re-schedule your targets after you changed the configuration. The generated schedule command line is now using this new `run-schedule` command that reads all the newest bits from the configuration.

You might want to `unschedule` and `schedule` all your profiles one last time to replace the existing command line using the new `run-schedule` command.

More information: https://creativeprojects.github.io/resticprofile/schedules/commands/index.html#run-schedule-command

### direct support for crontab files

Previously resticprofile was using the `crontab` tool to read and write `crond` schedules. Now you can directly setup a `crontab` file no matter which tool is going to consume it.

More information: https://creativeprojects.github.io/resticprofile/schedules/index.html

### keep content of configuration variables between commands

This was a big issue for some time. The wait is over! You can now set a configuration variable anywhere in a script and use its value later.

More information: https://creativeprojects.github.io/resticprofile/configuration/run_hooks/index.html#passing-environment-variables

### "RESTICPROFILE_{FLAGNAME}" env vars

All cli flags can now be set using environment variables.

More information: [#334](https://github.com/creativeprojects/resticprofile/issues/334)

### allow controlling command output redirection

Allow redirection of the console messages to a log file or to syslog.

More information: [#343](https://github.com/creativeprojects/resticprofile/pull/343)

### Add "--stderr" to redirect console to stderr (for "cat" and "dump")

More information: https://github.com/creativeprojects/resticprofile/pull/353

### Add option to set working directory for restic backup

More information: https://github.com/creativeprojects/resticprofile/pull/354

### many bug fixes!

## Changelog
* [Add "--stderr" to redirect console to stderr (for "cat" and "dump") (#353)](https://github.com/creativeprojects/resticprofile/commit/1bbac2a)
* [Add documentation on how to ingest resticprofile stats into telegraf (#366)](https://github.com/creativeprojects/resticprofile/commit/0598026)
* [Add option to set working directory for restic backup (#354)](https://github.com/creativeprojects/resticprofile/commit/f9b5dac)
* [Fix group priority (#339)](https://github.com/creativeprojects/resticprofile/commit/00360a2)
* [Relative source followup: Evaluate symlinks (#355)](https://github.com/creativeprojects/resticprofile/commit/a05ad39)
* [Squashed commit of the following: (#259)](https://github.com/creativeprojects/resticprofile/commit/56f20af)
* [add "profile" flag as an alias for "name" (#357)](https://github.com/creativeprojects/resticprofile/commit/6b00c2c)
* [also search configuration from ~/.config/resticprofile on macOS (#370)](https://github.com/creativeprojects/resticprofile/commit/6da0317)
* [crond: add support for crontab file only (on any OS) (#289)](https://github.com/creativeprojects/resticprofile/commit/5739b13)
* [don't replace header value in stringifyHeaders (#327)](https://github.com/creativeprojects/resticprofile/commit/9719411)
* [drop-ins: move systemd drop-ins into unified schedule struct (#341)](https://github.com/creativeprojects/resticprofile/commit/2a7d5fc)
* [drop-ins: transparent timer drop-in support (#340)](https://github.com/creativeprojects/resticprofile/commit/7e03741)
* [env-file: Implement dotenv and {{env}} support (#323)](https://github.com/creativeprojects/resticprofile/commit/9b2bd8c)
* [fix remaining unit tests failing in Windows (#360)](https://github.com/creativeprojects/resticprofile/commit/16f5dc5)
* [fixed link to config reference / global section (#349)](https://github.com/creativeprojects/resticprofile/commit/3b8613e)
* [flags: added "RESTICPROFILE_{FLAGNAME}" env vars (#335)](https://github.com/creativeprojects/resticprofile/commit/5fad1cb)
* [log: allow controlling command output redirection (#343)](https://github.com/creativeprojects/resticprofile/commit/85d5afc)
* [restic: use repository-file if the repo flag contains a password (#336)](https://github.com/creativeprojects/resticprofile/commit/e5b17d3)
* [schedule: added unified schedule config struct (#333)](https://github.com/creativeprojects/resticprofile/commit/4255868)
* [syslog: local syslog and stdout redirection (#344)](https://github.com/creativeprojects/resticprofile/commit/667180e)
* [build with go 1.22 and macOs arm64 (#317)](https://github.com/creativeprojects/resticprofile/commit/e93874d)
* [update test to pass on Windows 11 (#359)](https://github.com/creativeprojects/resticprofile/commit/e6eed3d)
* [upgrade packages (CVE-2024-24786) (#347)](https://github.com/creativeprojects/resticprofile/commit/655e9c5)
* [upgrade packages (CVE-2024-6104)](https://github.com/creativeprojects/resticprofile/commit/08a01c4)



# v0.26.0 (2024-02-20)

## 🦆 Second time lucky February release 🚀 

🆕 A lot of pre-built binaries have been added on this release. This is to align with the list of pre-built binaries provided by `restic`

Otherwise this is mostly a maintenance release with a few fixes:
- fixes multiple backup profiles exporting prometheus files to `node_exporter`
- fix missing fields in the `show` command
- weekly docker image build (rebuilt under the same version tag, and as `latest`)
- restrict the `copy` command to a list of snapshots in the configuration

## Changelog
* [Add pre-built binary targets to release pipeline (#324)](https://github.com/creativeprojects/resticprofile/commit/2ab1f3a)
* [Add prometheus label to build info (#319)](https://github.com/creativeprojects/resticprofile/commit/233e4b8)
* [Change priority warning message (#310)](https://github.com/creativeprojects/resticprofile/commit/b1b03db)
* [Fix missing fields in show command (#315)](https://github.com/creativeprojects/resticprofile/commit/9eba431)
* [Merge pull request #312 from creativeprojects/nightly](https://github.com/creativeprojects/resticprofile/commit/a9273f8)
* [Restrict copy command to a list of snapshots (#291)](https://github.com/creativeprojects/resticprofile/commit/15004f9)
* [add separate goreleaser config for rebuilding the docker image only (#309)](https://github.com/creativeprojects/resticprofile/commit/1c076dc)
* [add snapshot build to docker hub](https://github.com/creativeprojects/resticprofile/commit/bbb3760)
* [publish docker images and manifest manually (#313)](https://github.com/creativeprojects/resticprofile/commit/8aebd28)



# v0.25.0 (2024-02-08)

## 💖 February release 💝

### ⚠️ Potential breaking change

The prometheus library used to send metrics to the proxy using `protobuf`.
By default it's now using the more widely used `text` format.

If you have any issue sending metrics to your proxy, you can revert to the previous behaviour by adding this option to your profile (it's **not** a global option)

```yaml
my_profile:
  prometheus-push-format: protobuf
```

More information about the different formats: https://prometheus.io/docs/instrumenting/exposition_formats/#exposition-formats


## New in this version
* fix for CVE-2023-48795
* new option `prometheus-push-format` with values `text` or `protobuf` (#281)
* new option to set log output in `global` section (#277)
* more control over the default systemd service files (#267)
* and bug fixes!

**Thanks to all our contributors for the good work!**

## Changelog
* [Add `prometheus-push-format` to allow selecting text format (#281)](https://github.com/creativeprojects/resticprofile/commit/765c2af)
* [Pass context to own commands and profile runner (#280)](https://github.com/creativeprojects/resticprofile/commit/ae9554a)
* [Remove rclone binary in make clean target (#283)](https://github.com/creativeprojects/resticprofile/commit/fccc05b)
* [Trying new configuration for CodeQL (#306)](https://github.com/creativeprojects/resticprofile/commit/0871d28)
* [Upgrade packages (#307)](https://github.com/creativeprojects/resticprofile/commit/1317f60)
* [chore: prep next release & allow deployment to fail on PR](https://github.com/creativeprojects/resticprofile/commit/63f8faf)
* [chore: various fixes (#285)](https://github.com/creativeprojects/resticprofile/commit/3d72803)
* [doc: add information about windows path in variables](https://github.com/creativeprojects/resticprofile/commit/952380f)
* [doc: add various missing information (#278)](https://github.com/creativeprojects/resticprofile/commit/f346571)
* [logging: allow to setup default log output in global (#277)](https://github.com/creativeprojects/resticprofile/commit/1b3292c)
* [macOS: create LaunchAgents folder if it doesn't exist](https://github.com/creativeprojects/resticprofile/commit/99484bf)
* [profile: support source with "-" (dash) prefix (#276)](https://github.com/creativeprojects/resticprofile/commit/37dcf84)
* [systemd drop-ins support, option to wait for network-online.target (#267)](https://github.com/creativeprojects/resticprofile/commit/a1b7840)



# v0.24.0 (2023-10-24)

## 🎃 October release 👻 

- upgrade dependencies to fix CVE-2023-3978, CVE-2023-39325 and CVE-2023-44487
- fix the broken documentation (some tabs were not accessible)
- can now stop the wait for a lock (restic or resticprofile lock). Before the fix the `CTRL-C` or other signal was ignored until the lock was acquired or timed out.
- resticprofile is now available on [scoop](https://scoop.sh/#/apps?q=resticprofile)! thanks @hgraeber 
- detect if the host is running on battery and cancel an action depending on how much battery is left - see [schedule section](https://creativeprojects.github.io/resticprofile/schedules/configuration/index.html#schedule-ignore-on-battery)
- bug fixes (see changelog)

## Changelog
* [Allow to interrupt the wait for a lock (#249)](https://github.com/creativeprojects/resticprofile/commit/ac99302)
* [Detect if running on battery (#235)](https://github.com/creativeprojects/resticprofile/commit/b08ac73)
* [Docs for installation from scoop (#268)](https://github.com/creativeprojects/resticprofile/commit/1e6d07a)
* [Fix doc shortcodes (#271)](https://github.com/creativeprojects/resticprofile/commit/0d73c2b)
* [Template: Reduce log level for `Getwd()` failures (#251)](https://github.com/creativeprojects/resticprofile/commit/f5e751c)
* [systemd: prevent paging in systemd schedules (#270)](https://github.com/creativeprojects/resticprofile/commit/e20973e)
* [upgrade packages - CVE-2023-3978 - CVE-2023-39325 - CVE-2023-44487](https://github.com/creativeprojects/resticprofile/commit/bcfaaa7)



# v0.23.0 (2023-08-11)

## 🌞 New summer edition 🌻 

This release mostly fixes a few regression from version 0.21.0 and adds a handful of new features:
* Allow to set `base-dir` and `source-dir` in a profile so you can start resticprofile from any current folder
* Allow to set `keep-tag`, `tag` & `group-by` as empty string
* Support for restic v0.16 new flags

Thanks to all the contributors for the great work 👍🏻 

## Changelog
* [Added {{ "data" | base64 }} & {{ "data" | hex }} (#213)](https://github.com/creativeprojects/resticprofile/commit/b0767ad)
* [Allow to set a base-dir inside the profile (#183) (#192)](https://github.com/creativeprojects/resticprofile/commit/b00c8a0)
* [Fix 194: Allow to set "keep-tag", “tag” & “group-by” as empty string (#220)](https://github.com/creativeprojects/resticprofile/commit/4b2ea31)
* [Fix 218: Args filter must not remove paths (#222)](https://github.com/creativeprojects/resticprofile/commit/10057d4)
* [Fix 223, 230: Escape args and absolute restic path for pwsh (#224)](https://github.com/creativeprojects/resticprofile/commit/0c8c985)
* [Fix 242: iexclude-file not converted to abs path (#243)](https://github.com/creativeprojects/resticprofile/commit/2bf0d28)
* [Fix links for versioned JSON schema files (#244)](https://github.com/creativeprojects/resticprofile/commit/6cbfdb3)
* [Fix schedule tests (#236)](https://github.com/creativeprojects/resticprofile/commit/fc66fab)
* [Fix: Do not add `--tag` for `tag: true` (#221)](https://github.com/creativeprojects/resticprofile/commit/304e418)
* [Restic: Add restic v16 release (#238)](https://github.com/creativeprojects/resticprofile/commit/82fccdb)
* [Restic: Fixed unit tests for restic v16 (#239)](https://github.com/creativeprojects/resticprofile/commit/3f4e0bf)
* [Retention: Align host filter with "backup" (#227)](https://github.com/creativeprojects/resticprofile/commit/ba9297a)
* [Schedule: Capture `os.Environ` on schedule creation (#212)](https://github.com/creativeprojects/resticprofile/commit/347501d)
* [Support `lock-wait` with `--lock-retry` in restic 0.16 (#240)](https://github.com/creativeprojects/resticprofile/commit/9c05157)
* [Support build when GOPATH is unset / fix mockery build warning (#234)](https://github.com/creativeprojects/resticprofile/commit/6cc332d)
* [Variables: Allow to escape "$" with "$$" (#216)](https://github.com/creativeprojects/resticprofile/commit/3476fbc)



# v0.22.0 (2023-05-06)

## ⚠️  Breaking change

The default value of the `job` tag on prometheus gateway push has changed from `command` to `profile.command`.
But don't worry: you can easily revert back to the original value by adding this option in your configuration:

```yaml
prometheus-push-job: "${COMMAND}"
```

## Fixes:
* Complicated scheduling on Windows was sometimes setting up a random delay before starting a job

## Changelog
* [New option to specify Prometheus Pushgateway job name (#193)](https://github.com/creativeprojects/resticprofile/commit/c00fc9a)
* [Templates: Add `map`/`splitR`/`contains`/`matches` (#197)](https://github.com/creativeprojects/resticprofile/commit/0c7405f)
* [Upgrade task scheduler library (#206)](https://github.com/creativeprojects/resticprofile/commit/7545bbf)


# v0.21.1 (2023-04-05)

## Bug fixes!

This small release fixes two regression bugs introduced in version `0.21.0`
- the `~` was no longer replaced by the user's home directory for some flag (`repository` and others...)
- environment variables were no longer replaced by their value for some flag (`repository` and others...)

Thanks @jkellerer for the quick fixes 😉 👍🏻 

## New features from `v0.21.0`
This release adds a verification of all the flags permitted by restic. Only the flags compatible with your version of restic will be generated (prior to this version, any flag like `unknown-flag` would end up on the restic command line as `--unknown-flag`.

Also this version generates a JSON schema: this is the configuration schema which can auto-complete options, and verify that your configuration is correct. It works with any compatible editor for the JSON, TOML and YAML configuration file format. Thanks @jkellerer for the awesome work on this 🎉 . [More information here](https://creativeprojects.github.io/resticprofile/configuration/jsonschema/index.html).

Other nice feature is the introduction of the `help` command which works for both all the **resticprofile** and **restic** commands and flags. Try it out!

## Changelog
* [Fix #187: homedir/env-vars in repo and other flags (#188)](https://github.com/creativeprojects/resticprofile/commit/487db0b)
* [Fix #189: completion of own commands (#190)](https://github.com/creativeprojects/resticprofile/commit/b78d1a0)
* [Updated clog to 0.13 (#191)](https://github.com/creativeprojects/resticprofile/commit/b6a98ef)



# v0.21.0 (2023-04-03)

### This is a great milestone for resticprofile 🥳 

This release adds a verification of all the flags permitted by restic. Only the flags compatible with your version of restic will be generated (prior to this version, any flag like `unknown-flag` would end up on the restic command line as `--unknown-flag`.

Also this version generates a JSON schema: this is the configuration schema which can auto-complete options, and verify that your configuration is correct. It works with any compatible editor for the JSON, TOML and YAML configuration file format. Thanks @jkellerer for the awesome work on this 🎉 . [More information here](https://creativeprojects.github.io/resticprofile/configuration/jsonschema/index.html).

Other nice feature is the introduction of the `help` command which works for both all the **resticprofile** and **restic** commands and flags. Try it out!

And as usual, a lot of bug fixes 😉 

## Changelog
* [Add variables `.OS` and `.Arch` to all templates (#181)](https://github.com/creativeprojects/resticprofile/commit/683bf78)
* [Allow config encoding in UTF16 and ISO88591](https://github.com/creativeprojects/resticprofile/commit/74b0c46)
* [Enhanced: Catch any --help & --dry-run (http only) (#178)](https://github.com/creativeprojects/resticprofile/commit/d6a51ad)
* [Feature: JSON schema (#103) & generated reference](https://github.com/creativeprojects/resticprofile/commit/3f5cdc6)
* [Fix #164: Failure on uppercase mixin names](https://github.com/creativeprojects/resticprofile/commit/1aad0fb)
* [Implement tempDir and log without locking (#168) (#174)](https://github.com/creativeprojects/resticprofile/commit/e65f9bc)
* [Make URL and header confidential in HTTP-hook (#175)](https://github.com/creativeprojects/resticprofile/commit/f2a9c04)
* [Restic: Add restic v15 release (#180)](https://github.com/creativeprojects/resticprofile/commit/e069b77)
* [add build tag to disable self-update (#184)](https://github.com/creativeprojects/resticprofile/commit/d2e789c)
* [add suppport for user_logged_on (#160)](https://github.com/creativeprojects/resticprofile/commit/8231087)
* [build with go 1.20](https://github.com/creativeprojects/resticprofile/commit/f782ef5)
* [upgrade packages CVE-2022-41721](https://github.com/creativeprojects/resticprofile/commit/bd3813a)



# v0.20.0 (2023-01-23)

Not too many new features in this release, but I wanted to build a new docker image with the new restic 0.15.0

## Improvements
- Adds shell command hooks to the following additional commands: `dump`, `find`, `ls`, `mount`, `restore`, `snapshots`, `stats` and `tag`.
- Docker image now contains ssh (to allow `sftp` repository), curl, tzdata and ca-certificates

## Changelog
* [Merge pull request #154 from jkellerer/ft-151](https://github.com/creativeprojects/resticprofile/commit/51634c8)
* [New docker image (#161)](https://github.com/creativeprojects/resticprofile/commit/00ab266)
* [Run-hooks for all non-conflicting commands (#151)](https://github.com/creativeprojects/resticprofile/commit/fce7165)
* [add QEMU for github agent to build an arm64 image](https://github.com/creativeprojects/resticprofile/commit/7439620)
* [dry-run should not send web hooks #157](https://github.com/creativeprojects/resticprofile/commit/7c2f806)



# v0.19.0 (2022-11-11)

## New version 0.19.0 of resticprofile!
With:
- New help system showing all flags from resticprofile **and restic**. Just type `resticprofile help backup` and see (thanks @jkellerer)
- `copy` command now has `run-before`, `run-after`, `run-after-fail` and `run-finally` targets. Also more targets are now available for `backup` and for a profile. [See the documentation](https://creativeprojects.github.io/resticprofile/configuration/run_hooks/index.html).
- groups of profiles can keep running after a profile failed (this is a global settings)
- Prevent your system from idle sleeping (Windows, macOS, and unix type OS using `systemd`)
- May contain nuts

## Changelog
* [Add --help to own commands (e.g. generate --help) (#139)](https://github.com/creativeprojects/resticprofile/commit/c360880)
* [Add last backup time to prometheus metrics (#132)](https://github.com/creativeprojects/resticprofile/commit/1edf995)
* [Config: Add run before/after/fail to more restic commands than backup (#138)](https://github.com/creativeprojects/resticprofile/commit/8a70e7b)
* [Continue profile in group on error (#149)](https://github.com/creativeprojects/resticprofile/commit/419e66c)
* [Fixes zsh completion script (use of outdated CLI) (#150)](https://github.com/creativeprojects/resticprofile/commit/a880b44)
* [Prevent system from idle sleeping (#140)](https://github.com/creativeprojects/resticprofile/commit/49c4920)
* [add devcontainer config](https://github.com/creativeprojects/resticprofile/commit/d4c1032)
* [chore: upgrade packages](https://github.com/creativeprojects/resticprofile/commit/0c43c2c)



# v0.18.0 (2022-08-29)

Following the release of the long awaited restic 0.14.0, here’s a new resticprofile!

A few big features were introduced in this version:

- HTTP hooks using a similar syntax to command hooks but sending HTTP messages to monitoring platforms
- Better support of the new-ish `copy` command
- Addition of mixins on configuration file v2 (in preview)
- Ability to choose your preferred shell on Windows (cmd, powershell or bash)
- Send resticprofile logs to a syslog server
- Add armv8 (arm64) CPU target to docker images
- Upgrade to restic 0.14.0 in docker image
- Add rclone to docker image
- Also search for a configuration file in the resticprofile program folder on Windows (to be used in portable mode)

## Changelog
* [Add http hooks (#114)](https://github.com/creativeprojects/resticprofile/commit/7ba20ad)
* [Add support for syslog (#127)](https://github.com/creativeprojects/resticprofile/commit/fb62876)
* [Allow to choose `shell` in global config (#112)](https://github.com/creativeprojects/resticprofile/commit/a5c1147)
* [ConfigV2: Mixins (#115)](https://github.com/creativeprojects/resticprofile/commit/d492d7d)
* [ConfigV2: Replace list params - fix #108 (#113)](https://github.com/creativeprojects/resticprofile/commit/ca9a418)
* [Initialise copy repository using --copy-chunker-params (#117)](https://github.com/creativeprojects/resticprofile/commit/2ed3045)
* [Mixins: List merging & inheritance update (#121)](https://github.com/creativeprojects/resticprofile/commit/43ce0be)
* [add binary dir in path and .BinaryDir template var (#134)](https://github.com/creativeprojects/resticprofile/commit/1b8774c)
* [add goreleaser config to also build arm64 images](https://github.com/creativeprojects/resticprofile/commit/eb70a8a)
* [add rclone to docker image #131](https://github.com/creativeprojects/resticprofile/commit/35d07ee)
* [upgrade packages (CVE-2022-28948)](https://github.com/creativeprojects/resticprofile/commit/0381942)




# v0.17.0 (2022-05-16)

## Here it is!

It's been a while since we released a new version 😞 
The resticprofile team has been busy preparing some really cool new features:

- simplify the command line by allowing the use of [profile].[command] - Thanks @Syphdias  for the PR #89 
- shell completion (for `bash`, can also be used by `zsh` with bash compatibility) - Thanks @jkellerer  for the PR #90 
- run a shell command to use as a stdin input (like `mysqldump`) - Thanks @jkellerer for the PR #98 
- run a shell command in the background when non fatal errors are detected from restic - Thanks @jkellerer for the PR #99 

and a lot of bug fixes 👍🏻 

## Changelog
* [Add run flag to to be able to have profile and job name as one argument (#89)](https://github.com/creativeprojects/resticprofile/commit/83e423e)
* [Added "generate" command to create resources (#110)](https://github.com/creativeprojects/resticprofile/commit/20a3b81)
* [Added missing cmds that filter by host, tag & path](https://github.com/creativeprojects/resticprofile/commit/1d2b214)
* [Feature: Take command output as stdin for restic (#98)](https://github.com/creativeprojects/resticprofile/commit/8cef157)
* [Fix and unit tests for #91](https://github.com/creativeprojects/resticprofile/commit/fe336ff)
* [Fix config includes when any pattern has no match (#95)](https://github.com/creativeprojects/resticprofile/commit/ad511c4)
* [Implemented stream-error callbacks (#99)](https://github.com/creativeprojects/resticprofile/commit/3e856ce)
* [Merge pull request #101 from jkellerer/fix-common-args](https://github.com/creativeprojects/resticprofile/commit/3688b90)
* [Merge pull request #94 from jkellerer/ft-add-missing-commands](https://github.com/creativeprojects/resticprofile/commit/4f13930)
* [Only pass common CLI args to command hooks [#100]](https://github.com/creativeprojects/resticprofile/commit/a72f635)
* [Shell completion (#90)](https://github.com/creativeprojects/resticprofile/commit/7553a24)
* [display a neat stack trace on panic](https://github.com/creativeprojects/resticprofile/commit/ea52e93)
* [fix for profiles command not showing inherited commands fixes second part of #97](https://github.com/creativeprojects/resticprofile/commit/ff13660)
* [remove "includes" section from profiles in "profiles" command](https://github.com/creativeprojects/resticprofile/commit/69dd965)



# v0.16.1 (2022-01-30)

New maintenance version, with bug fixes:
- fix multiplication of arguments when commands are retried
- fix status file telling the backup was successful when a warning happened (file/dir not found)

## Changelog
* [Fix args are multiplied when commands are retried (#84)](https://github.com/creativeprojects/resticprofile/commit/e18906b)
* [Fix for #88 with unit tests](https://github.com/creativeprojects/resticprofile/commit/358f060)
* [don't inherit profile description](https://github.com/creativeprojects/resticprofile/commit/b011437)
* [prepare for future versions of the configuration](https://github.com/creativeprojects/resticprofile/commit/16ad4b1)
* [show schedules separately](https://github.com/creativeprojects/resticprofile/commit/da8c255)
* [simple implementation of a config file v2](https://github.com/creativeprojects/resticprofile/commit/358a5db)
* [update packages](https://github.com/creativeprojects/resticprofile/commit/e2bd2cf)



# v0.16.0 (2021-10-18)

This release adds a few new features:

- support for splitting configuration into multiple files
- support for `run-finally` that runs shell commands every time after restic
- ability to define your own systemd unit and timer files (from go templates)
- support for the restic `copy` command
- fix for some cosmetic issues with `crond` scheduler

## Changelog

* [Add support for copy command (#73)](https://github.com/creativeprojects/resticprofile/commit/1fa7d48)
* [Change codecov uploader to use GitHub Action v2 (#79)](https://github.com/creativeprojects/resticprofile/commit/271c128)
* [Optional: Allow disabling path in retention with 'false' (#67)](https://github.com/creativeprojects/resticprofile/commit/85fcc20)
* [Scheduler refactoring (#76)](https://github.com/creativeprojects/resticprofile/commit/8773899)
* [Support "run-finally" in backup-section & profile (#70)](https://github.com/creativeprojects/resticprofile/commit/48c822c)
* [Supporting config `includes` (e.g. `profiles.d`) (#65)](https://github.com/creativeprojects/resticprofile/commit/ac13d06)
* [Systemd template (#75)](https://github.com/creativeprojects/resticprofile/commit/245a440)
* [add tests on crontab](https://github.com/creativeprojects/resticprofile/commit/122620e)
* [crontab RemoveJob returns error if the entry was not found](https://github.com/creativeprojects/resticprofile/commit/0dad5a8)
* [upgrade dependencies](https://github.com/creativeprojects/resticprofile/commit/ff2fc1f)


## Docker images

- `docker pull creativeprojects/resticprofile:0.16.0`


# v0.15.0 (2021-08-29)

## ⚠️ Important

Version 0.15 fixed some issues with escaping parameters to the restic command line. **If you've used any of these characters in file or directory names in your configuration, please make sure your backup is still working as expected:** `space`, `*`, `?`, `\`, `[`, `'`, `"`.

If for some reason the fix broke your configuration, there's a new flag `legacy-arguments` that you can activate in the `global` section to revert back to the broken implementation:

```yaml
global:
    legacy-arguments: true
```

## New features in 0.15:
- add `.Hostname` in configuration template (#55)
- add description field in profile section
- add support for prometheus file export and push gateway
- hide confidential values in output (#58)
- add `.TmpDir` variable to configuration template (#62)

## This version also includes fixes for:
- resolving special paths starting with a `~` (unixes only)
- warn when the restic binary was not found at the specified location (but still found at a different location)
- resolve glob expressions in backup sources (#63)

## Changelog

* [Add support for prometheus export and push (#57)](https://github.com/creativeprojects/resticprofile/commit/abf0b00)
* [Added "{{.TmpDir}}" to TemplateData (#62)](https://github.com/creativeprojects/resticprofile/commit/37a69d7)
* [Feature: Hide confidential values in output (#58)](https://github.com/creativeprojects/resticprofile/commit/8218e70)
* [Resolve glob expressions in backup sources (#63)](https://github.com/creativeprojects/resticprofile/commit/0be5d5f)
* [Shell escape (#60)](https://github.com/creativeprojects/resticprofile/commit/8cc2574)
* [Update non-confidential values to support shell.Args (#68)](https://github.com/creativeprojects/resticprofile/commit/34f693c)
* [add Hostname pre-defined variable to template resolver (#55)](https://github.com/creativeprojects/resticprofile/commit/0841959)
* [add description field in profile section](https://github.com/creativeprojects/resticprofile/commit/f58116b)
* [build with go 1.17](https://github.com/creativeprojects/resticprofile/commit/0f9b21e)
* [don't send status summary in dry-run](https://github.com/creativeprojects/resticprofile/commit/16308e9)
* [show description in output of profiles command](https://github.com/creativeprojects/resticprofile/commit/88052c6)


## Docker images

- `docker pull creativeprojects/resticprofile:0.15.0`


# v0.14.1 (2021-06-23)

## Bug fix release

This release changes the way the restic binary is searched:
- restic binary path in `global` configuration can now contain the `~` character like `~restic/bin/restic`
- if you specified a path in the `global` configuration and it cannot find the file, a warning will be displayed and resticprofile will keep trying to find a suitable binary

## Changelog

* [use shell to resolve special paths with ~](https://github.com/creativeprojects/resticprofile/commit/4686dd2)
* [warning when the restic-binary was not found](https://github.com/creativeprojects/resticprofile/commit/823304a)


## Docker images

- `docker pull creativeprojects/resticprofile:0.14.1`


# v0.14.0 (2021-05-11)

## Release 0.14.0

- New locking/unlocking features. Thanks `jkellerer` for the PR

## Changelog

* [Added resticprofile flags --no-lock & --lock-wait (#33)](https://github.com/creativeprojects/resticprofile/commit/331b710)
* [Summary from plain output when not run in terminal (#48)](https://github.com/creativeprojects/resticprofile/commit/7fc5b3a)
* [Updated contrib script systemd/send-error.sh (#49)](https://github.com/creativeprojects/resticprofile/commit/b7edeb5)
* [add macOS arm64 target to install.sh script](https://github.com/creativeprojects/resticprofile/commit/fa42cf2)
* [add token as an environment variable](https://github.com/creativeprojects/resticprofile/commit/a87f76d)


## Docker images

- `docker pull creativeprojects/resticprofile:latest`
- `docker pull creativeprojects/resticprofile:0.14.0`


# v0.13.2 (2021-04-20)

This version fixes a defect where extended status wasn't returning the extended information on Windows.

## Changelog

* [add Homebrew Tap (#45)](https://github.com/creativeprojects/resticprofile/commit/cab8909)
* [add github token in config](https://github.com/creativeprojects/resticprofile/commit/794f404)
* [fix test too slow on build agent](https://github.com/creativeprojects/resticprofile/commit/5e06dbd)
* [fix windows bogus prefix (#47)](https://github.com/creativeprojects/resticprofile/commit/2f79a46)


## Docker images

- `docker pull creativeprojects/resticprofile:latest`
- `docker pull creativeprojects/resticprofile:0.13.2`


# v0.13.1 (2021-03-26)

Bug fix:
- regression from v0.13.0: a message was sent to stderr when `initialize` parameter was set and the repository already exists

## Changelog

* [Increase test coverage (#40)](https://github.com/creativeprojects/resticprofile/commit/02e6414)
* [fix #41: a message  was sent to stderr when parameter initialize=true and repo exists](https://github.com/creativeprojects/resticprofile/commit/d64dc4f)


## Docker images

- `docker pull creativeprojects/resticprofile:latest`
- `docker pull creativeprojects/resticprofile:0.13.1`


# v0.13.0 (2021-03-24)

This version adds two new features:
- parameter `no-error-on-warning` to consider a backup successful when restic produced a new snapshot but some files were missing (https://github.com/creativeprojects/resticprofile/discussions/38)
- resticprofile now catches the error output (stderr) to be written in the status file, also makes the environment variable `RESTIC_STDERR` available to the targets `run-after-fail`.

## Changelog

* [add RESTIC_STDERR env variable to run-after-fail](https://github.com/creativeprojects/resticprofile/commit/8968e33)
* [quick implementation of ignore warning](https://github.com/creativeprojects/resticprofile/commit/6b7dea4)
* [quick mock to do some testing with a fake restic](https://github.com/creativeprojects/resticprofile/commit/833d24d)
* [returns stderr output in the status file](https://github.com/creativeprojects/resticprofile/commit/3c54cc6)


## Docker images

- `docker pull creativeprojects/resticprofile:latest`
- `docker pull creativeprojects/resticprofile:0.13.0`


# v0.12.0 (2021-03-18)

This release mainly brings 2 new features and a few enhancements:

- add support for `--all` in `status`, `schedule`, `unschedule` commands
- add backup statistics to the status file (via a new `extended-status` flag)

## Changelog

* [Add backup statistics in status file (#36)](https://github.com/creativeprojects/resticprofile/commit/237a87f)
* [Added fail env variable ERROR_COMMANDLINE (#32)](https://github.com/creativeprojects/resticprofile/commit/6622a39)
* [Added support for --all to status & (un)schedule (#31)](https://github.com/creativeprojects/resticprofile/commit/0887354)
* [Enhanced "unschedule" to remove all possible jobs (#28)](https://github.com/creativeprojects/resticprofile/commit/16e6c19)
* [add profile name when running status --all](https://github.com/creativeprojects/resticprofile/commit/47489a2)
* [fix an issue where status --all was stopping at the first profile with no schedule](https://github.com/creativeprojects/resticprofile/commit/01ae05a)
* [update goreleaser config to v0.154](https://github.com/creativeprojects/resticprofile/commit/f50131b)
* [upgrade packages](https://github.com/creativeprojects/resticprofile/commit/c700d60)
* [upgrade self-update library](https://github.com/creativeprojects/resticprofile/commit/4569f0d)
* [use go 1.16](https://github.com/creativeprojects/resticprofile/commit/7eb663d)


## Docker images

- `docker pull creativeprojects/resticprofile:latest`
- `docker pull creativeprojects/resticprofile:0.12.0`


# v0.11.0 (2021-01-20)

## Highlights:

- detect systemd using `systemctl`: https://github.com/creativeprojects/resticprofile/issues/25
- add **experimental** support for scheduling tasks with **crond**
- add support for background tasks on Mac OS via a new `schedule-priority` parameter
- add support for scheduling `forget` and `prune` commands

## Deprecation

Scheduling in the `retention` section is now deprecated, please use the `forget` section instead (https://github.com/creativeprojects/resticprofile/issues/23)

## Changelog

* [Add background types and low priorityIO to launchd plist (#19)](https://github.com/creativeprojects/resticprofile/commit/be72758)
* [Add support for scheduling forget (#26)](https://github.com/creativeprojects/resticprofile/commit/d0de636)
* [Added prune as supported, schedulable command (#24)](https://github.com/creativeprojects/resticprofile/commit/ec02310)
* [Move CI build from Travis CI to GitHub Actions](https://github.com/creativeprojects/resticprofile/commit/8f48acf)
* [Refactoring schedule package (#30)](https://github.com/creativeprojects/resticprofile/commit/2e1564a)
* [Schedule priority (#29)](https://github.com/creativeprojects/resticprofile/commit/8fd7a2e)
* [add crond support to unix targets (except macOS)](https://github.com/creativeprojects/resticprofile/commit/b8c3e6a)
* [add deprecation notice for schedule on retention](https://github.com/creativeprojects/resticprofile/commit/3e93f45)
* [add docker build to goreleaser](https://github.com/creativeprojects/resticprofile/commit/4b815f2)
* [add pre-release parameter to self-update](https://github.com/creativeprojects/resticprofile/commit/5830a34)
* [add schedule-priority parameter](https://github.com/creativeprojects/resticprofile/commit/a8939ed)
* [add working directory to crontab line](https://github.com/creativeprojects/resticprofile/commit/641dc91)
* [allow verbose flag after the command](https://github.com/creativeprojects/resticprofile/commit/08141a3)
* [display systemd timer status in a nicer way also display log >= "warning" instead of "err"](https://github.com/creativeprojects/resticprofile/commit/2a6b0c0)
* [move other-flags before the sub-sections](https://github.com/creativeprojects/resticprofile/commit/a19f124)
* [refactor show command to remove empty lines](https://github.com/creativeprojects/resticprofile/commit/967c4a8)
* [search for systemctl instead of systemd](https://github.com/creativeprojects/resticprofile/commit/6544052)
* [upgrade clog package](https://github.com/creativeprojects/resticprofile/commit/c406e80)


## Docker images

- `docker pull creativeprojects/resticprofile:latest`
- `docker pull creativeprojects/resticprofile:0.11.0`


# v0.10.1 (2020-11-17)

This update changes the way systemd units are generated:
These used to be of type `oneshot` but it means they can be started more than once.
They have been changed to `notify` which is like `simple` but resticprofile is notifying systemd that the schedule has started and stopped.

To change the type on your existing schedules, you'll need to run the commands `unschedule` and `schedule` again.

## Changelog

* [change systemd unit to "notify"](https://github.com/creativeprojects/resticprofile/commit/ba711f2)
* [update packages verify self-update binary using checksums.txt file](https://github.com/creativeprojects/resticprofile/commit/9f2edb8)
* [update restic to 0.11.0 in docker image](https://github.com/creativeprojects/resticprofile/commit/cff661b)




# v0.10.0 (2020-11-12)

New resticprofile version with bug fixes and a big new feature:
- fix --exclude and --iexclude parameters in unix environment not properly escaping `?` and `*` characters
- don't escape ` ` (space) if already escaped
- self-update now works on ARM CPUs (like raspberry pi)
- configuration files can embed Go templates for modularity

I think it is now feature complete for a version 1.0 😄 

## Changelog

* [Escape `exclude` globs passed to `/sh`](https://github.com/creativeprojects/resticprofile/commit/e2047e0)
* [add proper character escaping: - count the number of \ in front - do not escape it again if it was already escaped](https://github.com/creativeprojects/resticprofile/commit/be97fae)
* [change self-updater library to work with ARM cpus](https://github.com/creativeprojects/resticprofile/commit/b30cbe5)
* [detect arm version from runtime internal register](https://github.com/creativeprojects/resticprofile/commit/8d95c1e)
* [new version of the updater](https://github.com/creativeprojects/resticprofile/commit/c9b0c34)
* [put armv7 target back](https://github.com/creativeprojects/resticprofile/commit/604da32)
* [squash merge of branch config-single-template: add templating and variable expansion](https://github.com/creativeprojects/resticprofile/commit/836cbae)
* [upgrade clog package](https://github.com/creativeprojects/resticprofile/commit/f177173)
* [upgrade selfupdate package](https://github.com/creativeprojects/resticprofile/commit/48cf85e)




# v0.9.2 (2020-11-02)

A few minor features and bug fixes:
- add version command
- add force-inactive-lock flag in profiles

## Changelog

* [Add "option" in example configurations in README and azure.conf fixes https://github.com/creativeprojects/resticprofile/issues/13](https://github.com/creativeprojects/resticprofile/commit/ca6ff07)
* [Add version command](https://github.com/creativeprojects/resticprofile/commit/06f98e9)
* [add tests on set and get PID from lock file](https://github.com/creativeprojects/resticprofile/commit/f3aa7b9)
* [remove openbsd from goreleaser library github.com/shirou/gopsutil cannot be compiled for openbsd](https://github.com/creativeprojects/resticprofile/commit/cd10c9d)
* [squash merge of branch pid: should fix https://github.com/creativeprojects/resticprofile/issues/14](https://github.com/creativeprojects/resticprofile/commit/8c07f35)
* [use go 1.15 for ci builds](https://github.com/creativeprojects/resticprofile/commit/618be92)
* [write down child PIDs in lock file](https://github.com/creativeprojects/resticprofile/commit/94f0ef1)




# v0.9.1 (2020-08-03)

Two new features in this release:
- add a few environment variables when running scripts (`run-before`, `run-after`, `run-after-fail`)
- add a status file generated after running a profile (to send to some monitoring software)

## Changelog

* [Merge branch run-after-fail](https://github.com/creativeprojects/resticprofile/commit/58ecde0)
* [merge from branch status-file](https://github.com/creativeprojects/resticprofile/commit/05afb52)




# v0.9.0 (2020-07-29)

A few new big features in this release:
- add `--dry-run` flag
- redirect console to a file with `--log` (for running in the background)
- generate random keys with `generate-random` command
- schedule/unschedule profiles automatically with command `schedule`/`unschedule` (also `status` to check a scheduled task)

## Changelog

* [add SUDO_USER to systemd environment](https://github.com/creativeprojects/resticprofile/commit/c0c067e)
* [add dry-run flag](https://github.com/creativeprojects/resticprofile/commit/67e24a4)
* [add file logger](https://github.com/creativeprojects/resticprofile/commit/b31c4cf)
* [add journalctl to status output](https://github.com/creativeprojects/resticprofile/commit/ef87550)
* [add schedule documentation to README](https://github.com/creativeprojects/resticprofile/commit/ba9d132)
* [add schedule parameter in config](https://github.com/creativeprojects/resticprofile/commit/db4fd20)
* [add windows task scheduler](https://github.com/creativeprojects/resticprofile/commit/05d0308)
* [create plist file for darwin](https://github.com/creativeprojects/resticprofile/commit/2a03874)
* [create system job on darwin](https://github.com/creativeprojects/resticprofile/commit/b1c12ab)
* [full implementation of systemd user unit](https://github.com/creativeprojects/resticprofile/commit/c73ccb3)
* [generate random keys](https://github.com/creativeprojects/resticprofile/commit/fce69a1)
* [generate schedule for darwin](https://github.com/creativeprojects/resticprofile/commit/5ddecf0)
* [implement schedule-log](https://github.com/creativeprojects/resticprofile/commit/4647e6e)
* [move logger into external package](https://github.com/creativeprojects/resticprofile/commit/631038d)
* [only ask for the user password once](https://github.com/creativeprojects/resticprofile/commit/43e7f81)
* [redirect terminal when elevated mode](https://github.com/creativeprojects/resticprofile/commit/af04717)
* [send message from elevated process to parent process via some simple http calls](https://github.com/creativeprojects/resticprofile/commit/b049747)
* [send terminal output remotely](https://github.com/creativeprojects/resticprofile/commit/106ea1b)
* [sudo trick in the example configuration](https://github.com/creativeprojects/resticprofile/commit/fbac6f4)
* [system daemon with launchd](https://github.com/creativeprojects/resticprofile/commit/0465a19)
* [systemd user/system](https://github.com/creativeprojects/resticprofile/commit/8538fce)




# v0.8.3 (2020-07-13)

Bug fixing release:
- restic flags were not generated for commands `forget` and `mount` https://github.com/creativeprojects/resticprofile/issues/9

## Changelog

* [add restic flags for forget and mount commands](https://github.com/creativeprojects/resticprofile/commit/4e08c7b)
* [add test to verify forget flags are loaded for all configuration types](https://github.com/creativeprojects/resticprofile/commit/74d36fe)
* [version 0.8.3](https://github.com/creativeprojects/resticprofile/commit/4b1d937)




# v0.8.2 (2020-07-09)

This is a bug fixing release:
- in the configuration file, some strings containing a comma were split into an array of strings. This is now fixed.

## Changelog

* [add table of contents](https://github.com/creativeprojects/resticprofile/commit/7a7836f)
* [build 0.8.2](https://github.com/creativeprojects/resticprofile/commit/0bf1400)
* [document new features in README](https://github.com/creativeprojects/resticprofile/commit/1ecc090)
* [remove armv7 packages from goreleaser](https://github.com/creativeprojects/resticprofile/commit/e2f1d8e)
* [remove default decode hooks (which can do funny things with tags)](https://github.com/creativeprojects/resticprofile/commit/3183207)




# v0.8.1 (2020-07-02)

This is mostly a bug fixing release:
- allow for spaces in directories and files (these were ignored before and were messing up the command line)
- add a `show` command to see the profile details (mostly for debugging really)
- add a `--format` flag to specify the type of configuration file format (if you want to use a different extension)

## Changelog

* [add --format flag to specify the config format](https://github.com/creativeprojects/resticprofile/commit/d2ebc7b)
* [add command 'show' to display profile properties](https://github.com/creativeprojects/resticprofile/commit/85a5d59)
* [add show command to help](https://github.com/creativeprojects/resticprofile/commit/b3349bc)
* [allow spaces in directories and files](https://github.com/creativeprojects/resticprofile/commit/3ac1afa)
* [change expected structure in unit tests](https://github.com/creativeprojects/resticprofile/commit/52c74cd)
* [display correct XDG directories in error message](https://github.com/creativeprojects/resticprofile/commit/1006fcf)
* [update dependencies](https://github.com/creativeprojects/resticprofile/commit/98706fb)




# v0.8.0 (2020-06-29)

This version is introducing a few new features:
- experimental support for HCL configuration files
- a configuration file with no extension is searched using all supported file extensions: `-c profiles` would either load profiles.conf, profiles.yaml, profiles.toml, profiles.json or profiles.hcl
- a new global parameter to check available memory before starting a profile (default is 100MB)

## Changelog

* [Initial support for HCL configuration file](https://github.com/creativeprojects/resticprofile/commit/d74a5d6)
* [Refactoring y/n question](https://github.com/creativeprojects/resticprofile/commit/1e645e2)
* [add OS & ARCH to panic data](https://github.com/creativeprojects/resticprofile/commit/4aa7bcb)
* [add missing build info to docker image](https://github.com/creativeprojects/resticprofile/commit/5e417b3)
* [add note with self-updating on linux/arm](https://github.com/creativeprojects/resticprofile/commit/254377b)
* [add safeguard to prevent running restic when memory is running low: "min-memory" in "global"](https://github.com/creativeprojects/resticprofile/commit/b1a7674)
* [add section on where to locate configuration file](https://github.com/creativeprojects/resticprofile/commit/e10a2ab)
* [bump version](https://github.com/creativeprojects/resticprofile/commit/8c548bf)
* [example of configuration in HCL](https://github.com/creativeprojects/resticprofile/commit/742f6dd)
* [refactor configuration as a dependency as opposed to a global](https://github.com/creativeprojects/resticprofile/commit/69a4385)
* [search for configuration file without an extension resticprofile will load the first file with a valid extension (conf, yaml, toml, json, hcl)](https://github.com/creativeprojects/resticprofile/commit/8efe9c4)
* [update xdg paths in documentation](https://github.com/creativeprojects/resticprofile/commit/428fdb8)




# v0.7.1 (2020-06-25)

This a maintenance release:
- add a new parameter `run-after-fail` which is running the shell commands after any kind of failure (during other commands or during restic execution)
- a minor breaking change if you use the repository auto-initialization, it will now run **after** the `run-before` scripts. It makes more sense this way (in case you mount your backup disks in `run-before` for example)

## Changelog

* [add raspberry pi to supported platforms](https://github.com/creativeprojects/resticprofile/commit/2321bf9)
* [add run-after-fail parameter to run shell commands after a restic command failed. the auto initialization of a repository now happens after the "run-before" scripts (in version 0.7.0 it was happening before)](https://github.com/creativeprojects/resticprofile/commit/1efeb81)
* [detect panic and display a nice message asking the user to post details on github](https://github.com/creativeprojects/resticprofile/commit/80ddc06)
* [fix self-update not working on windows](https://github.com/creativeprojects/resticprofile/commit/9ca91d0)
* [run self-update command with no configuration file](https://github.com/creativeprojects/resticprofile/commit/d32c172)




# v0.7.0 (2020-06-24)

This is a maintenance release:
- fixing a defect when starting a backup with no command definition for backup
- adding two new options to run scripts before and after a profile (not just backup like before)
- implementation of a new module to create systemd units. More to come on future releases

## Changelog

* [Bump up version](https://github.com/creativeprojects/resticprofile/commit/a8ff142)
* [Fix update confirmation message](https://github.com/creativeprojects/resticprofile/commit/b499f76)
* [Refactoring resticprofile commands](https://github.com/creativeprojects/resticprofile/commit/d71f45f)
* [add goreleaser target to build for raspberry pi](https://github.com/creativeprojects/resticprofile/commit/ab527c0)
* [initial support for systemd](https://github.com/creativeprojects/resticprofile/commit/744ac5b)
* [run shell commands before and after a profile. In previous versions, you could only run commands before and after a backup](https://github.com/creativeprojects/resticprofile/commit/88cb74f)
* [stop creating empty configuration file at XDG config location](https://github.com/creativeprojects/resticprofile/commit/c4c8d9a)




# v0.6.1 (2020-04-22)



## Changelog

* [Activate self-update flag](https://github.com/creativeprojects/resticprofile/commit/b17b23c)
* [Add .tar.gz binary for mac os](https://github.com/creativeprojects/resticprofile/commit/713e429)
* [Build tar.gz for windows so we can download it from git bash](https://github.com/creativeprojects/resticprofile/commit/4c4c0d8)
* [Clean up lock file after pressing CTRL-C](https://github.com/creativeprojects/resticprofile/commit/a2fe661)
* [Fix nil pointer panic when retention not defined](https://github.com/creativeprojects/resticprofile/commit/2a76a04)




# v0.6.0 (2020-04-07)

Complete rewrite of resticprofile in go:

* I tried python and I wasn't particularly impressed
* I tried go and I loved it (also that's what all the cool kids do nowadays)

The default configuration `profiles.conf` is still expected to be in TOML format. But now, resticprofile also supports YAML and JSON file format. Simply feed a `.toml`, `.json` or a `.yaml` file to select the desired format.

The configuration file from the previous versions remains unchanged. There are two additions to it:

* `priority` flags: acts like `nice` on unixes and now available on **all** platforms
* `lock` flag to avoid running two profiles at the same time (this is a *local* lock)


# v0.5.2 (2019-09-26)

Accept a repository from the environment instead of the configuration file

# v0.5.1 (2019-08-12)

- Add a timestamp before each message
- Use colours more readable on light console
- Add a flag to disable all ansi characters (to redirect the output to a file)


# v0.5.0 (2019-06-26)

Adds support for stdin file stream

# v0.4.1 Beta (2019-06-26)

- Adding 'mount' command support
- Allow chained inheritance of environment variables
- Allow inheritance on 'retention' section


# v0.4.0 Beta (2019-06-25)

- Removing ugly KeyboardInterrupt message when you hit CTRL-C
- Allow the host flag to be a boolean (if set to yes the current hostname will be used)
- Also search for restic binary under Windows (like the chocolatey bin folder)
- Adding two options to check repository before and/or after a backup: 'check-before' and 'check-after'
- Run shell commands before and after a backup: 'run-before' and 'run-after'


# v0.3.0 Beta (2019-06-24)

- Added "groups": ability to define groups of profiles to run at once


# v0.2.0 Beta (2019-06-24)

- Fixed compatibility with Windows
- Fixed defect in profile inheritance


# v0.1.0 Beta (2019-06-24)

First 'usable' version of resticprofile!

**Please note the groups are not working yet**
(groups can run multiple configuration in one command)


