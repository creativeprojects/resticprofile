![Build](https://github.com/creativeprojects/resticprofile/workflows/Build/badge.svg)
![Run tests on FreeBSD, OpenBSD, NetBSD](https://github.com/creativeprojects/resticprofile/actions/workflows/BSD-tests.yml/badge.svg)
[![codecov](https://codecov.io/gh/creativeprojects/resticprofile/branch/master/graph/badge.svg?token=cUozgF9j4I)](https://codecov.io/gh/creativeprojects/resticprofile)

# resticprofile
Configuration profiles manager for [restic backup](https://restic.net/)

**resticprofile** is the missing link between a configuration file and restic backup command line.

## Features

Here is a non-exhaustive list of what resticprofile offers:

* **Profiles**
    * No need to remember command parameters and environment variables
    * Create multiple profiles in one configuration file
    * Profiles can inherit options from other profiles
    * Create groups of profiles to run sequentially
    * Easily schedule backups, retentions, and checks (supports *systemd*, *crond*, *launchd*, and *Windows Task Scheduler*)
    * Use template syntax in your configuration file
    * **[new for v0.29.0]** Schedule a group of profiles (configuration `v2` only)
* **Automation**
    * Run the forget command before or after a backup (in a section called *retention*)
    * Check a repository before or after a backup
    * Run shell commands before or after running a profile, useful for mounting and unmounting backup disks
    * Run a shell command if an error occurs
    * Send a backup stream via _stdin_
    * Start restic at different priorities (Priority Class in Windows, *nice* in Unix, and/or _ionice_ in Linux)
    * Automatically clear stale locks
* **Monitoring**
    * Generate a simple status file for monitoring software to ensure backups are running smoothly
    * Export a prometheus file after a backup or send the report to a push gateway
    * Run shell commands in the background when non-fatal errors are detected
    * Send messages to HTTP hooks before, after a successful or failed job (backup, forget, check, prune, copy)
    * Send resticprofile logs to a syslog server
* **Checks**
    * Check for enough memory before starting a backup
    * Avoid scheduling a job when the system is on battery
* **Misc**
    * Generate cryptographically secure random keys for a restic key file
    * Automatically initialize the secondary repository using the `copy-chunker-params` flag
    * Prevent the system from idle sleeping
    * View help for both restic and resticprofile via the `help` command or `-h` flag

## Configuration files

The configuration file accepts various formats:
* [TOML](https://github.com/toml-lang/toml) : configuration file with extension _.toml_ or _.conf_
* [JSON](https://en.wikipedia.org/wiki/JSON) : configuration file with extension _.json_
* [YAML](https://en.wikipedia.org/wiki/YAML) : configuration file with extension _.yaml_
* [HCL](https://github.com/hashicorp/hcl): configuration file with extension _.hcl_


## Getting started

We recommend you start by reading the [getting started](https://creativeprojects.github.io/resticprofile/configuration/getting_started/index.html) section

## Using resticprofile

The full documentation has been moved to [creativeprojects.github.io](https://creativeprojects.github.io/resticprofile/)

## Survey

What are your most important features?
Please fill in the [survey](https://github.com/creativeprojects/resticprofile/issues/415)

