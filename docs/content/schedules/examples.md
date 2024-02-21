---
title: "Schedule Examples"
weight: 30
---


### Examples of scheduling commands under Windows

{{% notice note %}}
If you create a task with `user` permission under Windows, you will need to enter your password to validate the task.

It's a requirement of the task scheduler. I'm inviting you to review the code to make sure I'm not emailing your password to myself. Seriously you shouldn't trust anyone.
{{% /notice %}}

Example of the `schedule` command under Windows (with git bash):

```shell
$ resticprofile -c examples/windows.yaml -n self schedule

Analyzing backup schedule 1/2
=================================
  Original form: Mon..Fri *:00,15,30,45
Normalized form: Mon..Fri *-*-* *:00,15,30,45:00
    Next elapse: Wed Jul 22 21:30:00 BST 2020
       (in UTC): Wed Jul 22 20:30:00 UTC 2020
       From now: 1m52s left

Analyzing backup schedule 2/2
=================================
  Original form: Sat,Sun 0,12:00
Normalized form: Sat,Sun *-*-* 00,12:00:00
    Next elapse: Sat Jul 25 00:00:00 BST 2020
       (in UTC): Fri Jul 24 23:00:00 UTC 2020
       From now: 50h31m52s left

Creating task for user Creative Projects
Task Scheduler requires your Windows password to validate the task: 

2020/07/22 21:28:15 scheduled job self/backup created

Analyzing retention schedule 1/1
=================================
  Original form: sun 3:30
Normalized form: Sun *-*-* 03:30:00
    Next elapse: Sun Jul 26 03:30:00 BST 2020
       (in UTC): Sun Jul 26 02:30:00 UTC 2020
       From now: 78h1m44s left

2020/07/22 21:28:22 scheduled job self/retention created
```

To see the status of the triggers, you can use the `status` command:

```shell
$ resticprofile -c examples/windows.yaml -n self status

Analyzing backup schedule 1/2
=================================
  Original form: Mon..Fri *:00,15,30,45
Normalized form: Mon..Fri *-*-* *:00,15,30,45:00
    Next elapse: Wed Jul 22 21:30:00 BST 2020
       (in UTC): Wed Jul 22 20:30:00 UTC 2020
       From now: 14s left

Analyzing backup schedule 2/2
=================================
  Original form: Sat,Sun 0,12:*
Normalized form: Sat,Sun *-*-* 00,12:*:00
    Next elapse: Sat Jul 25 00:00:00 BST 2020
       (in UTC): Fri Jul 24 23:00:00 UTC 2020
       From now: 50h29m46s left

           Task: \resticprofile backup\self backup
           User: Creative Projects
    Working Dir: D:\Source\resticprofile
           Exec: D:\Source\resticprofile\resticprofile.exe --no-ansi --config examples/windows.yaml --name self backup
        Enabled: true
          State: ready
    Missed runs: 0
  Last Run Time: 2020-07-22 21:30:00 +0000 UTC
    Last Result: 0
  Next Run Time: 2020-07-22 21:45:00 +0000 UTC

Analyzing retention schedule 1/1
=================================
  Original form: sun 3:30
Normalized form: Sun *-*-* 03:30:00
    Next elapse: Sun Jul 26 03:30:00 BST 2020
       (in UTC): Sun Jul 26 02:30:00 UTC 2020
       From now: 77h59m46s left

           Task: \resticprofile backup\self retention
           User: Creative Projects
    Working Dir: D:\Source\resticprofile
           Exec: D:\Source\resticprofile\resticprofile.exe --no-ansi --config examples/windows.yaml --name self forget
        Enabled: true
          State: ready
    Missed runs: 0
  Last Run Time: 1999-11-30 00:00:00 +0000 UTC
    Last Result: 267011
  Next Run Time: 2020-07-26 03:30:00 +0000 UTC

```

To remove the schedule, use the `unschedule` command:

```shell
$ resticprofile -c examples/windows.yaml -n self unschedule
2020/07/22 21:34:51 scheduled job self/backup removed
2020/07/22 21:34:51 scheduled job self/retention removed
```


### Examples of scheduling commands under Linux

