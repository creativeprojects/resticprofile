---
title: "Reference"
date: 2022-05-16T20:07:43+01:00
weight: 50
---


## Configuration file reference

`[global]`

`global` is a fixed name

None of these flags are passed on the restic command line

* **ionice**: true / false
* **ionice-class**: integer
* **ionice-level**: integer
* **nice**: true / false OR integer
* **priority**: string = `Idle`, `Background`, `Low`, `Normal`, `High`, `Highest`
* **default-command**: string
* **initialize**: true / false
* **restic-binary**: string
* **restic-lock-retry-after**: duration
* **restic-stale-lock-age**: duration
* **min-memory**: integer (MB)
* **scheduler**: string (`crond` is the only non-default value)
* **systemd-unit-template**: string (file containing a go template to generate systemd unit file)
* **systemd-timer-template**: string (file containing a go template to generate systemd timer file)

`[profile]`

`profile` is the name of your profile

Flags used by resticprofile only

* ****inherit****: string
* **description**: string
* **initialize**: true / false
* **lock**: string: specify a local lockfile
* **force-inactive-lock**: true / false
* **run-before**: string OR list of strings
* **run-after**: string OR list of strings
* **run-after-fail**: string OR list of strings
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

