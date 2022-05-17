---
title: "Upgrade"
date: 2022-04-23T23:57:06+01:00
weight: 20
---

Once installed, you can easily upgrade resticprofile to the latest release using this command:

```
$ resticprofile self-update
```

_Please note on versions before 0.10.0, there was an issue with self-updating from linux with ARM processors (like a raspberry pi). This was fixed in version 0.10.0_

resticprofile will check for a new version from GitHub releases and asks you if you want to update to the new version. If you add the flag `-q` or `--quiet` to the command line, it will update automatically without asking.

```
$ resticprofile --quiet self-update
```

and since version 0.11.0:

```
$ resticprofile self-update --quiet
```
