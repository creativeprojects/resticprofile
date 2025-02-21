---
title: "No root backup on Linux"
weight: 25
hidden: true # for this to work we might need https://pkg.go.dev/kernel.org/pub/linux/libs/security/libcap/cap
---

{{%notice info %}}
This section is mostly copied from the `restic` documentation:

[Backing up your system without running restic as root](https://restic.readthedocs.io/en/latest/080_examples.html#backing-up-your-system-without-running-restic-as-root)
{{% /notice %}}

## Backing up your system without running resticprofile as root

Creating a complete backup of a machine requires a privileged process that is able to read all files. On UNIX-like systems this is traditionally the root user. Processes running as root have superpower. They cannot only read all files but do also have the power to modify the system in any possible way.

With great power comes great responsibility. If a process running as root malfunctions, is exploited, or simply configured in a wrong way it can cause any possible damage to the system. This means you only want to run programs as root that you trust completely. And even if you trust a program, it is good and common practice to run it with the least possible privileges.

### Capabilities on Linux

Fortunately, Linux has functionality to divide root’s power into single separate capabilities. You can remove these from a process running as root to restrict it. And you can add capabilities to a process running as a normal user, which is what we are going to do.

### Full backup without root

To be able to completely backup a system, restic has to read all the files. Luckily Linux knows a capability that allows precisely this. We can assign this single capability to restic and then run it as an unprivileged user.

First we create a new user called restic that is going to create the backups:

```shell
useradd --system --create-home --shell /sbin/nologin restic
```

Then we download and install the **resticprofile** and **restic** binary into the user’s home directory (please refer to the respective installation sections). Let's save both binaries in the `~/bin` directory.

Before we assign any special capability to the binaries, we restrict their permissions so that only root and the newly created restic user can execute them. Otherwise another - *possibly untrusted* - user could misuse the privileged binaries to circumvent file access controls.

```shell
chown root:restic ~restic/bin/restic
chmod 750 ~restic/bin/restic
chown root:restic ~restic/bin/resticprofile
chmod 750 ~restic/bin/resticprofile
```

Finally we can use **setcap** to add an extended attribute to the binaries. On every execution the system will read the extended attribute, interpret it and assign capabilities accordingly.

```shell
setcap cap_dac_read_search=+ep ~restic/bin/restic
setcap cap_dac_read_search=pie ~restic/bin/resticprofile
```

{{%notice warning %}}
The capabilities of the **setcap** command only applies to this specific copy of the binaries. If you run `restic self-update` or in any other way replace or update the binaries, the capabilities you added above will not be in effect for the new binaries. To mitigate this, simply run the **setcap** commands again, to make sure that the new binaries have the same and intended capabilities.
{{% /notice %}}

From now on the user restic can run **resticprofile** to backup the whole system.

```shell
sudo -u restic /home/restic/bin/resticprofile backup
```
