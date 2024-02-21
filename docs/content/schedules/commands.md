---
title: "Schedule Commands"
weight: 20
tags: ["v0.25.0"]
---


resticprofile accepts these internal commands:
- **schedule**
- **unschedule**
- **status**

These resticprofile commands either operate on the profile selected by `--name`, on the profiles selected by a group, or on all profiles when the flag `--all` is passed on the command line.

Examples:
```shell
resticprofile --name profile schedule 
resticprofile --name group schedule 
resticprofile schedule --all 
```

Please note, schedules are always independent of each other no matter whether they have been created with `--all`, by group or from a single profile.

### schedule command

Install all the schedules defined on the selected profile or profiles.

Please note on systemd, we need to `start` the timer once to enable it. Otherwise it will only be enabled on the next reboot. If you **dont' want** to start (and enable) it now, pass the `--no-start` flag to the command line.

Also if you use the `--all` flag to schedule all your profiles at once, make sure you use only the `user` mode or `system` mode. A combination of both would not schedule the tasks properly:
- if the user is not privileged, only the `user` tasks will be scheduled
- if the user **is** privileged, **all schedule will end-up as a `system` schedule**

### unschedule command

Remove all the schedules defined on the selected profile or profiles.

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
You cannot specify the profile name using the `--name` flag.
{{% /notice %}}
