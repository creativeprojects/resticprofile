---
title: "Systemd"
weight: 105
# tags: ["v0.25.0", "v0.29.0"]
---



**systemd** is a common service manager in use by many Linux distributions.
resticprofile has the ability to create systemd timer and service files.
systemd can be used in place of cron to schedule backups.

User systemd units are created under the user's systemd profile (`~/.config/systemd/user`).

System units are created in `/etc/systemd/system`

## systemd calendars

resticprofile uses systemd
[OnCalendar](https://www.freedesktop.org/software/systemd/man/systemd.time.html#Calendar%20Events)
format to schedule events.

Testing systemd calendars can be done with the systemd-analyze application.
systemd-analyze will display when the next trigger will happen:

```shell
systemd-analyze calendar 'daily'

  Original form: daily
Normalized form: *-*-* 00:00:00
    Next elapse: Sat 2020-04-18 00:00:00 CDT
       (in UTC): Sat 2020-04-18 05:00:00 UTC
       From now: 10h left
```

## First time schedule

When you schedule a profile with the `schedule` command, under the hood resticprofile will
- create the unit file (of type `notify`)
- create the timer file
- run `systemctl daemon-reload` (only if `schedule-permission` is set to `system`)
- run `systemctl enable`
- run `systemctl start`

## Run after the network is up

Specifying the profile option `schedule-after-network-online: true` means that the scheduled services will wait
for a network connection before running.
This is done via an [After=network-online.target](https://systemd.io/NETWORK_ONLINE/) entry in the service.


## systemd drop-in files

It is possible to automatically populate `*.conf.d`
[drop-in files](https://www.freedesktop.org/software/systemd/man/latest/systemd-system.conf.html#main-conf)
for profiles, which allows easy overriding
of the generated services, without modifying the service templates. For example:

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

Generated with the following, see [systemd credentials docs](https://systemd.io/CREDENTIALS/)
for more details. This could allow, for example,
using a TPM-backed encrypted password, outside of the
resticprofile config itself

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

## Priority and CPU scheduling

resticprofile allows you to set the `nice` value, the CPU scheduling policy and IO nice values for the systemd service.
This is only working properly for resticprofile >= 0.29.0.

| systemd unit option  | resticprofile option |
|----------------------|----------------------|
| CPUSchedulingPolicy  | set to `idle` if schedule `priority` = `background` , otherwise default to standard policy |
| Nice                 | `nice` from `global` section |
| IOSchedulingClass    | `ionice-class` from `global` section |
| IOSchedulingPriority | `ionice-level` from `global` section |

{{% notice note %}}
When setting the `CPUSchedulingPolicy` to `idle` (by setting `priority` to `background`), the backup might never execute if all your CPU cores are always busy.
{{% /notice %}}

## How to change the default systemd unit and timer file using a template

By default, an opinionated systemd unit and timer are automatically generated by resticprofile.

Since version 0.16.0, you now can describe your own templates if you need to add things in it (typically like sending an email on failure).

The format used is a [go template](https://pkg.go.dev/text/template) and you need to specify your own unit and/or timer file in the global section of the configuration (it will apply to all your profiles):

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

Here are the defaults if you don't specify your own (which I recommend to use as a starting point for your own templates)

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
