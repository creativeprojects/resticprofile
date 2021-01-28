# Feeding the resticprofile status file into Zabbix

I have created a Zabbix template which is reading the status file from resticprofile, and is sending an alert if a profile didn't finish in the last 30 hours (a bit more than a day in case a profile takes a bit longer) or if it failed. The assumption is that each profile target is running once a day.

There's one MACRO needed:
`{$BACKUP_STATUS_FILE}` which contain the full path of the status file

**If your profiles are running more than once a day, you will need to edit the template according to your needs.**

## Running profiles manually

I recommend making a different profile for scheduling and for running commands manually.
Let's consider this example:

```toml

[profile1]
repository = "rest:http://user:password@server:8000/backup"
password-file = "key"
status-file = "/home/backup/status.json"

    [profile1.backup]
    source = [ "/" ]
    schedule = "01:47"
    schedule-permission = "system"

```

The status file will contain an entry for `profile1.backup`.

Now if you need to `check` this repository via

```
$ resticprofile -n profile1 check
```

When the `check` is finished, resticprofile will generate an entry for `profile1.check` in the status file.
Meaning that tomorrow you will get an alert that the `profile1.check` didn't run.

An easy fix is to create a profile to run commands manually:

```toml

[manual]
repository = "rest:http://user:password@server:8000/backup"
password-file = "key"

[scheduled]
inherit = "manual"
status-file = "/home/backup/status.json"

    [scheduled.backup]
    source = [ "/" ]
    schedule = "01:47"
    schedule-permission = "system"

```

With this configuration, there won't be a new entry in the status file when you `check` the repository:

```
$ resticprofile -n manual check
```
