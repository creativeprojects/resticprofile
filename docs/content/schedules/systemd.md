---
title: "Systemd"
weight: 105
---


**systemd** is a common service manager used by many Linux distributions. resticprofile can create systemd timer and service files.

User systemd units are created under the user's systemd profile (`~/.config/systemd/user`).

System units are created in `/etc/systemd/system`.

## systemd calendars

resticprofile uses the systemd [OnCalendar](https://www.freedesktop.org/software/systemd/man/systemd.time.html#Calendar%20Events) format to schedule events.

Test systemd calendars with [systemd-analyze](https://www.freedesktop.org/software/systemd/man/latest/systemd-analyze.html#systemd-analyze%20calendar%20EXPRESSION...). It will display the next trigger time:

```shell
systemd-analyze calendar 'daily'

  Original form: daily
Normalized form: *-*-* 00:00:00
    Next elapse: Sat 2020-04-18 00:00:00 CDT
       (in UTC): Sat 2020-04-18 05:00:00 UTC
       From now: 10h left
```

## First time schedule

When you schedule a profile with the `schedule` command, resticprofile will:
- Create the unit file (type [notify](https://www.freedesktop.org/software/systemd/man/latest/systemd.service.html#Type=))
- Create the timer file
- Run `systemctl daemon-reload` (if `schedule-permission` is set to `system`)
- Run `systemctl enable`
- Run `systemctl start`

## Priority and CPU scheduling

resticprofile allows you to set the `nice` value, CPU scheduling policy, and IO nice values for the systemd service. This works properly for resticprofile >= 0.29.0.

| systemd unit option  | resticprofile option |
|----------------------|----------------------|
| CPUSchedulingPolicy  | Set to `idle` if `priority` is `background`, otherwise defaults to standard policy |
| Nice                 | `nice` from `global` section |
| IOSchedulingClass    | `ionice-class` from `global` section |
| IOSchedulingPriority | `ionice-level` from `global` section |

{{% notice note %}}
When setting `CPUSchedulingPolicy` to `idle` (by setting `priority` to `background`), the backup might never execute if all your CPU cores are always busy.
{{% /notice %}}


## Permission

Until version v0.30.0, the `user` permission was actually `user_logged_on` unless you activated [lingering](https://wiki.archlinux.org/title/Systemd/User#Automatic_start-up_of_systemd_user_instances) for the user.

This is now fixed:

| Permission         | Type of unit                              | Without lingering                | With lingering      |
|--------------------|-------------------------------------------|----------------------------------|---------------------|
| **system**         | system service                            | can run any time                 | can run any time    |
| **user**           | system service with User= field defined   | can run any time                 | can run any time    |
| **user_logged_on** | user service                              | runs only when user is logged on | can run any time    |



## Run after the network is up

Setting the profile option `schedule-after-network-online: true` ensures scheduled services wait for a network connection before running. This is achieved with an [After=network-online.target](https://systemd.io/NETWORK_ONLINE/) entry in the service.


## systemd drop-in files

You can automatically populate `*.conf.d` [drop-in files](https://www.freedesktop.org/software/systemd/man/latest/systemd-system.conf.html#main-conf) for profiles, allowing easy overrides of generated services without [modifying service templates]({{% relref "/schedules/systemd/#how-to-change-the-default-systemd-unit-and-timer-file-using-a-template" %}}). For example:

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}
```toml
version = "1"

[root]
  systemd-drop-in-files = ["99-drop-in-example.conf"]

  [root.backup]
    schedule = "hourly"
    schedule-permission = "system"
    schedule-lock-wait = "45m"
    schedule-after-network-online = true
```
{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
---
version: 1

root:
  systemd-drop-in-files:
    - "99-drop-in-example.conf"

  backup:
    schedule: hourly
    schedule-permission: system
    schedule-lock-wait: 45m
    schedule-after-network-online: true
```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
"version" = "1"

"root" = {
  "systemd-drop-in-files" = ["99-drop-in-example.conf"]
  "backup" = {
    "schedule" = "hourly"
    "schedule-permission" = "system"
    "schedule-lock-wait" = "45m"
    "schedule-after-network-online" = true
  }
}
```

{{% /tab %}}
{{% tab title="json" %}}

```json
{
  "version": "1",
  "root": {
    "systemd-drop-in-files": ["99-drop-in-example.conf"],
    "backup": {
      "schedule": "hourly",
      "schedule-permission": "system",
      "schedule-lock-wait": "45m",
      "schedule-after-network-online": true
    }
  }
}
```

{{% /tab %}}
{{< /tabs >}}


Where `99-drop-in-example.conf` is in the same directory as `profiles.toml` and with the contents

```conf
[Service]
Environment=RCLONE_CONFIG=%d/rclone.conf
SetCredentialEncrypted=restic-repo-password: \
        Whxqht+dQJax1aZeCGLxmiAAAAABAAAADAAAABAAAABl6ctIWEqgRC4yHbgAAAAA8umMn \
        +6KYd8tAL58jUmtf/5wckDcxQSeuo+xd9OzN5XG7QW0iBIRRGCuWvvuAAiHEAKSk9MR8p \
        EDSaSm
SetCredentialEncrypted=rclone.conf: \
        Whxqht+dQJax1aZeCGLxmiAAAAABAAAADAAAABAAAAC+vNhJYedv5QmyDHYAAAAAimeli \
        +Oo+URGN47SUBf7Jm1n3gdu22+Sd/eL7CjzpYQvHAMOCY8xz9hp9kW9/DstWHTfdsHJo7 \
        thOpk4IbSSazCPwEr39VVQONLxzpRlY22LkQKLoGAVD4Yifk+U5aJJ4FlRW/VGpPoef2S \
        rGvQzqQI7kNX+v7EPXj4B0tSUeBBJJCEu4mgajZNAhwHtbw==
```

Generated with the following. See [systemd credentials docs](https://systemd.io/CREDENTIALS/) for more details. This allows using a TPM-backed encrypted password outside the resticprofile config.

```shell
systemd-ask-password -n | sudo systemd-creds encrypt --name=restic-repo-password -p - -
sudo systemd-creds encrypt --name=rclone.conf -p - - <<EOF
[restic-example]
type = smb
host = example
user = restic
pass = $(systemd-ask-password -n "smb restic user password" | rclone obscure -)
EOF
```

## How to change the default systemd unit and timer file using a template

By default, resticprofile automatically generates a systemd unit and timer.

You can create custom templates to add features (e.g., sending an email on failure).

The format is a [Go template](https://pkg.go.dev/text/template). Specify your custom unit and/or timer file in the global section of the configuration to apply it to all profiles:

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
[global]
  systemd-unit-template = "service.tmpl"
  systemd-timer-template = "timer.tmpl"
```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
---
global:
  systemd-unit-template: service.tmpl
  systemd-timer-template: timer.tmpl
```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
"global" = {
  "systemd-unit-template" = "service.tmpl"
  "systemd-timer-template" = "timer.tmpl"
}
```

{{% /tab %}}
{{% tab title="json" %}}

```json
{
  "global": {
    "systemd-unit-template": "service.tmpl",
    "systemd-timer-template": "timer.tmpl"
  }
}
```

{{% /tab %}}
{{< /tabs >}}


Here are the defaults if you don't specify your own. I recommend using them as a starting point for your templates.

### Default unit file

```ini
[Unit]
Description={{ .JobDescription }}
{{ if .AfterNetworkOnline }}After=network-online.target
{{ end }}
[Service]
Type=notify
WorkingDirectory={{ .WorkingDirectory }}
ExecStart={{ .CommandLine }}
{{ if .Nice }}Nice={{ .Nice }}
{{ end -}}
{{ if .CPUSchedulingPolicy }}CPUSchedulingPolicy={{ .CPUSchedulingPolicy }}
{{ end -}}
{{ if .IOSchedulingClass }}IOSchedulingClass={{ .IOSchedulingClass }}
{{ end -}}
{{ if .IOSchedulingPriority }}IOSchedulingPriority={{ .IOSchedulingPriority }}
{{ end -}}
{{ range .Environment -}}
Environment="{{ . }}"
{{ end -}}
```

### Default timer file

```ini
[Unit]
Description={{ .TimerDescription }}

[Timer]
{{ range .OnCalendar -}}
OnCalendar={{ . }}
{{ end -}}
Unit={{ .SystemdProfile }}
Persistent=true

[Install]
WantedBy=timers.target
```

### Template variables

These are available for both the unit and timer templates:

* JobDescription   *string*
* TimerDescription *string*
* WorkingDirectory *string*
* CommandLine      *string*
* OnCalendar       *array of strings*
* SystemdProfile   *string*
* Nice             *integer*
* Environment      *array of strings*
