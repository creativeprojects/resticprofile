---
title: "Example: MacOS"
weight: 33
---

Here's an example of scheduling a backup profile named `azure-dev`:

```shell
% resticprofile -n azure-dev schedule
2025/03/30 18:40:13 using configuration file: profiles.yaml

Profile (or Group) azure-dev: backup schedule
=============================================
  Original form: 22:22
Normalized form: *-*-* 22:22:00
    Next elapse: Sun Mar 30 22:22:00 BST 2025
       (in UTC): Sun Mar 30 21:22:00 UTC 2025
       From now: 3h41m46s left

2025/03/30 18:40:13 scheduled job azure-dev/backup created

```

{{% notice info %}}

In some cases, macOS may request permission to access the network, an external volume (like a USB drive), or a protected directory. A message window will appear while the backup runs in the background.
{{% /notice %}}

![access files popup window]({{< absolute "configuration/schedules/removable_volume.png" nohtml >}})

You may want to start the task now so you can grant special permissions:

1. Retrieve the task name using the status command:

```shell
% resticprofile -n azure-dev status
2025/03/30 18:40:21 using configuration file: profiles.yaml

Profile (or Group) azure-dev: backup schedule
=============================================
  Original form: *-*-* 22:22:00
Normalized form: *-*-* 22:22:00
    Next elapse: Sun Mar 30 22:22:00 BST 2025
       (in UTC): Sun Mar 30 21:22:00 UTC 2025
       From now: 3h41m38s left

            service: local.resticprofile.azure-dev.backup
         permission: user
            program: /usr/local/bin/resticprofile
  working directory: /Users/cp/resticprofile
        stdout path: local.resticprofile.azure-dev.backup.log
        stderr path: local.resticprofile.azure-dev.backup.log
              state: not running
           runs (*): 0
 last exit code (*): (never exited)
                 * : since last (re)schedule or last reboot
```

The name of the task can be seen on the line `service: ...`

2. start the task manually

```shell
% launchctl start local.resticprofile.azure-dev.backup
```

You can check the task is currently running:

```shell
2025/03/30 18:42:07 using configuration file: profiles.yaml

Profile (or Group) azure-dev: backup schedule
=============================================
  Original form: *-*-* 22:22:00
Normalized form: *-*-* 22:22:00
    Next elapse: Sun Mar 30 22:22:00 BST 2025
       (in UTC): Sun Mar 30 21:22:00 UTC 2025
       From now: 3h39m52s left

            service: local.resticprofile.azure-dev.backup
         permission: user
            program: /usr/local/bin/resticprofile
  working directory: /Users/cp/resticprofile
        stdout path: local.resticprofile.azure-dev.backup.log
        stderr path: local.resticprofile.azure-dev.backup.log
              state: running
           runs (*): 1
 last exit code (*): (never exited)
                 * : since last (re)schedule or last reboot

```
