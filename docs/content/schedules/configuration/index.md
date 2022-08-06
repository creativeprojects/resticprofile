---
title: "Schedule Configuration"
weight: 10
---


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


#### Changing schedule-permission from user to system, or system to user

If you need to change the permission of a schedule, **please be sure to `unschedule` the profile before**.

This order is important:

- `unschedule` the job first. resticprofile does **not keep track of how your profile was installed**, so you have to remove the schedule first
- now you can change your permission (`user` to `system`, or `system` to `user`)
- `schedule` your updated profile

### schedule-lock-mode

Starting from version 0.14.0, `schedule-lock-mode` accepts 3 values:
- `default`: Wait on acquiring a lock for the time duration set in `schedule-lock-wait`, before failing a schedule.
   Behaves like `fail` when `schedule-lock-wait` is "0" or not specified.
- `fail`: Any lock failure causes a schedule to abort immediately. 
- `ignore`: Skip resticprofile locks. restic locks are not skipped and can abort the schedule.

### schedule-lock-wait

Sets the amount of time to wait for a resticprofile and restic lock to become available. Is only used when `schedule-lock-mode` is unset or `default`.

### schedule-log

`schedule-log` can be used in two ways:
- Allow to redirect all output from resticprofile **and restic** to a file. The parameter should point to a file (`/path/to/file`)
- Redirects all resticprofile log entries to the syslog server. In that case the parameter is a URL like: `udp://server:514` or `tcp://127.0.0.1:514`

If there's no server answering on the port specified, resticprofile will send the logs to the default output instead.

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

