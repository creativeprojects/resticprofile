---
title: Features
weight: 1
description: Features of resticprofile
---

Here is a non-exhaustive list of what resticprofile offers:

* **Profiles**
    * No need to remember command parameters and environment variables
    * Create multiple profiles in one configuration file
    * Profiles can inherit options from other profiles
    * Create groups of profiles to run sequentially
    * Easily [schedule]({{% relref "/configuration/schedules" %}}) backups, retentions, and checks (supports *systemd*, *crond*, *launchd*, and *Windows Task Scheduler*)
    * Use [template syntax]({{% relref "/configuration/profiles/templates" %}}) in your configuration file
    * **[new for v0.29.0]** Schedule a group of profiles (configuration `v2` only)
* **Automation**
    * Run the forget command before or after a backup (in a section called *retention*)
    * Check a repository before or after a backup
    * Run [shell commands]({{% relref "/configuration/hooks/run_hooks" %}}) before or after running a profile, useful for mounting and unmounting backup disks
    * Run a [shell command]({{% relref "/configuration/hooks/run_hooks" %}}) if an error occurs
    * Send a backup stream via _stdin_
    * Start restic at different [priorities]({{% relref "/configuration/profiles/priority" %}}) (Priority Class in Windows, *nice* in Unix, and/or _ionice_ in Linux)
    * Automatically clear [stale locks]({{% relref "/usage/locks" %}})
* **Monitoring**
    * Generate a simple [status file]({{% relref "/configuration/monitoring/status" %}}) for monitoring software to ensure backups are running smoothly
    * Export a [prometheus]({{% relref "/configuration/monitoring/prometheus" %}}) file after a backup or send the report to a push gateway
    * Run shell commands in the background when non-fatal errors are detected
    * Send messages to [HTTP hooks]({{% relref "/configuration/hooks/http_hooks" %}}) before, after a successful or failed job (backup, forget, check, prune, copy)
    * Send resticprofile [logs]({{% relref "/configuration/monitoring/index" %}}) to a syslog server
* **Checks**
    * Check for [enough memory]({{% relref "/usage/memory" %}}) before starting a backup
    * Avoid scheduling a job when the system is on battery
* **Misc**
    * Generate cryptographically secure random keys for a restic [key file]({{% relref "/usage/examples/keyfile" %}})
    * Automatically [initialize the secondary repository]({{% relref "/configuration/hooks/copy" %}}) using the `copy-chunker-params` flag
    * Prevent the system from [idle sleeping]({{% relref "/configuration/schedules/configuration/#preventing-system-sleep" %}})
    * View help for both restic and resticprofile via the `help` command or `-h` flag
