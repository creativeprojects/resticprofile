---
title: "Warnings"
date: 2022-05-16T20:24:23+01:00
weight: 30
---

# Warnings from restic

Until version 0.13.0, resticprofile was always considering a restic warning as an error. This will remain the **default**.
But the version 0.13.0 introduced a parameter to avoid this behaviour and consider that the backup was successful instead.

A restic warning occurs when it cannot read some files, but a snapshot was successfully created.

## no-error-on-warning

```yaml
profile:
    inherit: default
    backup:
        no-error-on-warning: true
```
