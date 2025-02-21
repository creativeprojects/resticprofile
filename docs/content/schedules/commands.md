---
title: "Schedule Commands"
weight: 20
tags: ["v0.25.0", "v0.29.0", "v0.30.0"]
---


resticprofile accepts these internal commands:
- **schedule**
- **unschedule**
- **status**

These resticprofile commands either operate on the profile selected by `--name`, on the profiles selected by a group (before `v0.29.0`), on groups (from v0.29.0), or on all profiles when the flag `--all` is passed on the command line.

{{% notice style="warning" %}}
Before version `0.29.0`, the `--name` flag on a group was used to select all profiles in the group for scheduling them. It was similar to running the schedule commands on each profile individually.

Version `0.29.0` introduced group scheduling: The group schedule works at the group level (and will run all profiles one by one when the group schedule is triggered).
{{% /notice %}}

Examples:
```shell
resticprofile --name profile schedule 
resticprofile --name group schedule 
resticprofile schedule --all 
```

Please note, schedules are always independent of each other no matter whether they have been created with `--all`, by group or from a single profile.

### schedule command

Install all the schedules defined on the selected profile or profiles.

Please note on systemd, we need to `start` the timer once to enable it. Otherwise it will only be enabled on the next reboot. If you **don't want** to start (and enable) it now, pass the `--no-start` flag to the command line.

Also if you use the `--all` flag to schedule all your profiles at once, make sure you use only the `user` mode or `system` mode. A combination of both would not schedule the tasks properly:
- if the user is not privileged, only the `user` tasks will be scheduled
- if the user **is** privileged, **all schedules will end up as a `system` schedule**
{{% notice style=tip %}}
Before version `v0.30.0` resticprofile **did not keep a state** of the schedule and unschedule commands. If you need to make many changes in your profiles (like moving, renaming, deleting etc.) **it was recommended to unschedule everything using the `--all` flag before changing your profiles.**. This is no longer the case since version `v0.30.0`.
{{% /notice %}}

### unschedule command

Remove all the schedules defined on the selected profile, or all profiles via the `--all` flag.

Before version `v0.30.0`, using the `--all` flag wasn't removing schedules on deleted (or renamed) profiles.

> [!NOTE]
> The behavior of the `unschedule` command has changed in version `v0.30.0`:
>
> it now deletes **any schedule associated with the profile name, or any profile of the configuration file with `--all`** (even profiles deleted from the configuration file)

### status command

Print the status on all the installed schedules of the selected profile or profiles. 

The display of the `status` command will be OS dependant. Please refer to the [examples]({{% relref "/schedules/examples" %}}) on which output you can expect from it.

### run-schedule command

This is the command that is used internally by the scheduler to tell resticprofile to execute in the context of a schedule. It means it will set the proper log output (`schedule-log`) and all other flags specific to the schedule.

If you're scheduling resticprofile manually you can use this command. It will execute the profile using all the `schedule-*` parameters defined in the profile.

This command is only taking one argument: name of the command to execute, followed by the profile name, both separated by a `@` sign.

```shell
resticprofile run-schedule backup@profile
```

{{% notice info %}}
For the `run-schedule` command, you cannot specify the profile name using the `--name` flag.
{{% /notice %}}
