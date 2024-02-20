---
title: "Launchd"
date: 2022-05-16T20:13:32+01:00
weight: 110
---




`launchd` is the service manager on macOS. resticprofile can schedule a profile via a _user agent_ or a _daemon_ in launchd.

## User agent

A user agent is generated when you set `schedule-permission` to `user`.

It consists of a `plist` file in the folder `~/Library/LaunchAgents`:

A user agent **mostly** runs with the privileges of the user. But if you backup some specific files, like your contacts or your calendar for example, you will need to give more permissions to resticprofile **and** restic.

For this to happen, you need to start the agent or daemon from a console window first (resticprofile will ask if you want to do so)

If your profile is a backup profile called `remote`, the command to run manually is:

```shell
launchctl start local.resticprofile.remote.backup
```

Once you grant the permission, the background agents/daemon will be able to run normally.

There's some information in this thread: https://github.com/restic/restic/issues/2051

*TODO: I'm going to try to compile a comprehensive how-to guide from all the information from the thread. Stay tuned!*

### Special case of schedule-permission=user with sudo

Please note if you schedule a user agent while running resticprofile with sudo: the user agent will be registered to the root user, and not your initial user context. It means you can only see it (`status`) and remove it (`unschedule`) via sudo.

## Daemon

A launchd daemon is generated when you set `schedule-permission` to `system`. 

It consists of a `plist` file in the folder `/Library/LaunchDaemons`. You have to run resticprofile with sudo to `schedule`, check the  `status` and `unschedule` the profile.
