---
archetype: "home"
title: resticprofile
description: Configuration profiles manager for restic backup
---

Configuration profiles manager for [restic backup](https://restic.net/)

**resticprofile** bridges the gap between a configuration file and restic backup. Although creating a configuration file for restic has been [discussed](https://github.com/restic/restic/issues/16), it remains a low priority.

With resticprofile:

* No need to remember command parameters and environment variables
* Create multiple profiles in one configuration file
* Profiles can inherit options from other profiles
* Run the forget command before or after a backup (in a section called *retention*)
* Check a repository before or after a backup
* Create groups of profiles to run sequentially
* Run [shell commands]({{% relref "/configuration/run_hooks" %}}) before or after running a profile, useful for mounting and unmounting backup disks
* Run a [shell command]({{% relref "/configuration/run_hooks" %}}) if an error occurs
* Send a backup stream via _stdin_
* Start restic at different [priorities]({{% relref "/configuration/priority" %}}) (Priority Class in Windows, *nice* in Unix, and/or _ionice_ in Linux)
* Check for [enough memory]({{% relref "/usage/memory" %}}) before starting a backup
* Generate cryptographically secure random keys for a restic [key file]({{% relref "/usage/keyfile" %}})
* Easily [schedule]({{% relref "/schedules" %}}) backups, retentions, and checks (supports *systemd*, *crond*, *launchd*, and *Windows Task Scheduler*)
* Generate a simple [status file]({{% relref "/status" %}}) for monitoring software to ensure backups are running smoothly
* Use [template syntax]({{% relref "/configuration/templates" %}}) in your configuration file
* Automatically clear [stale locks]({{% relref "/usage/locks" %}})
* Export a [prometheus]({{% relref "/status/prometheus" %}}) file after a backup or send the report to a push gateway
* Run shell commands in the background when non-fatal errors are detected
* Send messages to [HTTP hooks]({{% relref "/configuration/http_hooks" %}}) before, after a successful or failed job (backup, forget, check, prune, copy)
* Automatically [initialize the secondary repository]({{% relref "/configuration/copy" %}}) using the `copy-chunker-params` flag
* Send resticprofile [logs]({{% relref "/configuration/logs" %}}) to a syslog server
* Prevent the system from [idle sleeping]({{% relref "/configuration/sleep" %}})
* View help for both restic and resticprofile via the `help` command or `-h` flag
* Avoid scheduling a job when the system is on battery
* **[new for v0.29.0]** Schedule a group of profiles (configuration `v2` only)

The configuration file supports various formats:
* [TOML](https://github.com/toml-lang/toml): files with extensions *.toml* and *.conf* (for compatibility with versions before 0.6.0)
* [JSON](https://en.wikipedia.org/wiki/JSON): files with extension *.json*
* [YAML](https://en.wikipedia.org/wiki/YAML): files with extension *.yaml*
* [HCL](https://github.com/hashicorp/hcl): files with extension *.hcl*