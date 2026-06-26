---
title: "Windows Task Scheduler"
weight: 150
---

## Minimum version supported

Windows 10 is the minimum version supported for scheduling.

## Elevated mode

If your schedule profile has the permission set to `system`, resticprofile needs to run in elevated mode to set up the schedules.

Generally, you don't need to worry about this: resticprofile will restart itself in elevated mode. You'll see a popup window asking for elevated privileges.

### resticprofile is blocked from restarting in elevated mode

![Unwanted software]({{< absolute "schedules/task_scheduler/unwanted_my_ass.png" nohtml >}})

I can't prevent this without buying a developer certificate. If you know any free or cheap certificate for open-source software, please let me know.

#### Solution

You'll need to start an elevated shell yourself.

- Right-click on `Command Prompt`, `Windows Terminal`, or `Windows Powershell` (choose your favorite)
- Click on `Run as administrator`

![Start elevated command prompt]({{< absolute "schedules/task_scheduler/runas.png" nohtml >}})

It's easy to spot a terminal window opened with Administrator privileges:

![Administrator prefix]({{< absolute "schedules/task_scheduler/administrator.png" nohtml >}})

> [!IMPORTANT]
> Running the schedule command might cause Windows to delete _resticprofile.exe_, treating it as a threat.

## Start when available

If your computer might be asleep or off during a scheduled backup time, you can enable `schedule-start-when-available` to run the task as soon as the computer becomes available.

```yaml
profile:
  backup:
    schedule: "03:00"
    schedule-start-when-available: true
```

This sets the "Start the task as soon as possible after a scheduled start is missed" option in Windows Task Scheduler.

## Run after login

Enable `schedule-after-login` (with `schedule-permission: user_logged_on`) to add an "At log on" trigger to the task, so it runs when you log in. It can be combined with `schedule` times.

```yaml
profile:
  backup:
    schedule-after-login: true
    schedule-permission: user_logged_on
```

Note: when the task runs as the system account, the logon trigger applies to any user logging on.

Note: the registered logon trigger is not reported back by `resticprofile status` because the underlying `schtasks /query` text output does not expose triggers; the task itself works normally.
