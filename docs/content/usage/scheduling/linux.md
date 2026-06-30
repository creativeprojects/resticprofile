---
title: "Example: Linux"
weight: 32
---

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
