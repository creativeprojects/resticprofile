---
title: "Example: Windows"
weight: 31
---

{{% notice note %}}
If you create a task with `user` permission under Windows, you will need to enter your password to validate the task.

It's a requirement of the task scheduler. I'm inviting you to review the code to make sure I'm not emailing your password to myself. Seriously you shouldn't trust anyone.
{{% /notice %}}

Example of the `schedule` command under Windows (with git bash):

```shell
$ resticprofile -c examples/windows.yaml -n self schedule

Analyzing backup schedule 1/2
=================================
  Original form: Mon..Fri *:00,15,30,45
Normalized form: Mon..Fri *-*-* *:00,15,30,45:00
    Next elapse: Wed Jul 22 21:30:00 BST 2020
       (in UTC): Wed Jul 22 20:30:00 UTC 2020
       From now: 1m52s left

Analyzing backup schedule 2/2
=================================
  Original form: Sat,Sun 0,12:00
Normalized form: Sat,Sun *-*-* 00,12:00:00
    Next elapse: Sat Jul 25 00:00:00 BST 2020
       (in UTC): Fri Jul 24 23:00:00 UTC 2020
       From now: 50h31m52s left

Creating task for user Creative Projects
Task Scheduler requires your Windows password to validate the task: 

2020/07/22 21:28:15 scheduled job self/backup created

Analyzing retention schedule 1/1
=================================
  Original form: sun 3:30
Normalized form: Sun *-*-* 03:30:00
    Next elapse: Sun Jul 26 03:30:00 BST 2020
       (in UTC): Sun Jul 26 02:30:00 UTC 2020
       From now: 78h1m44s left

2020/07/22 21:28:22 scheduled job self/retention created
```

To see the status of the triggers, you can use the `status` command:

```shell
$ resticprofile -c examples/windows.yaml -n self status

Analyzing backup schedule 1/2
=================================
  Original form: Mon..Fri *:00,15,30,45
Normalized form: Mon..Fri *-*-* *:00,15,30,45:00
    Next elapse: Wed Jul 22 21:30:00 BST 2020
       (in UTC): Wed Jul 22 20:30:00 UTC 2020
       From now: 14s left

Analyzing backup schedule 2/2
=================================
  Original form: Sat,Sun 0,12:*
Normalized form: Sat,Sun *-*-* 00,12:*:00
    Next elapse: Sat Jul 25 00:00:00 BST 2020
       (in UTC): Fri Jul 24 23:00:00 UTC 2020
       From now: 50h29m46s left

           Task: \resticprofile backup\self backup
           User: Creative Projects
    Working Dir: D:\Source\resticprofile
           Exec: D:\Source\resticprofile\resticprofile.exe --no-ansi --config examples/windows.yaml --name self backup
        Enabled: true
          State: ready
    Missed runs: 0
  Last Run Time: 2020-07-22 21:30:00 +0000 UTC
    Last Result: 0
  Next Run Time: 2020-07-22 21:45:00 +0000 UTC

Analyzing retention schedule 1/1
=================================
  Original form: sun 3:30
Normalized form: Sun *-*-* 03:30:00
    Next elapse: Sun Jul 26 03:30:00 BST 2020
       (in UTC): Sun Jul 26 02:30:00 UTC 2020
       From now: 77h59m46s left

           Task: \resticprofile backup\self retention
           User: Creative Projects
    Working Dir: D:\Source\resticprofile
           Exec: D:\Source\resticprofile\resticprofile.exe --no-ansi --config examples/windows.yaml --name self forget
        Enabled: true
          State: ready
    Missed runs: 0
  Last Run Time: 1999-11-30 00:00:00 +0000 UTC
    Last Result: 267011
  Next Run Time: 2020-07-26 03:30:00 +0000 UTC

```

To remove the schedule, use the `unschedule` command:

```shell
$ resticprofile -c examples/windows.yaml -n self unschedule
2020/07/22 21:34:51 scheduled job self/backup removed
2020/07/22 21:34:51 scheduled job self/retention removed
```

