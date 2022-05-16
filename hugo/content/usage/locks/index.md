---
title: "Locks"
date: 2022-05-16T20:26:09+01:00
weight: 20
---


# Locks

restic is already using a lock to avoid running some operations at the same time.

Since resticprofile can run several commands in a profile, it could be better to run the whole batch in a lock so nobody can interfere in the meantime.

For this to happen you can specify a lock file in each profile:

```yaml
src:
    lock: "/tmp/resticprofile-profile-src.lock"
    backup:
        check-before: true
        exclude:
        - /**/.git
        source:
        - ~/go
    retention:
        after-backup: true
        before-backup: false
        compact: false
        keep-within: 30d
        prune: true
```

For this profile, a lock will be set using the file `/tmp/resticprofile-profile-src.lock` for the duration of the profile: *check*, *backup* and *retention* (via the forget command)

**Please note restic locks and resticprofile locks are completely independent**

## Stale locks

In some cases, resticprofile as well as restic may leave a lock behind if the process died (or the machine rebooted).

For that matter, if you add the flag `force-inactive-lock` to your profile, resticprofile will detect and remove stale locks: 
* **resticprofile locks**: Check for the presence of a process with the PID indicated in the lockfile. If it can't find any, it will try to delete the lock and continue the operation (locking again, running profile and so on...)
* **restic locks**: Evaluate if a restic command failed on acquiring a lock. If the lock is older than `restic-stale-lock-age`, invoke `restic unlock` and retry the command that failed (can be disabled by setting `restic-stale-lock-age` to 0, default is 2h).

```yaml
global:
  restic-stale-lock-age: 2h

src:
    lock: "/tmp/resticprofile-profile-src.lock"
    force-inactive-lock: true
```

## Lock wait

By default, restic and resticprofile fail when a lock cannot be acquired as another process is currently holding it.

Depending on the use case (e.g. scheduled backups), it may be more appropriate to wait on another process to finish instead of failing immediately.

For that matter, if you add the commandline flag `--lock-wait` or configure schedules with `schedule-lock-wait`, resticprofile will wait on other backup processes:
* **resticprofile locks**: Retry acquiring the lockfile until it either succeeds (when the other resticprofile process released the lock) or fail as the lock-wait duration has passed without success.
* **restic locks**: Evaluate if a restic command failed on acquiring a lock. If the lock is not considered stale, retry the restic command every `restic-lock-retry-after` (default 1 minute) until it acquired the lock, or fail as the lock-wait duration has passed.

Note: The lock wait duration is cumulative. If various locks in one profile-run require lock wait, the total wait time may not exceed the duration that was specified. 

## restic lock management

resticprofile can retry restic commands that fail on acquiring a lock and can also ask restic to unlock stale locks. The behaviour is controlled by 2 settings inside the `global` section:

```yaml
global:
  # Retry a restic command that failed on acquiring a lock every minute 
  # (at least), for up to the time specified in "--lock-wait duration". 
  restic-lock-retry-after: 1m
  # Ask restic to unlock a stale lock when its age is more than 2 hours
  # and the option "force-inactive-lock" is enabled in the profile.
  restic-stale-lock-age: 2h
```

If restic lock management is not desired, it can be disabled by setting both values to 0.
