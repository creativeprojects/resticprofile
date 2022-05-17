---
title: "Memory"
date: 2022-05-16T20:20:55+01:00
weight: 15
---


## Minimum memory required

restic can be memory hungry. I'm running a few servers with no swap (I know: it is _bad_) and I managed to kill some of them during a backup.
For that matter I've introduced a parameter in the `global` section called `min-memory`. The **default value is 100MB**. You can disable it by using a value of `0`.

It compares against `(total - used)` which is probably the best way to know how much memory is available (that is including the memory used for disk buffers/cache).



