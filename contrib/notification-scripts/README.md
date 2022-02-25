# Email with failure details - "resticprofile-send-error.sh" 

In `profiles.yaml` you set:

```yaml
default: 
  ...
  run-after-fail: 
    - 'resticprofile-send-error.sh -s name@domain.tl'
```

Usage:

```
resticprofile-send-error.sh [options] user1@domain user2@domain ...
Options:
   -s         Only send mail when operating on schedule (RESTICPROFILE_ON_SCHEDULE=1)
   -o name,.. Only send mail when PROFILE_NAME is in the list of specified names
   -c command Set the profile command (instead of PROFILE_COMMAND)
   -n name    Set the profile name (instead of PROFILE_NAME)
   -p         Print mail to stdout instead of sending it
   -f         Send mail even when no profile name is specified
```

## Quick installation

```sh
curl -ssL https://github.com/creativeprojects/resticprofile/raw/master/contrib/notification-scripts/resticprofile-send-error.sh \
     > /usr/local/bin/resticprofile-send-error.sh \
  && chmod +x /usr/local/bin/resticprofile-send-error.sh
```

Requirement: sendmail must be working

## Sample mail

In this example, the failure is caused by a custom pre-script complaining about a non-existing path:

```
Date: Fri, 23 Apr 2021 23:25:03 +0200
To: admins@domain.tl
From: "resticprofile hyper1.domain.tl" <root@hyper1.domain.tl>
Subject: restic failed: "backup" in "vms"

run-before backup on profile 'vms': exit status 1

---- 
COMMANDLINE:

"vm-snapshot create-all-to-folder /storage/vms /storage/vms/updates live-backup"

----
STDERR:

Panic: Expected [source] and [target] to be folders and [prefix] to be specified

---- 
DETAILS:

● resticprofile-backup@profile-vms.service - resticprofile backup for profile vms in /etc/resticprofile/profiles.yaml
  Loaded: loaded (/etc/systemd/system/resticprofile-backup@profile-vms.service; static; vendor preset: enabled)
  Active: active (running) since Fri 2021-04-23 23:25:02 CEST; 566ms ago
Main PID: 27259 (resticprofile)
   Tasks: 13 (limit: 4915)
  Memory: 12.1M
  CGroup: /system.slice/system-resticprofile\x2dbackup.slice/resticprofile-backup@profile-vms.service
          ├─27259 /usr/local/bin/resticprofile --no-prio --no-ansi --config /etc/resticprofile/profiles.yaml --name vms --lock-wait 45m0s backup
          ├─27294 /usr/bin/sh -c resticprofile-send-error.sh admins@domain.tl
          ├─27295 bash /usr/local/bin/resticprofile-send-error.sh admins@domain.tl
          ├─27296 bash /usr/local/bin/resticprofile-send-error.sh admins@domain.tl
          └─27299 systemctl status --full resticprofile-backup@profile-vms

Apr 23 23:25:02 hyper1 systemd[1]: Starting resticprofile backup for profile vms in /etc/resticprofile/profiles.yaml...
Apr 23 23:25:02 hyper1 systemd[1]: Started resticprofile backup for profile vms in /etc/resticprofile/profiles.yaml.
Apr 23 23:25:02 hyper1 resticprofile[27259]: 2021/04/23 23:25:02 profile 'vms': initializing repository (if not existing)
Apr 23 23:25:03 hyper1 resticprofile[27259]: Panic: Expected [source] and [target] to be folders and [prefix] to be specified
Apr 23 23:25:03 hyper1 resticprofile[27259]: Usage /usr/bin/vm-snapshot (exists|create|create-to-file|delete|delete-from-file|create-all-to-folder|delete-all-from-folder)

---- 
CONFIG:

2021/04/23 23:25:03 using configuration file: /etc/resticprofile/profiles.yaml
global:
   default-command:          snapshots
   restic-binary:            restic
   restic-lock-retry-after:  1m0s
   restic-stale-lock-age:    2h0m0s
   min-memory:               100

vms:
   repository:           rest:https://u:p@backup.domain.tl:8443/shared-repo/
   password-file:        /etc/resticprofile/shared-repo.key
   initialize:           true
   lock:                 /tmp/resticprofile.shared-repo.lock
   force-inactive-lock:  true
   run-after-fail:       resticprofile-send-error.sh admins@domain.tl

   env:
       tmpdir:  /tmp

   backup:
       schedule:             *-*-* 23:25
       schedule-permission:  system
       schedule-lock-wait:   45m0s
       run-before:           vmctl dumpxml-all
                             vm-snapshot create-all-to-folder /storage/vms /storage/vms/updates live-backup
       run-after:            vm-snapshot delete-all-from-folder /storage/vms/updates live-backup
       source:               /storage/vms
       exclude:              \*\*/updates/\*\*
       tag:                  vms
       one-file-system:      true
       exclude-caches:       true

   retention:
       after-backup:  true
       keep-daily:    14
       keep-weekly:   4
```

See details in [#20](https://github.com/creativeprojects/resticprofile/issues/20)
