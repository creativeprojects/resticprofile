---
title: "Reference"
date: 2022-05-16T20:07:43+01:00
weight: 50
---


## Configuration file reference

### Section `global`

`global` is a fixed section name, at the root of the configuration file

{{% notice info %}}
None of these flags are directly passed on to the restic command line
{{% /notice %}}


| Name | Type | Default | Notes |
|:-----|:-----|:--------|:------|
| **ionice** | true / false | false |
| **ionice-class** | integer | 0 |
| **ionice-level** | integer | 0 |
| **nice** | true / false OR integer | 0 |
| **priority** | string | `Normal` | values are `Idle`, `Background`, `Low`, `Normal`, `High`, `Highest` |
| **default-command** | string | `snapshots` |
| **initialize** | true / false | false | auto-initialize a repository |
| **restic-binary** | string | | full path of the restic program |
| **restic-lock-retry-after** | duration [^duration] | 1 minute | see [locks]({{< ref "/usage/locks" >}}) |
| **restic-stale-lock-age** | duration [^duration] | 2 hours | see [locks]({{< ref "/usage/locks" >}}) |
| **min-memory** | integer (MB) | 100MB | see [memory]({{< ref "/usage/memory" >}}) |
| **shell** | string | OS specific | shell binary to run commands |
| **scheduler** | string | | `crond` is the only non-default value |
| **systemd-unit-template** | string | | file containing a go template to generate systemd unit file - see [systemd templates]({{< ref "/schedules/systemd" >}}) |
| **systemd-timer-template** | string | | file containing a go template to generate systemd timer file - see [systemd templates]({{< ref "/schedules/systemd" >}}) |
| **send-timeout** | duration [^duration] | 30 seconds | timeout when sending messages to a webhook - see [HTTP Hooks]({{< ref "http_hooks" >}}) |
| **ca-certificates** | string, or list of strings | | certificates (file in PEM format) to authenticate HTTP servers - see [HTTP Hooks]({{< ref "http_hooks" >}}) |
| **prevent-sleep** | true / false | false | prevent the system from sleeping - see [Preventing system sleep]({{< ref "sleep" >}}) |
| **group-next-on-error** | true / false | false | if set to `true` it allows the next profile(s) to run after a failure |

### Profile sections

The name of this section is the name of your profile.

{{% notice note %}}
You cannot use the names `global` or `groups`.
{{% /notice %}}


Flags used by resticprofile only

* ****inherit****: string
* **description**: string
* **initialize**: true / false
* **lock**: string: specify a local lockfile
* **force-inactive-lock**: true / false
* **run-before**: string OR list of strings
* **run-after**: string OR list of strings
* **run-after-fail**: string OR list of strings
* **run-finally**: string OR list of strings
* **status-file**: string
* **prometheus-save-to-file**: string
* **prometheus-push**: string

Flags passed to the restic command line

* **cacert**: string
* **cache-dir**: string
* **cleanup-cache**: true / false
* **json**: true / false
* **key-hint**: string
* **limit-download**: integer
* **limit-upload**: integer
* **no-cache**: true / false
* **no-lock**: true / false
* **option**: string OR list of strings
* **password-command**: string
* **password-file**: string
* **quiet**: true / false
* **repository**: string **(will be passed as 'repo' to the command line)**
* **repository-file**: string
* **tls-client-cert**: string
* **verbose**: true / false OR integer

`[[profile.stream-error]]`
* **pattern**: regex (pattern matching stderr of `restic`) 
* **run**: string (command to run when stderr line is matched)
* **max-runs**: number
* **min-matches**: number

`[profile.backup]`

Flags used by resticprofile only

* **run-before**: string OR list of strings
* **run-after**: string OR list of strings
* **run-after-fail**: string OR list of strings
* **run-finally**: string OR list of strings
* **check-before**: true / false
* **check-after**: true / false
* **schedule**: string OR list of strings
* **schedule-permission**: string (`user` or `system`)
* **schedule-lock-mode**: string (`default`, `fail` or `ignore`) 
* **schedule-lock-wait**: duration 
* **schedule-log**: string
* **stdin-command**: string OR list of strings
* **extended-status**: true / false
* **no-error-on-warning**: true / false

Flags passed to the restic command line

* **exclude**: string OR list of strings
* **exclude-caches**: true / false
* **exclude-file**: string OR list of strings
* **exclude-if-present**: string OR list of strings
* **files-from**: string OR list of strings
* **force**: true / false
* **host**: true / false OR string
* **iexclude**: string OR list of strings
* **ignore-inode**: true / false
* **one-file-system**: true / false
* **parent**: string
* **stdin**: true / false
* **stdin-filename**: string
* **tag**: string OR list of strings
* **time**: string
* **with-atime**: true / false
* **source**: string OR list of strings

`[profile.retention]`

Flags used by resticprofile only