With this example of configuration for Linux:

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
version = "1"

[default]
  password-file = "key"
  repository = "/tmp/backup"

[test1]
  inherit = "default"

  [test1.backup]
    source = "./"
    schedule = "*:00,15,30,45"
    schedule-permission = "user"
    schedule-lock-wait = "15m"

  [test1.check]
    schedule = "*-*-1"
    schedule-permission = "user"
    schedule-lock-wait = "15m"
```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

default:
  password-file: key
  repository: /tmp/backup

test1:
  inherit: default
  backup:
    source: ./
    schedule: "*:00,15,30,45"
    schedule-permission: user
    schedule-lock-wait: 15m
  check:
    schedule: "*-*-1"
    schedule-permission: user
    schedule-lock-wait: 15m

```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
"default" = {
  "password-file" = "key"
  "repository" = "/tmp/backup"
}

"test1" = {
  "inherit" = "default"

  "backup" = {
    "source" = "./"
    "schedule" = "*:00,15,30,45"
    "schedule-permission" = "user"
    "schedule-lock-wait" = "15m"
  }

  "check" = {
    "schedule" = "*-*-1"
    "schedule-permission" = "user"
    "schedule-lock-wait" = "15m"
  }
}
```

{{% /tab %}}
{{% tab title="json" %}}

```json
{
  "version": "1",
  "default": {
    "password-file": "key",
    "repository": "/tmp/backup"
  },
  "test1": {
    "inherit": "default",
    "backup": {
      "source": "./",
      "schedule": "*:00,15,30,45",
      "schedule-permission": "user",
      "schedule-lock-wait": "15m"
    },
    "check": {
      "schedule": "*-*-1",
      "schedule-permission": "user",
      "schedule-lock-wait": "15m"
    }
  }
}
```

{{% /tab %}}
{{< /tabs >}}


```shell
$ resticprofile -c examples/linux.yaml -n test1 schedule

Analyzing backup schedule 1/1
=================================
  Original form: *:00,15,30,45
Normalized form: *-*-* *:00,15,30,45:00
    Next elapse: Thu 2020-07-23 17:15:00 BST
       (in UTC): Thu 2020-07-23 16:15:00 UTC
       From now: 6min left

2020/07/23 17:08:51 writing /home/user/.config/systemd/user/resticprofile-backup@profile-test1.service
2020/07/23 17:08:51 writing /home/user/.config/systemd/user/resticprofile-backup@profile-test1.timer
Created symlink /home/user/.config/systemd/user/timers.target.wants/resticprofile-backup@profile-test1.timer → /home/user/.config/systemd/user/resticprofile-backup@profile-test1.timer.
2020/07/23 17:08:51 scheduled job test1/backup created

Analyzing check schedule 1/1
=================================
  Original form: *-*-1
Normalized form: *-*-01 00:00:00
    Next elapse: Sat 2020-08-01 00:00:00 BST
       (in UTC): Fri 2020-07-31 23:00:00 UTC
       From now: 1 weeks 1 days left

2020/07/23 17:08:51 writing /home/user/.config/systemd/user/resticprofile-check@profile-test1.service
2020/07/23 17:08:51 writing /home/user/.config/systemd/user/resticprofile-check@profile-test1.timer
Created symlink /home/user/.config/systemd/user/timers.target.wants/resticprofile-check@profile-test1.timer → /home/user/.config/systemd/user/resticprofile-check@profile-test1.timer.
2020/07/23 17:08:51 scheduled job test1/check created
```

The `status` command shows a combination of `journalctl` displaying errors (only) in the last month and `systemctl status`:

```shell
$ resticprofile -c examples/linux.yaml -n test1 status

Analyzing backup schedule 1/1
=================================
  Original form: *:00,15,30,45
Normalized form: *-*-* *:00,15,30,45:00
    Next elapse: Tue 2020-07-28 15:15:00 BST
       (in UTC): Tue 2020-07-28 14:15:00 UTC
       From now: 4min 44s left

