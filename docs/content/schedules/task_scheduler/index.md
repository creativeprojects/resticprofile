---
title: "Windows Task Scheduler"
weight: 150
---

## Minimum version supported

Windows 10 is the minimum version supported for scheduling.

## Elevated mode

If you have any schedule profile with the permission set to `system`, resticprofile will need to run in elevated permission mode to be able to set up the schedules.

In general, you shouldn't worry about it: resticprofile will restart itself in elevated mode. You'll see the popup window asking to give resticprofile elevated privileges.

### resticprofile is blocked from restarting in elevated mode

![Unwanted software]({{< absolute "schedules/task_scheduler/unwanted_my_ass.png" nohtml >}})

There's nothing I can do on my side to prevent this, but to buy a $300 certificate. Developer certificates for Windows are way more expensive than certificates for macOS. If you know any other way for free open-source software, please let me know.

#### Solution

You'll need to start an elevated shell yourself.

- Right click on `Command Prompt`, `Windows Terminal` or `Windows Powershell` (just pick your favorite)
- Click on `Run as administrator`

![Start elevated command prompt]({{< absolute "schedules/task_scheduler/runas.png" nohtml >}})

It's easy to spot a terminal window opened with Administrator privileges:

![Administrator prefix]({{< absolute "schedules/task_scheduler/administrator.png" nohtml >}})

> [!IMPORTANT]
> By running the schedule command, Windows might also just delete _resticprofile.exe_ instead, by pretending it's a threat. I suppose installing a third party backup tool is kind-of a threat to their OneDrive backup offering after all.