* **before-backup**: true / false
* **after-backup**: true / false
* **schedule**: string OR list of strings
* **schedule-permission**: string (`user` or `system`)
* **schedule-lock-mode**: string (`default`, `fail` or `ignore`)
* **schedule-lock-wait**: duration
* **schedule-log**: string

Flags passed to the restic command line

* **keep-last**: integer
* **keep-hourly**: integer
* **keep-daily**: integer
* **keep-weekly**: integer
* **keep-monthly**: integer
* **keep-yearly**: integer
* **keep-within**: string
* **keep-tag**: string OR list of strings
* **host**: true / false OR string
* **tag**: true / false, string OR list of strings
* **path**: true / false, string OR list of strings
* **compact**: true / false
* **group-by**: string
* **dry-run**: true / false
* **prune**: true / false

`[profile.snapshots]`

Flags passed to the restic command line

* **compact**: true / false
* **group-by**: string
* **host**: true / false OR string
* **last**: true / false
* **path**: true / false, string OR list of strings
* **tag**: true / false, string OR list of strings

`[profile.forget]`

Flags used by resticprofile only

* **schedule**: string OR list of strings
* **schedule-permission**: string (`user` or `system`)
* **schedule-lock-mode**: string (`default`, `fail` or `ignore`)
* **schedule-lock-wait**: duration
* **schedule-log**: string

Flags passed to the restic command line

* **keep-last**: integer
* **keep-hourly**: integer
* **keep-daily**: integer
* **keep-weekly**: integer
* **keep-monthly**: integer
* **keep-yearly**: integer
* **keep-within**: string
* **keep-tag**: string OR list of strings
* **host**: true / false OR string
* **tag**: true / false, string OR list of strings
* **path**: true / false, string OR list of strings
* **compact**: true / false
* **group-by**: string
* **dry-run**: true / false
* **prune**: true / false

`[profile.check]`

Flags used by resticprofile only

* **schedule**: string OR list of strings
* **schedule-permission**: string (`user` or `system`)
* **schedule-lock-mode**: string (`default`, `fail` or `ignore`)
* **schedule-lock-wait**: duration
* **schedule-log**: string

Flags passed to the restic command line

* **check-unused**: true / false
* **read-data**: true / false
* **read-data-subset**: string
* **with-cache**: true / false

`[profile.prune]`

Flags used by resticprofile only

* **schedule**: string OR list of strings
* **schedule-permission**: string (`user` or `system`)
* **schedule-lock-mode**: string (`default`, `fail` or `ignore`)
* **schedule-lock-wait**: duration
* **schedule-log**: string

`[profile.mount]`

Flags passed to the restic command line

* **allow-other**: true / false
* **allow-root**: true / false
* **host**: true / false OR string
* **no-default-permissions**: true / false
* **owner-root**: true / false
* **path**: true / false, string OR list of strings
* **snapshot-template**: string
* **tag**: true / false, string OR list of strings

`[profile.copy]`

Flags used by resticprofile only

* **initialize**: true / false
* **run-before**: string OR list of strings
* **run-after**: string OR list of strings
* **run-after-fail**: string OR list of strings
* **run-finally**: string OR list of strings
* **schedule**: string OR list of strings
* **schedule-permission**: string (`user` or `system`)
* **schedule-lock-mode**: string (`default`, `fail` or `ignore`)
* **schedule-lock-wait**: duration
* **schedule-log**: string

Flags passed to the restic command line

* **key-hint**: string **(will be passed as 'key-hint2')**
* **password-command**: command **(will be passed as 'password-command2')**
* **password-file**: string **(will be passed as 'password-file2')**
* **host**: true / false OR string
* **path**: true / false, string OR list of strings
* **repository**: repository **(will be passed as 'repo2')**
* **repository-file**: string **(will be passed as 'repository-file2')**
* **tag**: true / false, string OR list of strings

`[profile.dump]`

Flags passed to the restic command line

* **host**: true / false OR string
* **path**: true / false, string OR list of strings
* **tag**: true / false, string OR list of strings

`[profile.find]`

Flags passed to the restic command line

* **host**: true / false OR string
* **path**: true / false, string OR list of strings
* **tag**: true / false, string OR list of strings

`[profile.ls]`

Flags passed to the restic command line

* **host**: true / false OR string
* **path**: true / false, string OR list of strings
* **tag**: true / false, string OR list of strings

`[profile.restore]`

Flags passed to the restic command line

* **host**: true / false OR string
* **path**: true / false, string OR list of strings
* **tag**: true / false, string OR list of strings

`[profile.stats]`

Flags passed to the restic command line

* **host**: true / false OR string
* **path**: true / false, string OR list of strings
* **tag**: true / false, string OR list of strings

`[profile.tag]`

Flags passed to the restic command line

* **host**: true / false OR string
* **path**: true / false, string OR list of strings
* **tag**: true / false, string OR list of strings

[^duration]: A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". 