Recent log (>= warning in the last month)
==========================================
-- Logs begin at Wed 2020-06-17 11:09:19 BST, end at Tue 2020-07-28 15:10:10 BST. --
Jul 27 20:48:01 Desktop76 systemd[2986]: Failed to start resticprofile backup for profile test1 in examples/linux.yaml.
Jul 27 21:00:55 Desktop76 systemd[2986]: Failed to start resticprofile backup for profile test1 in examples/linux.yaml.
Jul 27 21:15:34 Desktop76 systemd[2986]: Failed to start resticprofile backup for profile test1 in examples/linux.yaml.

Systemd timer status
=====================
● resticprofile-backup@profile-test1.timer - backup timer for profile test1 in examples/linux.yaml
   Loaded: loaded (/home/user/.config/systemd/user/resticprofile-backup@profile-test1.timer; enabled; vendor preset: enabled)
   Active: active (waiting) since Tue 2020-07-28 15:10:06 BST; 8s ago
  Trigger: Tue 2020-07-28 15:15:00 BST; 4min 44s left

Jul 28 15:10:06 Desktop76 systemd[2951]: Started backup timer for profile test1 in examples/linux.yaml.


Analyzing check schedule 1/1
=================================
  Original form: *-*-1
Normalized form: *-*-01 00:00:00
    Next elapse: Sat 2020-08-01 00:00:00 BST
       (in UTC): Fri 2020-07-31 23:00:00 UTC
       From now: 3 days left

Recent log (>= warning in the last month)
==========================================
-- Logs begin at Wed 2020-06-17 11:09:19 BST, end at Tue 2020-07-28 15:10:10 BST. --
Jul 27 19:39:59 Desktop76 systemd[2986]: Failed to start resticprofile check for profile test1 in examples/linux.yaml.

Systemd timer status
=====================
● resticprofile-check@profile-test1.timer - check timer for profile test1 in examples/linux.yaml
   Loaded: loaded (/home/user/.config/systemd/user/resticprofile-check@profile-test1.timer; enabled; vendor preset: enabled)
   Active: active (waiting) since Tue 2020-07-28 15:10:07 BST; 7s ago
  Trigger: Sat 2020-08-01 00:00:00 BST; 3 days left

Jul 28 15:10:07 Desktop76 systemd[2951]: Started check timer for profile test1 in examples/linux.yaml.


```

And `unschedule`:

```shell
$ resticprofile -c examples/linux.yaml -n test1 unschedule
Removed /home/user/.config/systemd/user/timers.target.wants/resticprofile-backup@profile-test1.timer.
2020/07/23 17:13:42 scheduled job test1/backup removed
Removed /home/user/.config/systemd/user/timers.target.wants/resticprofile-check@profile-test1.timer.
2020/07/23 17:13:42 scheduled job test1/check removed
```

### Examples of scheduling commands under macOS

macOS has a very tight protection system when running scheduled tasks (also called agents).

Under macOS, resticprofile is asking if you want to start a profile right now so you can give the access needed to the task, which consists on a few popup windows (you can disable this behavior by adding the flag `--no-start` after the schedule command).

Here's an example of scheduling a backup to Azure (which needs network access):

```shell
% resticprofile -v -c examples/private/azure.yaml -n self schedule

Analyzing backup schedule 1/1
=================================
  Original form: *:0,15,30,45:00
Normalized form: *-*-* *:00,15,30,45:00
    Next elapse: Tue Jul 28 23:00:00 BST 2020
       (in UTC): Tue Jul 28 22:00:00 UTC 2020
       From now: 2m34s left


By default, a macOS agent access is restricted. If you leave it to start in the background it's likely to fail.
You have to start it manually the first time to accept the requests for access:

% launchctl start local.resticprofile.self.backup

Do you want to start it now? (Y/n):
2020/07/28 22:57:26 scheduled job self/backup created
```

Right after you started the profile, you should get some popup asking you to grant access to various files/folders/network.

If you backup your files to an external repository on a network, you should get this popup window:

!["resticprofile" would like to access files on a network volume](https://github.com/creativeprojects/resticprofile/raw/master/network_volume.png)

**Note:**
If you prefer not being asked, you can add the `--no-start` flag like so:

```shell
resticprofile -v -c examples/private/azure.yaml -n self schedule --no-start
```
