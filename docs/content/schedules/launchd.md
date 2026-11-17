---
title: "Launchd on macOS"
weight: 110
---

`launchd` is the service manager on macOS. resticprofile can schedule a profile using the `launchctl` tool.

## Run after login

When `schedule-after-login` is enabled (with `schedule-permission: user_logged_on`), the generated user agent sets `RunAtLoad` to `true`, so the profile runs at login. Note that `RunAtLoad` also runs the job once when the agent is first loaded (when the schedule is created), not only on subsequent logins.

## User permission

A user agent is generated when you set `schedule-permission` to `user` or `user_logged_on`. It consists of a `plist` file in `~/Library/LaunchAgents`.

If you include specific files in your backup (like contacts or calendar) or if you're connecting to an external restic repository, you may need to grant more permissions to resticprofile and restic (a popup window will ask for permission).

You can wait for the profile to start or start it manually. To start a backup profile called `remote` manually, use:

```shell
/bin/launchctl start local.resticprofile.remote.backup
```

Once you grant permission, the profile will run normally until you update resticprofile or restic. This is a macOS limitation.

![access files popup window]({{< absolute "schedules/removable_volume.png" nohtml >}})

{{% notice style="note" title="Local network access on macOS 15 (Sequoia) and later" %}}
Since macOS 15 (Sequoia), an application needs your permission to connect to devices on your local network. A scheduled profile that uses a repository or a Prometheus Pushgateway on your local network can fail with a `no route to host` error, even though the same command works from your terminal.

After scheduling the profile, start it manually with `launchctl start` (as shown above) and wait for the popup window asking for permission to access the local network.

You can review this permission later in **System Settings > Privacy & Security > Local Network**.
{{% /notice %}}

## System permission

A launchd daemon is generated when you set `schedule-permission` to `system`. It consists of a `plist` file in `/Library/LaunchDaemons`.

Run resticprofile with sudo to `schedule` and `unschedule` the profile. You can schedule and unschedule system and user profiles simultaneously using the `schedule --all` command.

## Log files

When `schedule-log` is not set, launchd saves the output of the scheduled job to a file named after the job, like `local.resticprofile.remote.backup.log` for the `backup` command of a profile called `remote`.

This file is created in the [working directory]({{% relref "/schedules/commands/index.html#working-directory" %}}) of the job, which is the directory you were in when you ran the `schedule` command.

To save the logs somewhere else, you can either:
- set [schedule-log]({{% relref "/schedules/configuration/index.html#schedule-log" %}}) to the full path of a log file. In that case launchd does not create its own log file.
- run the `schedule` command again from the directory where you want the log files.

{{% notice note %}}
The `base-dir` option of a profile has no effect on the location of these files: launchd opens them before resticprofile starts.
{{% /notice %}}

{{% notice info %}}
Before version `v0.33.0`, launchd still created its own log files when `schedule-log` was set. These files stayed empty.
{{% /notice %}}