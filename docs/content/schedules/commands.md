---
title: "Schedule Commands"
weight: 20
---

resticprofile supports the following commands:  
- **schedule**  
- **unschedule**  
- **status**  

These commands apply to the profile or group specified by `--name`, or to all profiles when `--all` is used.

{{% notice style="warning" %}}  
In versions prior to `0.29.0`, the `--name` flag for a group scheduled all profiles in the group individually, as if the schedule command was run on each profile.  

Starting with version `0.29.0`, group scheduling was introduced. This feature schedules at the group level, executing all profiles sequentially when triggered.  
{{% /notice %}}  

Examples:
```shell
resticprofile --name profile schedule 
resticprofile --name group schedule 
resticprofile schedule --all 
```

Schedules are always independent, whether created with `--all` or a single profile.

### schedule command

Install all schedules defined in the selected profile(s).

**Note:** On systemd, you must `start` the timer once to enable it. Otherwise, it will only activate after the next reboot. To skip starting (and enabling) it now, use the `--no-start` flag.

When using the `--all` flag to schedule all profiles at once, choose either `user` mode or `system` mode. Mixing both will cause scheduling issues:
- Non-privileged users can only schedule `user` tasks.
- Privileged users will schedule all tasks as `system` tasks.

Use the `--reload` flag to trigger a `systemctl daemon-reload` after setting up the schedule files. This is helpful if systemd fails to detect manually added dependencies in the service file. The flag is available starting from version 0.32.0.

{{% notice style=tip %}}
Before version `v0.30.0`, resticprofile did not track the state of schedule and unschedule commands. If you needed to make significant changes to profiles (e.g., moving, renaming, deleting), it was recommended to unschedule everything using the `--all` flag first. This is no longer required as of version `v0.30.0`.
{{% /notice %}}

### unschedule command

Remove all schedules from the selected profile or all profiles using the `--all` flag.

Before `v0.30.0`, the `--all` flag did not remove schedules for deleted or renamed profiles.

> [!NOTE]  
> Starting with `v0.30.0`, the behavior of the `unschedule` command changed:  
> - Without the `--all` flag, it deletes schedules associated with the profile name.  
> - With the `--all` flag, it removes all profiles from the configuration file, including deleted and renamed ones.  

### status command

Print the status of all installed schedules for the selected profiles.

The `status` command output varies by OS. See the [examples]({{% relref "/schedules/examples" %}}) for details.

### run-schedule command

This command allows the scheduler to instruct resticprofile to run according to a schedule. It configures the appropriate log output (`schedule-log`) and other schedule-specific flags.

If you're manually scheduling resticprofile, use this command. It runs the profile with all `schedule-*` parameters defined in the profile.

The command requires one argument: the command name followed by the profile name, separated by an `@` symbol.

```shell
resticprofile run-schedule backup@profile
```

{{% notice info %}}
The `--name` flag cannot be used to specify the profile name with the `run-schedule` command.
{{% /notice %}}
