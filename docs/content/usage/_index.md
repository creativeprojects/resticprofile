---
archetype: "chapter"
pre: "<b>3. </b>"
title: "Using resticprofile"
weight: 3
---

A command is either a restic command or a resticprofile own command.

## Generating random keys

Resticprofile includes a tool to generate cryptographically secure, base64-encoded random keys. Save the key to a file and use it as a strong key for Restic.

- On Linux, FreeBSD, Dragonfly, and Solaris, Reader uses `getrandom(2)`.
- On legacy Linux (< 3.17), it uses `/dev/urandom`.
- On macOS, and OpenBSD Reader, uses `arc4random_buf(3)`.
- On NetBSD, Reader uses the kern.arandom sysctl.
- On Windows, Reader uses the ProcessPrng API.

[Reference from the Go cryto library documentation](https://golang.org/pkg/crypto/rand/#pkg-variables)

```shell
resticprofile generate --random-key
```

Generates a 1024-byte random key (converted to 1368 Base64 characters) and displays it in the console.

To generate a key of a different size, specify the byte length in the command line.

```shell
resticprofile generate --random-key 2048
```


## Examples

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
