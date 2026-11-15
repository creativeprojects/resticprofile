---
title: "Launchd on macOS"
weight: 110
---

`launchd` is the service manager on macOS. resticprofile can schedule a profile using the `launchctl` tool.

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