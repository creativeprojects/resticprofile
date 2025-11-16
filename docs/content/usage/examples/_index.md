---
title: Examples
weight: 20
---

Here are a few examples how to run resticprofile (using the main example configuration file)

### See all snapshots of your `default` profile:

```shell
resticprofile
```

### See all available profiles in your configuration file (and the restic commands where some flags are defined):

```shell
resticprofile profiles

Profiles available (name, sections, description):
  root:           (backup, copy, forget, retention)
  self:           (backup, check, copy, forget, retention)
  src:            (backup, copy, retention, snapshots)

Groups available (name, profiles, description):
  full-backup:  [root, src]

```

### Backup root & src profiles (using _full-backup_ group shown earlier)

```shell
resticprofile --name "full-backup" backup
```
or using the alternative syntax:

```shell
resticprofile full-backup.backup
```

Assuming the _stdin_ profile from the configuration file shown before, the command to send a mysqldump to the backup is as simple as:

```shell
mysqldump --all-databases --order-by-primary | resticprofile --name stdin backup
```
or using the alternative syntax:

```shell
mysqldump --all-databases --order-by-primary | resticprofile stdin.backup
```

### Mount the default profile (_default_) in /mnt/restic:

```shell
resticprofile mount /mnt/restic
```

### Display quick help

```shell
resticprofile --help
```
