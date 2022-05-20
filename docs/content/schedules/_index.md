+++
chapter = true
pre = "<b>4. </b>"
title = "Schedules"
weight = 20
+++


# Scheduled backups

resticprofile is capable of managing scheduled backups for you using:
- **launchd** on macOS X
- **Task Scheduler** on Windows
- **systemd** where available (Linux and other BSDs)
- **crond** on supported platforms (Linux and other BSDs)

On unixes (except macOS) resticprofile is using **systemd** by default. **crond** can be used instead if configured in `global` `scheduler` parameter:

{{< tabs groupId="config-with-json" >}}
{{% tab name="toml" %}}

```toml
[global]
  scheduler = "crond"
```

{{% /tab %}}
{{% tab name="yaml" %}}

```yaml
---
global:
    scheduler: crond
```

{{% /tab %}}
{{% tab name="hcl" %}}

```hcl
"global" = {
  "scheduler" = "crond"
}
```

{{% /tab %}}
{{% tab name="json" %}}

```json
{
  "global": {
    "scheduler": "crond"
  }
}
```

{{% /tab %}}
{{% /tabs %}}




Each profile can be scheduled independently (groups are not available for scheduling yet - it will be available in version '2' of the configuration file).

These 5 profile sections are accepting a schedule configuration:
- backup
- check
- forget (version 0.11.0)
- prune (version 0.11.0)
- copy (version 0.16.0)

which mean you can schedule `backup`, `forget`, `prune`, `check` and `copy` independently (I recommend to use a local `lock` in this case).

## retention schedule is deprecated
**Important**:
starting from version 0.11.0 the schedule of the `retention` section is **deprecated**: Use the `forget` section instead.


## Schedule configuration

The schedule configuration consists of a few parameters which can be added on each profile:

{{< tabs groupId="config-with-json" >}}
{{% tab name="toml" %}}

```toml
[profile.backup]
  schedule = "*:00,30"
  schedule-permission = "system"
  schedule-priority = "background"
  schedule-log = "profile-backup.log"
  schedule-lock-mode = "default"
  schedule-lock-wait = "15m30s"
```

{{% /tab %}}
{{% tab name="yaml" %}}

```yaml
profile:
  backup:
    schedule: '*:00,30'
    schedule-permission: system
    schedule-priority: background
    schedule-log: profile-backup.log
    schedule-lock-mode: default
    schedule-lock-wait: 15m30s
```

{{% /tab %}}
{{% tab name="hcl" %}}

```hcl
"profile" "backup" {
  "schedule" = "*:00,30"
  "schedule-permission" = "system"
  "schedule-priority" = "background"
  "schedule-log" = "profile-backup.log"
  "schedule-lock-mode" = "default"
  "schedule-lock-wait" = "15m30s"
}
```

{{% /tab %}}
{{% tab name="json" %}}

```json
{
  "profile": {
    "backup": {
      "schedule": "*:00,30",
      "schedule-permission": "system",
      "schedule-priority": "background",
      "schedule-log": "profile-backup.log",
      "schedule-lock-mode": "default",
      "schedule-lock-wait": "15m30s"
    }
  }
}
```

{{% /tab %}}
{{% /tabs %}}



### schedule-permission

`schedule-permission` accepts two parameters: `user` or `system`:

* `user`: your backup will be running using your current user permissions on files. This is fine if you're only saving your documents (or any other file inside your profile). Please note on **systemd** that the schedule **will only run when your user is logged in**.

* `system`: if you need to access some system or protected files. You will need to run resticprofile with `sudo` on unixes and with elevated prompt on Windows (please note on Windows resticprofile will ask you for elevated permissions automatically if needed).

* *empty*: resticprofile will try its best guess based on how you started it (with sudo or as a normal user) and fallback to `user`

### schedule-lock-mode

Starting from version 0.14.0, `schedule-lock-mode` accepts 3 values:
- `default`: Wait on acquiring a lock for the time duration set in `schedule-lock-wait`, before failing a schedule.
   Behaves like `fail` when `schedule-lock-wait` is "0" or not specified.
- `fail`: Any lock failure causes a schedule to abort immediately. 
- `ignore`: Skip resticprofile locks. restic locks are not skipped and can abort the schedule.

### schedule-lock-wait

