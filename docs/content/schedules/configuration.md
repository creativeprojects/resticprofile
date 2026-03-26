---
title: "Schedule Configuration"
weight: 10
---


The schedule configuration includes several parameters that can be added to each profile:

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

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
{{% tab title="yaml" %}}

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
{{% tab title="hcl" %}}

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
{{% tab title="json" %}}

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
{{< /tabs >}}


## schedule-permission

`schedule-permission` accepts three parameters: `system`, `user`, or `user_logged_on`:

* `system`: Access system or protected files. Run resticprofile with `sudo` on Unix and with elevated prompt on Windows. On Windows, resticprofile will automatically request elevated permissions if needed.

* `user`: Run the backup using current user permissions. Suitable for saving documents or files within your profile. **This mode runs even when the user is not logged on**. It will ask for your user password on Windows. It needs root permission (via sudo) when using `systemd`.

* `user_logged_on`: **Not for crond** - Provides the same permissions as `user`, but runs only when the user is logged on. On Windows, it does not ask for your user password.

* *empty*: resticprofile will guess based on how it was started (with sudo or as a normal user). The fallback is `system` on Windows and `user_logged_on` on other platforms.

### Changing schedule-permission

To change the permission of a schedule, unschedule the profile first.

Follow this order:

- Unschedule the job first (before updating the permission in the configuration)
- Change your permission (user to system, or system to user).
- Schedule your updated profile.

## schedule-run-level

In Windows Task Manager, you can also specify a scheduled task privilege level.

The `schedule-run-level` parameter accepts three values:
- `lowest`: Runs the task with the least user privileges.
- `highest`: Runs the task with the highest user privileges available.
- `auto`: Uses `highest` if `schedule-permission` is set to `system`. Otherwise, defaults to `lowest`.

## schedule-lock-mode

`schedule-lock-mode` accepts 3 values:
- `default`: Waits for the duration set in `schedule-lock-wait` before failing a schedule. Acts like `fail` if `schedule-lock-wait` is "0" or not specified.
- `fail`: Any lock failure immediately aborts the schedule.
- `ignore`: Skips resticprofile locks. Restic locks are not skipped and can abort the schedule.

## schedule-lock-wait

Sets the wait time for a resticprofile and restic lock to become available. Used only when `schedule-lock-mode` is unset or `default`.


## schedule-log

`schedule-log` can be used in two ways:
- Redirect all output from resticprofile **and restic** to a file. The parameter should point to a file (`/path/to/file`).
- Redirect all resticprofile log entries to the syslog server. In this case, the parameter is a URL like `udp://server:514` or `tcp://127.0.0.1:514`.

If no server responds on the specified port, resticprofile will send the logs to the default output instead.


## schedule-priority

`schedule-priority` accepts two values:
- `background`: The process runs unnoticed while you work.
- `standard`: The process gets the same priority as other processes (won't run faster if the machine is idle).

`schedule-priority` is not available for crond.

## schedule

The `schedule` parameter accepts various forms of input from the [systemd calendar event](https://www.freedesktop.org/software/systemd/man/systemd.time.html#Calendar%20Events) type. This format is the same used to schedule on macOS and Windows.

The general form is:
```
weekdays year-month-day hour:minute:second
```

- Use `*` to mean any
- Use `,` to separate multiple entries
- Use `..` for a range

**Limitations**:
- The divider (`/`), the `~`, and timezones are not supported on macOS and Windows.
- The `year` and `second` fields have no effect on macOS and limited availability on Windows.

Here are a few examples (taken from the systemd documentation):

```

The user input is on the left, and the system's full format is on the right.

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

The `schedule` can be a string or an array of strings (to allow for multiple schedules).

## schedule-ignore-on-battery

If set to `true`, the schedule won't start if the system is running on battery (even if the charge is at 100%).

## schedule-ignore-on-battery-less-than

If set to a number, the schedule won't start if the system is running on battery and the charge is less than or equal to the specified number.

## schedule-hide-window

When `schedule-permission` is set to `user_logged_on`, Windows Task Scheduler runs tasks in the foreground.
This behavior may interrupt the user's activity and is often undesirable.

To prevent that, set this option to `true` to hide the task window by wrapping the execution in `conhost.exe --headless`.

Note: It works only on Windows and makes sense only with `user_logged_on` permission.

Note: The behavior of `conhost.exe` varies between Windows versions. It has been confirmed to work on Windows 11 (24H2) but not on Windows 10 (1607).

## schedule-start-when-available

When set to `true`, Windows Task Scheduler will start the task as soon as possible after a scheduled start is missed. This is useful when the computer might be asleep or off during the scheduled time.

For example, if a backup is scheduled for 3:00 AM but the computer is off, enabling this option will run the backup when the computer is next available.

Note: This option only works on Windows.

## Example 

Here's an example of a scheduling configuration:

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

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
{{% tab title="yaml" %}}

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
{{% tab title="hcl" %}}

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
{{% tab title="json" %}}

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
{{< /tabs >}}

