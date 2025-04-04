---
title: "Schedule Commands"
weight: 20
---


resticprofile accepts these commands:
- **schedule**
- **unschedule**
- **status**

These commands operate on the profile or group selected by `--name`, or on all profiles when `--all` is passed.

{{% notice style="warning" %}}
Before version `0.29.0`, the `--name` flag on a group selected all profiles in the group for scheduling, similar to running the schedule command on each profile individually.

Version `0.29.0` introduced group scheduling: The group schedule works at the group level and runs all profiles one by one when triggered.
{{% /notice %}}


Examples:
```shell
resticprofile --name profile schedule 
resticprofile --name group schedule 
resticprofile schedule --all 
```

Schedules are always independent, regardless of whether they are created with `--all` or from a single profile.

### schedule command

Install all schedules defined in the selected profile(s).

Note: On systemd, you need to `start` the timer once to enable it. Otherwise, it will only be enabled on the next reboot. If you don't want to start (and enable) it now, pass the `--no-start` flag to the command.

If you use the `--all` flag to schedule all profiles at once, use either `user` mode or `system` mode. Combining both will not schedule tasks properly:
- If the user is not privileged, only `user` tasks will be scheduled.
- If the user is privileged, all schedules will be `system` schedules.

{{% notice style=tip %}}
Before version `v0.30.0`, resticprofile did not keep a state of the schedule and unschedule commands. If you needed to make many changes to your profiles (e.g., moving, renaming, deleting), it was recommended to unschedule everything using the `--all` flag before making changes. This is no longer necessary since version `v0.30.0`.
{{% /notice %}}

### unschedule command

Remove all schedules defined on the selected profile, or all profiles using the `--all` flag.

Before version `v0.30.0`, the `--all` flag didn't remove schedules on deleted or renamed profiles.

> [!NOTE]
> The behavior of the `unschedule` command changed in version `v0.30.0`:
>
> It now deletes any schedule associated with the profile name, or any profile in the configuration file with `--all` (including deleted profiles).

### status command

Print the status of all installed schedules for the selected profile(s).

The `status` command output depends on the OS. Refer to the [examples]({{% relref "/schedules/examples" %}}) for expected output.

### run-schedule command


This command is used by the scheduler to tell resticprofile to execute within a schedule. It sets the proper log output (`schedule-log`) and other schedule-specific flags.

If you're scheduling resticprofile manually, use this command. It executes the profile with all `schedule-*` parameters defined in the profile.

This command takes one argument: the command name followed by the profile name, separated by an `@` sign.

```shell
resticprofile run-schedule backup@profile
```

{{% notice info %}}
For the `run-schedule` command, you cannot specify the profile name using the `--name` flag.
{{% /notice %}}