Sets the amount of time to wait for a resticprofile and restic lock to become available. Is only used when `schedule-lock-mode` is unset or `default`.

### schedule-log

Allow to redirect all output from resticprofile and restic to a file

### schedule-priority (systemd and launchd only)

Starting from version 0.11.0, `schedule-priority` accepts two values:
- `background`: the process shouldn't be noticeable when working on the machine at the same time (this is the default)
- `standard`: the process should get the same priority as any other process on the machine (but it won't run faster if you're not using the machine at the same time)

`schedule-priority` is not available for windows task scheduler, nor crond

### schedule

The `schedule` parameter accepts many forms of input from the [systemd calendar event](https://www.freedesktop.org/software/systemd/man/systemd.time.html#Calendar%20Events) type. This is by far the easiest to use: **It is the same format used to schedule on macOS and Windows**.

The most general form is:
```
weekdays year-month-day hour:minute:second
```

- use `*` to mean any
- use `,` to separate multiple entries
- use `..` for a range

**limitations**:
- the divider (`/`), the `~` and timezones are not (yet?) supported on macOS and Windows.
- the `year` and `second` fields have no effect on macOS. They do have limited availability on Windows (they don't make much sense anyway).

Here are a few examples (taken from the systemd documentation):

```
On the left is the user input, on the right is the full format understood by the system

  Sat,Thu,Mon..Wed,Sat..Sun → Mon..Thu,Sat,Sun *-*-* 00:00:00
      Mon,Sun 12-*-* 2,1:23 → Mon,Sun 2012-*-* 01,02:23:00
                    Wed *-1 → Wed *-*-01 00:00:00
           Wed..Wed,Wed *-1 → Wed *-*-01 00:00:00
                 Wed, 17:48 → Wed *-*-* 17:48:00
Wed..Sat,Tue 12-10-15 1:2:3 → Tue..Sat 2012-10-15 01:02:03
                *-*-7 0:0:0 → *-*-07 00:00:00
                      10-15 → *-10-15 00:00:00
        monday *-12-* 17:00 → Mon *-12-* 17:00:00
     Mon,Fri *-*-3,1,2 *:30 → Mon,Fri *-*-01,02,03 *:30:00
       12,14,13,12:20,10,30 → *-*-* 12,13,14:10,20,30:00
            12..14:10,20,30 → *-*-* 12..14:10,20,30:00
                03-05 08:05 → *-03-05 08:05:00
                      05:40 → *-*-* 05:40:00
        Sat,Sun 12-05 08:05 → Sat,Sun *-12-05 08:05:00
              Sat,Sun 08:05 → Sat,Sun *-*-* 08:05:00
           2003-03-05 05:40 → 2003-03-05 05:40:00
             2003-02..04-05 → 2003-02..04-05 00:00:00
                 2003-03-05 → 2003-03-05 00:00:00
                      03-05 → *-03-05 00:00:00
                     hourly → *-*-* *:00:00
                      daily → *-*-* 00:00:00
                    monthly → *-*-01 00:00:00
                     weekly → Mon *-*-* 00:00:00
                     yearly → *-01-01 00:00:00
                   annually → *-01-01 00:00:00
```

The `schedule` can be a string or an array of string (to allow for multiple schedules)

Here's an example of a scheduling configuration:

{{< tabs groupId="config-with-json" >}}
{{% tab name="toml" %}}

```toml
[default]
  repository = "d:\\backup"
  password-file = "key"

[self]
  inherit = "default"

  [self.retention]
    after-backup = true
    keep-within = "14d"

  [self.backup]
    source = "."
    schedule = [ "Mon..Fri *:00,15,30,45", "Sat,Sun 0,12:00" ]
    schedule-permission = "user"
    schedule-lock-wait = "10m"

  [self.prune]
    schedule = "sun 3:30"
    schedule-permission = "user"
    schedule-lock-wait = "1h"
```

{{% /tab %}}
{{% tab name="yaml" %}}

```yaml
default:
  repository: "d:\\backup"
  password-file: key

self:
  inherit: default
  retention:
    after-backup: true
    keep-within: 14d
  backup:
    source: "."
    schedule:
    - "Mon..Fri *:00,15,30,45" # every 15 minutes on weekdays
    - "Sat,Sun 0,12:00"        # twice a day on week-ends
    schedule-permission: user
    schedule-lock-wait: 10m
  prune:
    schedule: "sun 3:30"
    schedule-permission: user
    schedule-lock-wait: 1h
```

{{% /tab %}}
{{% tab name="hcl" %}}

```hcl
"default" = {
  "repository" = "d:\\backup"
  "password-file" = "key"
}

"self" = {
  "inherit" = "default"

  "retention" = {
    "after-backup" = true
    "keep-within" = "14d"
  }

  "backup" = {
    "source" = "."
    "schedule" = ["Mon..Fri *:00,15,30,45", "Sat,Sun 0,12:00"]
    "schedule-permission" = "user"
    "schedule-lock-wait" = "10m"
  }

  "prune" = {
    "schedule" = "sun 3:30"
    "schedule-permission" = "user"
    "schedule-lock-wait" = "1h"
  }
}
```

{{% /tab %}}
{{% tab name="json" %}}

```json
{
  "default": {
    "repository": "d:\\backup",
    "password-file": "key"
  },
  "self": {
    "inherit": "default",
    "retention": {
      "after-backup": true,
      "keep-within": "14d"
    },
    "backup": {
      "source": ".",
      "schedule": [
        "Mon..Fri *:00,15,30,45",
        "Sat,Sun 0,12:00"
      ],
      "schedule-permission": "user",
      "schedule-lock-wait": "10m"
    },
    "prune": {
      "schedule": "sun 3:30",
      "schedule-permission": "user",
      "schedule-lock-wait": "1h"
    }
  }
}
```

{{% /tab %}}
{{% /tabs %}}


## Scheduling commands

resticprofile accepts these internal commands:
- schedule
- unschedule
- status

All internal commands either operate on the profile selected by `--name`, on the profiles selected by a group, or on all profiles when the flag `--all` is passed.

Examples:
```
resticprofile --name profile schedule 
resticprofile --name group schedule 
resticprofile schedule --all 
```

Please note, schedules are always independent of each other no matter whether they have been created with `--all`, by group or from a single profile.

### schedule command

Install all the schedules defined on the selected profile or profiles.

Please note on systemd, we need to `start` the timer once to enable it. Otherwise it will only be enabled on the next reboot. If you **dont' want** to start (and enable) it now, pass the `--no-start` flag to the command line.

Also if you use the `--all` flag to schedule all your profiles at once, make sure you use only the `user` mode or `system` mode. A combination of both would not schedule the tasks properly:
- if the user is not privileged, only the `user` tasks will be scheduled
- if the user **is** privileged, **all schedule will end-up as a `system` schedule**

### unschedule command

Remove all the schedules defined on the selected profile or profiles.

### status command

Print the status on all the installed schedules of the selected profile or profiles. 

The display of the `status` command will be OS dependant. Please see the examples below on which output you can expect from it.

### Examples of scheduling commands under Windows

If you create a task with `user` permission under Windows, you will need to enter your password to validate the task. It's a requirement of the task scheduler. I'm inviting you to review the code to make sure I'm not emailing your password to myself. Seriously you shouldn't trust anyone.

Example of the `schedule` command under Windows (with git bash):

```
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

```
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

```
$ resticprofile -c examples/windows.yaml -n self unschedule
2020/07/22 21:34:51 scheduled job self/backup removed
2020/07/22 21:34:51 scheduled job self/retention removed
```

### Examples of scheduling commands under Linux

With this example of configuration for Linux:

{{< tabs groupId="config-with-json" >}}
{{% tab name="toml" %}}

```toml
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
{{% tab name="yaml" %}}

```yaml
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
{{% tab name="hcl" %}}

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
{{% tab name="json" %}}

```json
{
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
{{% /tabs %}}


```
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

```
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

```
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

```
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

```
% resticprofile -v -c examples/private/azure.yaml -n self schedule --no-start
```

## Changing schedule-permission from user to system, or system to user

If you need to change the permission of a schedule, **please be sure to `unschedule` the profile before**.

This order is important:

- `unschedule` the job first. resticprofile does **not keep track of how your profile was installed**, so you have to remove the schedule first
- now you can change your permission (`user` to `system`, or `system` to `user`)
- `schedule` your updated profile
