+++
archetype = "chapter"
pre = "<b>1. </b>"
title = "Installation"
weight = 1
+++

Here are the multiple ways you can install resticprofile:
{{% children  %}}

## Upgrading
After installation, you can run `resticprofile self-update` to update resticprofile to the latest release. The `-q` or `--quiet` flag updates the application without prompting.

## Note about Memory

restic can be memory hungry. I'm running a few servers with no swap and I managed to kill some of them during a backup.

For that matter I've introduced a parameter in the `global` section called `min-memory`. The **default value is 100MB**. You can disable it by using a value of `0`.

It compares against `(total - used)` which is probably the best way to know how much memory is available (that is including the memory used for disk buffers/cache).
